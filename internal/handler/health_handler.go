package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

const serviceVersion = "1.0.0"

// HealthHandler handles health check requests.
type HealthHandler struct{}

// NewHealthHandler creates a new health handler.
func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

// Health handles GET /health
func (h *HealthHandler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "spark-s3-microservice",
		"version": serviceVersion,
		"time":    time.Now().UTC().Format(time.RFC3339),
	})
}