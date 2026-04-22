package handlers

import (
	"net/http"

	"github.com/amirrajj-dev/taskio/internal/utils"
	"github.com/gin-gonic/gin"
)

type HealthHandler struct{}

func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

// HealthCheck godoc
// @Summary      Health check
// @Description  Check if the API server is running
// @Tags         Health
// @Accept       json
// @Produce      json
// @Success      200 {object} utils.SuccessResponse{data=string} "Server is healthy"
// @Router       /health [get]
func (h *HealthHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, utils.NewSuccessResponse("OK", nil, c.Request.URL.Path))
}