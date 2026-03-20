package handler

import (
	"net/http"

	"github.com/AmithSAI007/prj-apex-transcode-submitter/api/dto"
	"github.com/gin-gonic/gin"
)

// HealthCheck godoc
//
//	@Summary		Service health check
//	@Description	Returns the current health status of the service. Used by load balancers,
//	@Description	Cloud Run readiness probes, and monitoring systems to verify the service
//	@Description	is running and able to accept traffic.
//	@ID				healthCheck
//	@Tags			operations
//	@Produce		json
//	@Success		200	{object}	dto.HealthResponse	"Service is healthy"
//	@Router			/healthz [get]
func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, dto.HealthResponse{
		Status: "ok",
	})
}
