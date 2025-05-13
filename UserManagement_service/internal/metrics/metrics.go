package metrics

import (
	"log"
	"runtime"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var UserTotalRequests = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "user_service_requests_total",
	Help: "Total number of requests to User-Service",
}, []string{"path"})
var UserRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
	Name:    "user_service_duration_seconds",
	Help:    "Histogram for the request duration in seconds in User-Service",
	Buckets: []float64{0.1, 0.5, 1, 2, 5},
}, []string{"handler"})
var UserErrorsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "user_service_errors_total",
	Help: "Total number of errors encountered by the User-Service",
}, []string{"error_type"})
var UserTotalSuccessfulRequests = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "user_service_successful_requests_total",
	Help: "Total number of successful requests to User-Service",
}, []string{"handler"})
var UserBackendRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "user_service_backend_requests_total",
	Help: "Total number of requests forwarded to backend services",
}, []string{"service"})
var UserMemoryUsage = promauto.NewGauge(prometheus.GaugeOpts{
	Name: "user_service_memory_usage_bytes",
	Help: "Current memory usage in bytes",
})
var UserDBQueryDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
	Name:    "user_service_db_query_duration_seconds",
	Help:    "Histogram for the query duration in seconds to the database",
	Buckets: []float64{0.01, 0.05, 0.1, 0.5, 1},
}, []string{"query_type"})
var UserDBQueriesTotal = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "user_service_db_queries_total",
	Help: "Total number of queries executed on the database",
}, []string{"query_type"})
var UserDBErrorsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "user_service_db_errors_total",
	Help: "Total number of errors encountered when interacting with the database",
}, []string{"error_type", "query_type"})
var UserKafkaProducerMessagesSent = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "user_service_kafka_producer_messages_sent_total",
	Help: "Total number of messages sent to Kafka by User-Service",
}, []string{"topics"})
var UserKafkaProducerErrorsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "user_service_kafka_producer_send_errors_total",
	Help: "Total number of errors encountered while sending messages to Kafka by User-Service",
}, []string{"topics"})
var UserKafkaProducerBufferSize = promauto.NewGauge(prometheus.GaugeOpts{
	Name: "user_service_kafka_producer_queue_size",
	Help: "Current size of the Kafka producer message queue in User-Service",
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
				UserMemoryUsage.Set(float64(memStats.Alloc))
			case <-stop:
				return
			}
		}
	}()
}
func Stop() {
	close(stop)
	log.Println("[INFO] [User-Service] Successful close Metrics-Goroutine")
}
