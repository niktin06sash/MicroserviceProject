package middleware

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/niktin06sash/MicroserviceProject/API_service/internal/kafka"
)

func (mw *Middleware) Logging() gin.HandlerFunc {
	return func(c *gin.Context) {
		traceID := uuid.New().String()

		c.Set("traceID", traceID)

		ctx := c.Request.Context()
		ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
		defer cancel()

		c.Request = c.Request.WithContext(ctx)
		mw.KafkaProducer.NewAPILog(kafka.APILog{
			Level:     kafka.LogLevelInfo,
			Place:     "Logging",
			TraceID:   traceID,
			IP:        c.Request.RemoteAddr,
			Method:    c.Request.Method,
			Path:      c.Request.URL.Path,
			Timestamp: time.Now().Format(time.RFC3339),
			Message:   "",
		})
		c.Next()
	}
}
