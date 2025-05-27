package metrics

import (
	"log"
	"regexp"
	"runtime"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var APITotalRequests = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "api_service_requests_total",
	Help: "Total number of requests to API-Service",
}, []string{"path"})
var APIBadRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
	Name:    "api_service_bad_proxy_requests_duration_seconds",
	Help:    "Histogram for the bad proxy-requests duration in seconds in API-Service",
	Buckets: []float64{0.1, 0.5, 1, 2, 5},
}, []string{"handler", "path"})
var APITotalBadRequests = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "api_service_bad_proxy_requests_total",
	Help: "Total number of bad proxy-requests to API-Service",
}, []string{"path"})
var APIErrorsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "api_service_errors_total",
	Help: "Total number of errors encountered by the API-Service",
}, []string{"error_type"})
var APITotalSuccessfulRequests = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "api_service_successful_proxy_requests_total",
	Help: "Total number of successful proxy-requests to API-Service",
}, []string{"handler", "path"})
var APISuccessfulRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
	Name:    "api_service_successful_proxy_requests_duration_seconds",
	Help:    "Histogram for the successful proxy-requests duration in seconds in API-Service",
	Buckets: []float64{0.1, 0.5, 1, 2, 5},
}, []string{"handler", "path"})
var APIBackendRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "api_service_backend_requests_total",
	Help: "Total number of requests forwarded to backend services",
}, []string{"service"})
var APIRateLimitExceededTotal = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "api_service_rate_limit_exceeded_total",
	Help: "Total number of requests that exceeded the rate limit",
}, []string{"path"})
var APIMemoryUsage = promauto.NewGauge(prometheus.GaugeOpts{
	Name: "api_service_memory_usage_bytes",
	Help: "Current memory usage in bytes",
})
var APIKafkaProducerMessagesSent = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "api_service_kafka_producer_messages_sent_total",
	Help: "Total number of messages sent to Kafka by API-Service",
}, []string{"topics"})
var APIKafkaProducerErrorsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "api_service_kafka_producer_send_errors_total",
	Help: "Total number of errors encountered while sending messages to Kafka by API-Service",
}, []string{"topics"})
var APIKafkaProducerBufferSize = promauto.NewGauge(prometheus.GaugeOpts{
	Name: "api_service_kafka_producer_queue_size",
	Help: "Current size of the Kafka producer message queue in API-Service",
})

func NormalizePath(path string) string {
	re := regexp.MustCompile(`[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}`)
	normalizedPath := re.ReplaceAllString(path, "")
	return normalizedPath
}

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
	log.Println("[DEBUG] [API-Service] Successful close Metrics-Goroutine")
}
