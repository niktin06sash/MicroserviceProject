package middleware

import (
	"context"
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
			m.LogProducer.NewUserLog(kafka.LogLevelWarn, place, traceID, "Required Trace-ID")
			traceID = uuid.New().String()
		}
		ctx = context.WithValue(ctx, "traceID", traceID)
		r = r.WithContext(ctx)
		m.LogProducer.NewUserLog(kafka.LogLevelInfo, place, traceID, "New request has been received")
		next.ServeHTTP(w, r)
	})
}
