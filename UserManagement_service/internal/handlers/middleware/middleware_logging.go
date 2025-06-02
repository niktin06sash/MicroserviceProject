package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/brokers/kafka"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/metrics"
)

func (m *Middleware) Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ctx := context.WithValue(r.Context(), "starttime", start)
		metrics.UserTotalRequests.WithLabelValues(metrics.NormalizePath(r.URL.Path)).Inc()
		const place = Logging
		traceID := r.Header.Get("X-Trace-ID")
		if traceID == "" {
			m.KafkaProducer.NewUserLog(kafka.LogLevelWarn, place, traceID, "Required Trace-ID")
			traceID = uuid.New().String()
		}
		ctx = context.WithValue(ctx, "traceID", traceID)
		deadlinectx := r.Header.Get("X-Deadline")
		var deadline time.Time
		deadline, err := time.Parse(time.RFC3339, deadlinectx)
		if err != nil {
			fmterr := fmt.Sprintf("Failed to parse X-Deadline: %v", err)
			m.KafkaProducer.NewUserLog(kafka.LogLevelWarn, place, traceID, fmterr)
			deadline = time.Now().Add(15 * time.Second)
		}
		ctx, cancel := context.WithDeadline(ctx, deadline)
		defer cancel()
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}
