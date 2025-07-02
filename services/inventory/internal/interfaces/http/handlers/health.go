package handlers

import (
	"net/http"

	"services/inventory/internal/infrastructure/postgres"
	"services/inventory/internal/infrastructure/redis"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type HealthHandler struct {
	redisClient *redis.Client
	dbConn      *postgres.Connection
	logger      *logrus.Logger
}

func NewHealthHandler(redisClient *redis.Client, dbConn *postgres.Connection, logger *logrus.Logger) *HealthHandler {
	return &HealthHandler{
		redisClient: redisClient,
		dbConn:      dbConn,
		logger:      logger,
	}
}

// HealthCheck performs basic health check
func (h *HealthHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "inventory-service",
		"version": "1.0.0",
	})
}

// ReadinessCheck performs comprehensive readiness check
func (h *HealthHandler) ReadinessCheck(c *gin.Context) {
	checks := gin.H{
		"redis":    h.checkRedis(c),
		"database": h.checkDatabase(),
	}

	// Determine overall status
	allHealthy := true
	for _, status := range checks {
		if status != "healthy" {
			allHealthy = false
			break
		}
	}

	statusCode := http.StatusOK
	if !allHealthy {
		statusCode = http.StatusServiceUnavailable
	}

	c.JSON(statusCode, gin.H{
		"status": func() string {
			if allHealthy {
				return "ready"
			}
			return "not_ready"
		}(),
		"checks": checks,
	})
}

func (h *HealthHandler) checkRedis(c *gin.Context) string {
	if err := h.redisClient.Ping(c.Request.Context()); err != nil {
		h.logger.WithError(err).Error("Redis health check failed")
		return "unhealthy"
	}
	return "healthy"
}

func (h *HealthHandler) checkDatabase() string {
	if err := h.dbConn.DB.Ping(); err != nil {
		h.logger.WithError(err).Error("Database health check failed")
		return "unhealthy"
	}
	return "healthy"
}
