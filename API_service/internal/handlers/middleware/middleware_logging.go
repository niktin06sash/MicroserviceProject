package middleware

import (
	"context"
	"log"
	"net/http"

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
func LoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		traceID := uuid.New()
		ctx := context.WithValue(c.Request.Context(), "traceID", traceID)
		c.Request = c.Request.WithContext(ctx)
		logRequest(c.Request, "Logging", traceID.String(), false, "")
		c.Next()
	}
}
