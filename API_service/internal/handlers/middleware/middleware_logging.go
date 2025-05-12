package middleware

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/niktin06sash/MicroserviceProject/API_service/internal/metrics"
)

func (mw *Middleware) Logging() gin.HandlerFunc {
	return func(c *gin.Context) {
		traceID := uuid.New().String()
		c.Set("traceID", traceID)

		ctx := c.Request.Context()
		ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
		defer cancel()

		c.Request = c.Request.WithContext(ctx)
		start := time.Now()
		c.Set("starttime", start)
		metrics.APITotalRequests.Inc()
		c.Next()
	}
}
