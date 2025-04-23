package middleware

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func logRequest(r *http.Request, place string, requestID string, isError bool, errorMessage string) {
	ip := r.RemoteAddr
	method := r.Method
	path := r.URL.Path
	if isError {
		log.Printf("[ERROR] [API-Service] [%s] [TraceID: %s] [IP: %s] [Method: %s] [Path: %s] Error: %s", place, requestID, ip, method, path, errorMessage)
	} else {
		log.Printf("[INFO] [API-Service] [%s] [TraceID: %s] [IP: %s] [Method: %s] [Path: %s]", place, requestID, ip, method, path)
	}
}
func Middleware_Logging() gin.HandlerFunc {
	return func(c *gin.Context) {
		traceID := uuid.New().String()

		c.Set("traceID", traceID)

		ctx := c.Request.Context()
		ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()

		c.Request = c.Request.WithContext(ctx)
		logRequest(c.Request, "Logging", traceID, false, "")
		c.Next()
	}
}
