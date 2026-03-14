package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/AmithSAI007/prj-apex-transcode-submitter/api"
	"github.com/AmithSAI007/prj-apex-transcode-submitter/api/handler"
	"github.com/AmithSAI007/prj-apex-transcode-submitter/api/middleware"
	_ "github.com/AmithSAI007/prj-apex-transcode-submitter/docs"
	"github.com/AmithSAI007/prj-apex-transcode-submitter/internal/config"
	"github.com/AmithSAI007/prj-apex-transcode-submitter/internal/platform"
	"github.com/gin-gonic/gin"

	// "go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"

	// "go.opentelemetry.io/otel"
	"go.uber.org/zap"
)

// @title			Transcode Submitter API
// @version		1.0
// @description	API for submitting transcode tasks
// @host			localhost:8080
// @BasePath		/api/v1
func main() {

	// Load application configuration from environment variables and .env files.
	cfg, err := config.LoadConfig(".")
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize the structured logger (JSON in production, console in development).
	logger, err := config.NewLogger()
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer func() {
		_ = logger.Sync()
	}()

	// Create a root context for the application lifecycle.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Set up the OpenTelemetry tracing pipeline (OTLP exporter, sampler, propagators).
	otelShutdown, err := platform.InitTracer(cfg, ctx)
	if err != nil {
		logger.Fatal("Failed to initialize tracer", zap.Error(err))
	}
	defer func() {
		err = errors.Join(err, otelShutdown(ctx))
	}()

	eventHandler := handler.NewEventHandler(logger)

	router := gin.New()
	router.MaxMultipartMemory = 32 << 20        // 32 MiB
	router.Use(gin.Recovery())                  // Recover from panics and return 500.
	router.Use(middleware.RequestContext())     // Inject trace/request IDs.
	router.Use(middleware.ErrorHandler(logger)) // Log unhandled errors.

	handlers := &api.HandlerRegistry{
		EventHandler: eventHandler,
	}

	api.SetupRoutes(router, handlers)

	svr := &http.Server{
		Addr:              cfg.HttpPort,
		Handler:           router,
		ReadHeaderTimeout: 10 * time.Second,
	}

	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		logger.Info("Starting server", zap.String("port", cfg.HttpPort))
		if err := svr.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatal("Server failed", zap.Error(err))
		}
	}()

	<-ctx.Done()
	logger.Info("Shutting down server...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := svr.Shutdown(shutdownCtx); err != nil {
		logger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Server exiting")

}
