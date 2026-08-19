package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sainakuo/scalable-notification-system/internal/health"
)

type HealthHandler struct {
	postgres health.Checker
	redis    health.Checker
}

func NewHealthHandler(
	postgres health.Checker,
	redis health.Checker,
) *HealthHandler {
	if postgres == nil {
		panic("postgres health checker is nil")
	}

	if redis == nil {
		panic("redis health checker is nil")
	}

	return &HealthHandler{
		postgres: postgres,
		redis:    redis,
	}
}

func (h *HealthHandler) Check(c *gin.Context) {
	ctx, cancel := context.WithTimeout(
		c.Request.Context(),
		2*time.Second,
	)
	defer cancel()

	statusCode := http.StatusOK
	status := "ok"

	checks := gin.H{
		"postgres": "ok",
		"redis":    "ok",
	}

	if err := h.postgres.Check(ctx); err != nil {
		statusCode = http.StatusServiceUnavailable
		status = "unhealthy"
		checks["postgres"] = "unavailable"
	}

	if err := h.redis.Check(ctx); err != nil {
		statusCode = http.StatusServiceUnavailable
		status = "unhealthy"
		checks["redis"] = "unavailable"
	}

	c.JSON(statusCode, gin.H{
		"status": status,
		"checks": checks,
	})
}
