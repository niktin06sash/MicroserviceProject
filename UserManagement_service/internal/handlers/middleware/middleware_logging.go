package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/kafka"
)

func (m *Middleware) Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var place = "Logging"
		traceID := r.Header.Get("X-Trace-ID")
		if traceID == "" {
			m.KafkaProducer.NewUserLog(kafka.LogLevelError, place, traceID, "Required Trace-ID")
			traceID = uuid.New().String()
		}
		ctx := context.WithValue(r.Context(), "traceID", traceID)
		deadlinectx := r.Header.Get("X-Deadline")
		var deadline time.Time
		deadline, err := time.Parse(time.RFC3339, deadlinectx)
		if err != nil {
			fmterr := fmt.Sprintf("Failed to parse X-Deadline: %v", err)
			m.KafkaProducer.NewUserLog(kafka.LogLevelError, place, traceID, fmterr)
			deadline = time.Now().Add(15 * time.Second)
		}
		ctx, cancel := context.WithDeadline(ctx, deadline)
		defer cancel()
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}
