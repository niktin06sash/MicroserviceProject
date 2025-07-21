package kafka

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/erro"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/metrics"
	"github.com/segmentio/kafka-go"
)

type UserLog struct {
	Level     string `json:"-"`
	Service   string `json:"service"`
	Place     string `json:"place"`
	TraceID   string `json:"trace_id"`
	Timestamp string `json:"timestamp"`
	Message   string `json:"message"`
}

func (kf *KafkaProducer) NewUserLog(level, place, traceid, msg string) {
	newlog := UserLog{
		Level:     level,
		Service:   "User-Service",
		Place:     place,
		TraceID:   traceid,
		Timestamp: time.Now().Format(time.RFC3339),
		Message:   msg,
	}
	select {
	case <-kf.context.Done():
		log.Printf("[WARN] [User-Service] Producer closing, dropping log: %+v", newlog)
		return
	case kf.logchan <- newlog:
		metrics.UserKafkaProducerBufferSize.Set(float64(len(kf.logchan)))
	default:
		log.Printf("[WARN] [User-Service] Log channel is full, dropping log: %+v", newlog)
	}
}
func (kf *KafkaProducer) sendLogs(num int) {
	defer kf.wg.Done()
	for {
		select {
		case <-kf.context.Done():
			log.Printf("[DEBUG] [User-Service] [Worker: %v] Context canceled, stopping Kafka-worker...", num)
			return
		case logg, ok := <-kf.logchan:
			if !ok {
				return
			}
			metrics.UserKafkaProducerBufferSize.Set(float64(len(kf.logchan)))
			ctx, cancel := context.WithTimeout(kf.context, 5*time.Second)
			defer cancel()
			topic := "user-" + logg.Level + "-log-topic"
			data, err := json.Marshal(logg)
			if err != nil {
				log.Printf("[ERROR] [User-Service] [Worker: %v] Failed to marshal log: %v", num, err)
				continue
			}
		label:
			for i := 0; i < 3; i++ {
				select {
				case <-ctx.Done():
					log.Printf("[WARN] [User-Service] [Worker: %v] Context canceled or expired, dropping log: %v", num, err)
					continue
				default:
					err = kf.writer.WriteMessages(ctx, kafka.Message{
						Topic: topic,
						Key:   []byte(logg.TraceID),
						Value: data,
					})
					if err == nil {
						metrics.UserKafkaProducerMessagesSent.WithLabelValues(topic).Inc()
						break label
					}
					log.Printf("[WARN] [User-Service] [Worker: %v] Retry %d failed to send log: %v", num, i+1, err)
					time.Sleep(1 * time.Second)
				}
			}
			if err != nil {
				log.Printf("[ERROR] [User-Service] [Worker: %v] Failed to send log after all retries: %v, (%v)", num, err, logg)
				metrics.UserKafkaProducerErrorsTotal.WithLabelValues(topic).Inc()
				metrics.UserErrorsTotal.WithLabelValues(erro.ServerErrorType).Inc()
			}
			cancel()
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
			log.Printf("[DEBUG] [User-Service] Context canceled or expired before send Service Log")
			return
		default:
			data, err := json.Marshal(logg)
			if err != nil {
				log.Printf("[DEBUG] [User-Service] Failed to marshal log: %v", err)
				return
			}
			err = kf.writer.WriteMessages(kf.context, kafka.Message{
				Topic: topic,
				Value: data,
			})
			if err != nil {
				log.Printf("[DEBUG] [User-Service] Failed to send Service Log(%v): %v", logg, err)
			}
		}
	}
}
