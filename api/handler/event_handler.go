package handler

import (
	"errors"

	"github.com/AmithSAI007/prj-apex-transcode-submitter/api/dto"
	"github.com/AmithSAI007/prj-apex-transcode-submitter/pkg/utils"
	"github.com/AmithSAI007/prj-apex-transcode-submitter/pkg/validation"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	otrace "go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// HandleTranscodeTask godoc
//
//	@Summary		Submit transcode task
//	@Tags			transcode
//	@Accept			json
//	@Produce		json
//	@Param			request	body		dto.TranscodeRequest	true	"Transcode request payload"
//	@Success		200		{object}	dto.TranscodeResponse
//	@Failure		400		{object}	dto.ErrorResponse
//	@Failure		415		{object}	dto.ErrorResponse
//	@Failure		500		{object}	dto.ErrorResponse
//	@Router			/ [post]

const (
	maxBodySize     = 10 * 1024
	applicationJSON = "application/json"
)

type EventHandler struct {
	logger *zap.Logger
}

func NewEventHandler(logger *zap.Logger) *EventHandler {
	return &EventHandler{
		logger: logger,
	}
}

func (h *EventHandler) HandleTranscodeTask(c *gin.Context) {

	h.logger.Info("Received transcode task request")

	h.logger.Info("Transcode task request validated successfully", zap.Any("headers", c.Request.Header))

	c.JSON(200, dto.TranscodeResponse{
		Message: "Transcode task received successfully",
	})

}

func (h *EventHandler) respondWithServiceError(c *gin.Context, err error, message string) {
	traceID := utils.TraceIDFromContext(c.Request.Context())
	requestID := traceID

	span := otrace.SpanFromContext(c.Request.Context())
	if span != nil && span.IsRecording() {
		span.RecordError(err)
		span.SetStatus(codes.Error, message)
	}

	switch {
	case errors.Is(err, validation.ErrMalformedJSON):
		if span != nil && span.IsRecording() {
			span.AddEvent("validation.invalid_path", otrace.WithAttributes(
				attribute.String("error", err.Error()),
			))
		}
		c.JSON(400, dto.ErrorResponse{
			Error: dto.ErrorPayload{Code: dto.ErrorCodeInvalidRequest, Message: message, RequestID: requestID},
		})
	case errors.Is(err, validation.ErrInvalidContentType):
		if span != nil && span.IsRecording() {
			span.AddEvent("validation.invalid_content_type", otrace.WithAttributes(
				attribute.String("error", err.Error()),
			))
		}
		c.JSON(415, dto.ErrorResponse{
			Error: dto.ErrorPayload{Code: dto.ErrorCodeInvalidContentType, Message: message, RequestID: requestID},
		})
	case errors.Is(err, validation.ErrEmptyBody), errors.Is(err, validation.ErrPayloadTooLarge):
		if span != nil && span.IsRecording() {
			span.AddEvent("validation.request_error", otrace.WithAttributes(
				attribute.String("error", err.Error()),
			))
		}
		c.JSON(400, dto.ErrorResponse{
			Error: dto.ErrorPayload{Code: dto.ErrorCodeEmptyBody, Message: message, RequestID: requestID},
		})
	default:
		if span != nil && span.IsRecording() {
			span.AddEvent("internal.error", otrace.WithAttributes(
				attribute.String("error", err.Error()),
			))
		}
		c.JSON(500, dto.ErrorResponse{
			Error: dto.ErrorPayload{Code: dto.ErrorCodeInternal, Message: message, RequestID: requestID},
		})
	}
}
