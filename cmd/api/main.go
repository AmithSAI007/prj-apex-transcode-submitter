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
	"github.com/AmithSAI007/prj-apex-transcode-submitter/internal/repository"
	"github.com/AmithSAI007/prj-apex-transcode-submitter/internal/service"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.uber.org/zap"
)

// @title						Transcode Submitter API
// @version					1.0
// @description				Internal API consumed by Google Cloud Tasks to submit video transcoding jobs.
// @description				This service validates incoming task payloads, verifies raw media files in GCS,
// @description				checks idempotency via Firestore document state, and submits transcoding jobs
// @description				to the Google Cloud Transcoder API.
//
// @contact.name				Apex Platform Engineering
// @contact.email				platform-eng@apex.dev
//
// @license.name				Apache 2.0
// @license.url				https://www.apache.org/licenses/LICENSE-2.0.html
//
// @host						localhost:8080
// @BasePath					/api/v1
// @schemes					https http
//
// @x-google-backend			{"address": "https://transcode-submitter-HASH-uc.a.run.app"}
//
// @externalDocs.description	Apex Platform Architecture
// @externalDocs.url			https://docs.apex.dev/architecture/transcode-submitter
func main() {

	// Load application configuration from environment variables and .env files.
	cfg, err := config.LoadConfig(".")
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize the structured logger (JSON in production, console in development).
	logger, err := config.NewLogger(cfg.AppEnv)
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

	// ---- Platform clients ----

	gcsClient, err := platform.NewGCSClient(ctx)
	if err != nil {
		logger.Fatal("Failed to create GCS client", zap.Error(err))
	}
	defer func() {
		if cerr := gcsClient.Close(); cerr != nil {
			logger.Error("Failed to close GCS client", zap.Error(cerr))
		}
	}()

	fsClient, err := platform.NewClient(ctx, cfg.GCPProjectID, cfg.FirestoreDatabaseID)
	if err != nil {
		logger.Fatal("Failed to create Firestore client", zap.Error(err))
	}
	defer func() {
		if cerr := fsClient.Close(); cerr != nil {
			logger.Error("Failed to close Firestore client", zap.Error(cerr))
		}
	}()

	tcClient, err := platform.NewTranscoderClient(ctx, logger)
	if err != nil {
		logger.Fatal("Failed to create Transcoder client", zap.Error(err))
	}
	defer func() {
		if cerr := tcClient.Close(); cerr != nil {
			logger.Error("Failed to close Transcoder client", zap.Error(cerr))
		}
	}()

	// ---- Repository layer ----

	gcsRepo := repository.NewGCSRepositoryService(logger, gcsClient.Client())
	firestoreRepo := repository.NewFirestoreRepo(logger, fsClient.Client(), cfg.FirestoreCollection)
	transcoderRepo := repository.NewTranscoderService(logger, tcClient.Client(), cfg.GCPProjectID, cfg.ProjectRegion, cfg.TranscoderTemplateID)

	// ---- Service layer ----

	submitterSvc := service.NewSubmitterService(logger, *cfg, gcsRepo, firestoreRepo, transcoderRepo)

	// ---- Handler & router ----

	eventHandler := handler.NewEventHandler(logger, submitterSvc, *cfg)

	router := gin.New()
	router.MaxMultipartMemory = 32 << 20 // 32 MiB

	// Middleware stack (order matters):
	// 1. otelgin: creates root HTTP span from incoming W3C traceparent header.
	// 2. Recovery: catches panics and returns 500 before they crash the process.
	// 3. RequestContext: injects app-level trace/request IDs into context.
	// 4. ErrorHandler: catches unhandled Gin context errors and panics.
	// 5. RequestLogger: emits structured request completion log for audit.
	router.Use(otelgin.Middleware(cfg.OtelServiceName))
	router.Use(gin.Recovery())
	router.Use(middleware.RequestContext())
	router.Use(middleware.ErrorHandler(logger))
	router.Use(middleware.RequestLogger(logger))

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
		logger.Info("Starting server",
			zap.String("port", cfg.HttpPort),
			zap.String("service", cfg.OtelServiceName),
			zap.String("env", cfg.AppEnv),
		)
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
