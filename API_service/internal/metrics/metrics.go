package metrics

import (
	"log"
	"runtime"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var APITotalRequests = promauto.NewCounter(prometheus.CounterOpts{
	Name: "api-service_requests_total",
	Help: "Total number of requests to API-Service",
})
var APIRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
	Name:    "api-service_duration_seconds",
	Help:    "Histogram for the request duration in seconds in API-Service",
	Buckets: []float64{0.1, 0.5, 1, 2, 5},
}, []string{"handler"})
var APIErrorsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "api-service_errors_total",
	Help: "Total number of errors encountered by the API Gateway",
}, []string{"error_type"})
var APITotalSuccessfulRequests = promauto.NewCounter(prometheus.CounterOpts{
	Name: "api-service_successful_requests_total",
	Help: "Total number of successful requests to API-Service",
})
var APIBackendRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "api-service_backend_requests_total",
	Help: "Total number of requests forwarded to backend services",
}, []string{"service"})
var APIRateLimitExceededTotal = promauto.NewCounter(prometheus.CounterOpts{
	Name: "api-service_rate_limit_exceeded_total",
	Help: "Total number of requests that exceeded the rate limit",
})
var APIMemoryUsage = promauto.NewGauge(prometheus.GaugeOpts{
	Name: "api-service_memory_usage_bytes",
	Help: "Current memory usage in bytes",
})
var stop = make(chan struct{})

func Start() {
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				var memStats runtime.MemStats
				runtime.ReadMemStats(&memStats)
				APIMemoryUsage.Set(float64(memStats.Alloc))
			case <-stop:
				return
			}
		}
	}()
}
func Stop() {
	close(stop)
	log.Println("[INFO] [API-Service] Successful close Metrics-Goroutine")
}
