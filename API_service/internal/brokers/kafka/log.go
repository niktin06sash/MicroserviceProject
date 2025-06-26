package kafka

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/niktin06sash/MicroserviceProject/API_service/internal/erro"
	"github.com/niktin06sash/MicroserviceProject/API_service/internal/metrics"
	"github.com/segmentio/kafka-go"
)

type APILog struct {
	Level     string `json:"-"`
	Service   string `json:"service"`
	Place     string `json:"place"`
	TraceID   string `json:"trace_id"`
	IP        string `json:"ip"`
	Method    string `json:"method,omitempty"`
	Path      string `json:"path,omitempty"`
	Timestamp string `json:"timestamp"`
	Message   string `json:"message"`
}

func (kf *KafkaProducer) NewAPILog(c *http.Request, level, place, traceid, msg string) {
	if err := c.Context().Err(); err != nil {
		log.Printf("[WARN] [API-Service] Context canceled or expired, dropping log: %v", err)
		return
	}
	newlog := APILog{
		Level:     level,
		Service:   "API-Service",
		Place:     place,
		TraceID:   traceid,
		IP:        c.RemoteAddr,
		Method:    c.Method,
		Path:      c.URL.Path,
		Timestamp: time.Now().Format(time.RFC3339),
		Message:   msg,
	}
	select {
	case kf.logchan <- newlog:
		metrics.APIKafkaProducerBufferSize.Set(float64(len(kf.logchan)))
	default:
		log.Printf("[WARN] [API-Service] Log channel is full, dropping log: %+v", newlog)
	}
}
func (kf *KafkaProducer) sendLogs(num int) {
	defer kf.wg.Done()
	for {
		select {
		case <-kf.context.Done():
			log.Printf("[WARN] [API-Service] [Worker: %v] Context canceled", num)
			return
		case logg, ok := <-kf.logchan:
			if !ok {
				log.Printf("[INFO] [API-Service] [Worker: %v] Log channel closed, stopping worker", num)
				return
			}
			metrics.APIKafkaProducerBufferSize.Set(float64(len(kf.logchan)))
			ctx, cancel := context.WithTimeout(kf.context, 5*time.Second)
			defer cancel()
			topic := "api-" + logg.Level + "-log-topic"
			data, err := json.Marshal(logg)
			if err != nil {
				log.Printf("[ERROR] [API-Service] [Worker: %v] Failed to marshal log: %v", num, err)
				continue
			}
		label:
			for i := 0; i < 3; i++ {
				select {
				case <-ctx.Done():
					log.Printf("[WARN] [API-Service] [Worker: %v] Context canceled or expired, dropping log: %v", num, err)
					continue
				default:
					err = kf.writer.WriteMessages(ctx, kafka.Message{
						Topic: topic,
						Key:   []byte(logg.TraceID),
						Value: data,
					})
					if err == nil {
						metrics.APIKafkaProducerMessagesSent.WithLabelValues(topic).Inc()
						break label
					}
					log.Printf("[WARN] [API-Service] [Worker: %v] Retry %d failed to send log: %v", num, i+1, err)
					time.Sleep(1 * time.Second)
				}
			}
			if err != nil {
				log.Printf("[ERROR] [API-Service] [Worker: %v] Failed to send log after all retries: %v, (%v)", num, err, logg)
				metrics.APIKafkaProducerErrorsTotal.WithLabelValues(topic).Inc()
				metrics.APIErrorsTotal.WithLabelValues(erro.ServerErrorType).Inc()
			}
		}
	}
}

type serviceLog struct {
	Message string `json:"service_log"`
}

func (kf *KafkaProducer) LogStart() {
	kf.sendServiceLog(serviceLog{Message: logStartService})
}
func (kf *KafkaProducer) LogClose() {
	kf.sendServiceLog(serviceLog{Message: logCloseService})
}
func (kf *KafkaProducer) sendServiceLog(logg serviceLog) {
	for _, topic := range kf.topics {
		select {
		case <-kf.context.Done():
			log.Printf("[DEBUG] [API-Service] Context canceled or expired before send Service Log")
			return
		default:
			data, err := json.Marshal(logg)
			if err != nil {
				log.Printf("[DEBUG] [API-Service] Failed to marshal log: %v", err)
				return
			}
			err = kf.writer.WriteMessages(kf.context, kafka.Message{
				Topic: topic,
				Value: data,
			})
			if err != nil {
				log.Printf("[DEBUG] [API-Service] Failed to send Service Log(%v): %v", logg, err)
			}
		}
	}
}
