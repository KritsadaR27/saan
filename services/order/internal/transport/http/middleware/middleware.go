package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"order/pkg/logger"
)

// LoggingMiddleware logs HTTP requests
func LoggingMiddleware(logger logger.Logger) gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		logger.WithFields(map[string]interface{}{
			"client_ip":   param.ClientIP,
			"timestamp":   param.TimeStamp.Format(time.RFC3339),
			"method":      param.Method,
			"path":        param.Path,
			"status_code": param.StatusCode,
			"latency":     param.Latency,
			"user_agent":  param.Request.UserAgent(),
		}).Info("HTTP Request")
		return ""
	})
}

// CORSMiddleware handles CORS headers
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Header("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// ErrorHandlingMiddleware handles panics and errors
func ErrorHandlingMiddleware(logger logger.Logger) gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		logger.WithField("error", recovered).Error("Panic recovered")
		c.JSON(500, gin.H{
			"error": "Internal server error",
		})
		c.Abort()
	})
}
