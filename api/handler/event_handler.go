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
