package handler

import (
	"errors"
	"net/http"

	"github.com/AmithSAI007/prj-apex-transcode-submitter/api/dto"
	"github.com/AmithSAI007/prj-apex-transcode-submitter/api/validation"
	"github.com/AmithSAI007/prj-apex-transcode-submitter/internal/config"
	"github.com/AmithSAI007/prj-apex-transcode-submitter/internal/service"
	"github.com/AmithSAI007/prj-apex-transcode-submitter/pkg/constants"
	"github.com/AmithSAI007/prj-apex-transcode-submitter/pkg/utils"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

const (
	ContentTypeHeader   = "Content-Type"
	ContentLengthHeader = "Content-Length"
)

type EventHandler struct {
	logger *zap.Logger
	s      service.SubmitterInterface
	cfg    config.Config
}

func NewEventHandler(logger *zap.Logger, s service.SubmitterInterface, cfg config.Config) *EventHandler {
	return &EventHandler{
		logger: logger,
		s:      s,
		cfg:    cfg,
	}
}

// HandleTranscodeTask godoc
//
//	@Summary		Submit a video transcoding task
//	@Description	Receives a transcode task payload from Google Cloud Tasks, validates the
//	@Description	Cloud Tasks headers and JSON body, verifies the raw media file exists in GCS,
//	@Description	checks the Firestore document state for idempotency, and submits a transcoding
//	@Description	job to the Google Cloud Transcoder API. On success the Firestore document is
//	@Description	transitioned from "queued" to "transcoding".
//	@Description
//	@Description	**HTTP status code semantics (Cloud Tasks retry behaviour):**
//	@Description	- **200** — Task processed successfully, or duplicate detected (idempotent). Cloud Tasks will not retry.
//	@Description	- **400** — Permanent failure (bad payload, missing headers, invalid state). Cloud Tasks will not retry.
//	@Description	- **500** — Transient failure (infrastructure error). Cloud Tasks will retry with exponential back-off.
//	@ID				submitTranscodeTask
//	@Tags			transcode
//	@Accept			json
//	@Produce		json
//	@Param			X-CloudTasks-TaskName		header		string				true	"Cloud Tasks task name"		example(projects/my-project/locations/us-central1/queues/transcode-tasks/tasks/task-123)
//	@Param			X-CloudTasks-QueueName		header		string				true	"Cloud Tasks queue name"	example(transcode-tasks)
//	@Param			X-CloudTasks-RetryCount		header		int					false	"Cloud Tasks retry count"	example(0)
//	@Param			request						body		dto.TranscodeRequest	true	"Transcode task payload"
//	@Success		200		{object}	dto.TranscodeResponse	"Task accepted or duplicate acknowledged"
//	@Failure		400		{object}	dto.ErrorResponse		"Permanent error — invalid headers, malformed body, or domain validation failure"
//	@Failure		500		{object}	dto.ErrorResponse		"Transient error — infrastructure failure, retry is expected"
//	@Router			/ [post]
func (h *EventHandler) HandleTranscodeTask(c *gin.Context) {
	ctx := c.Request.Context()

	// The otelgin middleware creates the root HTTP span. We create a child
	// span here for handler-level logic so the root span stays clean.
	ctx, span := utils.Tracer().Start(ctx, "handler.HandleTranscodeTask")
	defer span.End()

	traceID := utils.TraceIDFromContext(ctx)
	logFields := append(utils.LogFieldsFromContext(ctx),
		zap.String(constants.LogKeyLayer, "handler"),
		zap.String(constants.LogKeyMethod, "HandleTranscodeTask"),
	)

	h.logger.Debug("received transcode task request", logFields...)

	headerMap := make(map[string]string)
	for key, values := range c.Request.Header {
		if len(values) > 0 {
			headerMap[key] = values[0]
		}
	}

	if err := validation.ValidateCloudTaskHeaders(headerMap, h.cfg.TranscodeTaskQueue); err != nil {
		span.SetStatus(codes.Error, "header validation failed")
		span.RecordError(err)
		h.respondToHeaderError(c, traceID, logFields, err)
		return
	}
	span.AddEvent("headers.validated")

	var req dto.TranscodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		span.SetStatus(codes.Error, "bind failed")
		span.RecordError(err)
		h.respondToBindError(c, traceID, logFields, err)
		return
	}

	// Enrich span and logs with business identifiers now that we have them.
	span.SetAttributes(
		attribute.String(constants.LogKeyVideoID, req.VideoID),
		attribute.String(constants.LogKeyUserID, req.UserID),
		attribute.String("event_id", req.EventID),
	)
	logFields = append(logFields,
		zap.String(constants.LogKeyVideoID, req.VideoID),
		zap.String(constants.LogKeyUserID, req.UserID),
		zap.String("event_id", req.EventID),
	)
	span.AddEvent("request.bound")

	h.logger.Info("processing transcode task",
		append(logFields,
			zap.Int64("file_size", req.FileSize),
			zap.String("content_type", req.ContentType),
		)...,
	)

	// Update the request context so downstream spans inherit the enriched context.
	c.Request = c.Request.WithContext(ctx)

	if err := h.s.ProcessTranscodeRequest(c.Request.Context(), &req); err != nil {
		h.respondToServiceResult(c, traceID, logFields, span, err)
		return
	}

	span.SetStatus(codes.Ok, "task processed")
	span.AddEvent("task.completed")

	h.logger.Info("transcode task accepted", logFields...)

	c.JSON(http.StatusOK, dto.TranscodeResponse{
		Message: "Transcode task received successfully",
	})
}

// respondToHeaderError handles Cloud Task header validation failures.
// All header errors are permanent (400) since retrying with the same
// headers will produce the same result.
func (h *EventHandler) respondToHeaderError(c *gin.Context, traceID string, logFields []zap.Field, err error) {
	h.logger.Warn("header validation failed",
		append(logFields, zap.Error(err))...,
	)

	switch {
	case errors.Is(err, validation.ErrMissingCloudTaskHeader):
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: dto.ErrorPayload{
				Code:      dto.ErrorCodeInvalidRequest,
				Message:   "Missing required Cloud Task header",
				RequestID: traceID,
			},
		})
	case errors.Is(err, validation.ErrInvalidCloudTasksQueueName):
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: dto.ErrorPayload{
				Code:      dto.ErrorCodeInvalidRequest,
				Message:   "Invalid Cloud Tasks queue name",
				RequestID: traceID,
			},
		})
	default:
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: dto.ErrorPayload{
				Code:      dto.ErrorCodeInvalidRequest,
				Message:   "Invalid Cloud Task headers",
				RequestID: traceID,
			},
		})
	}
}

// respondToBindError handles JSON binding and parsing failures.
// These are permanent errors (400) since the request body is malformed
// and retrying will not fix it.
func (h *EventHandler) respondToBindError(c *gin.Context, traceID string, logFields []zap.Field, err error) {
	h.logger.Warn("request binding failed",
		append(logFields, zap.Error(err))...,
	)

	c.JSON(http.StatusBadRequest, dto.ErrorResponse{
		Error: dto.ErrorPayload{
			Code:      dto.ErrorCodeInvalidRequest,
			Message:   "Invalid request body",
			RequestID: traceID,
			Details: []dto.ErrorDetail{
				{Message: err.Error()},
			},
		},
	})
}

// respondToServiceResult inspects the error returned by the service layer
// and maps it to the appropriate HTTP status code:
//   - PermanentError  -> 400 (Bad Request)  — do not retry
//   - IdempotencyError -> 200 (OK)          — already processed, acknowledge to stop retries
//   - TransientError  -> 500 (Internal Server Error) — Cloud Tasks should retry
//   - Unknown errors  -> 500 (Internal Server Error) — treated as transient by default
func (h *EventHandler) respondToServiceResult(c *gin.Context, traceID string, logFields []zap.Field, span trace.Span, err error) {
	var permErr *service.PermanentError
	var idempErr *service.IdempotencyError
	var transErr *service.TransientError

	switch {
	case errors.As(err, &permErr):
		span.SetStatus(codes.Error, "permanent error")
		h.logger.Warn("permanent service error",
			append(logFields, zap.Error(err))...,
		)
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: dto.ErrorPayload{
				Code:      dto.ErrorCodeInvalidRequest,
				Message:   permErr.Error(),
				RequestID: traceID,
			},
		})

	case errors.As(err, &idempErr):
		span.SetStatus(codes.Ok, "idempotent duplicate")
		span.AddEvent("idempotency.duplicate_acknowledged")
		h.logger.Info("idempotent duplicate detected",
			append(logFields, zap.Error(err))...,
		)
		// Return 200 so Cloud Tasks considers the task complete and stops retrying.
		c.JSON(http.StatusOK, dto.TranscodeResponse{
			Message: idempErr.Error(),
		})

	case errors.As(err, &transErr):
		span.SetStatus(codes.Error, "transient error")
		h.logger.Error("transient service error",
			append(logFields, zap.Error(err))...,
		)
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: dto.ErrorPayload{
				Code:      dto.ErrorCodeInternal,
				Message:   "An internal error occurred, please retry",
				RequestID: traceID,
			},
		})

	default:
		// Unknown error type — treat as transient to allow Cloud Tasks to retry.
		span.SetStatus(codes.Error, "unexpected error")
		h.logger.Error("unexpected service error",
			append(logFields, zap.Error(err))...,
		)
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: dto.ErrorPayload{
				Code:      dto.ErrorCodeInternal,
				Message:   "An internal error occurred, please retry",
				RequestID: traceID,
			},
		})
	}
}
