package kafka

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
)

type PhotoLog struct {
	Level     string `json:"-"`
	Service   string `json:"service"`
	Place     string `json:"place"`
	TraceID   string `json:"trace_id"`
	Timestamp string `json:"timestamp"`
	Message   string `json:"message"`
}

func (kf *KafkaProducer) NewPhotoLog(level, place, traceid, msg string) {
	newlog := PhotoLog{
		Level:     level,
		Service:   "Photo-Service",
		Place:     place,
		TraceID:   traceid,
		Timestamp: time.Now().Format(time.RFC3339),
		Message:   msg,
	}
	select {
	case <-kf.context.Done():
		log.Printf("[WARN] [Photo-Service] Producer closing, dropping log: %+v", newlog)
		return
	case kf.logchan <- newlog:
	default:
		log.Printf("[WARN] [Photo-Service] Log channel is full, dropping log: %+v", newlog)
	}
}
func (kf *KafkaProducer) sendLogs(num int) {
	defer kf.wg.Done()
	for {
		select {
		case <-kf.context.Done():
			log.Printf("[DEBUG] [Photo-Service] [Worker: %v] Context canceled, stopping Kafka-worker...", num)
			return
		case logg, ok := <-kf.logchan:
			if !ok {
				return
			}
			ctx, cancel := context.WithTimeout(kf.context, 5*time.Second)
			defer cancel()
			topic := "photo-" + logg.Level + "-log-topic"
			data, err := json.Marshal(logg)
			if err != nil {
				log.Printf("[ERROR] [Photo-Service] [Worker: %v] Failed to marshal log: %v", num, err)
				continue
			}
		label:
			for i := 0; i < 3; i++ {
				select {
				case <-ctx.Done():
					log.Printf("[WARN] [Photo-Service] [Worker: %v] Context canceled or expired, dropping log: %v", num, logg)
					continue
				default:
					err = kf.writer.WriteMessages(ctx, kafka.Message{
						Topic: topic,
						Key:   []byte(logg.TraceID),
						Value: data,
					})
					if err == nil {
						break label
					}
					log.Printf("[WARN] [Photo-Service] [Worker: %v] Retry %d failed to send log: %v", num, i+1, err)
					time.Sleep(1 * time.Second)
				}
			}
			if err != nil {
				log.Printf("[ERROR] [Photo-Service] [Worker: %v] Failed to send log after all retries: %v, (%v)", num, err, logg)
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
			log.Printf("[DEBUG] [Photo-Service] Context canceled or expired before send Service Log")
			return
		default:
			data, err := json.Marshal(logg)
			if err != nil {
				log.Printf("[DEBUG] [Photo-Service] Failed to marshal log: %v", err)
				return
			}
			err = kf.writer.WriteMessages(kf.context, kafka.Message{
				Topic: topic,
				Value: data,
			})
			if err != nil {
				log.Printf("[DEBUG] [Photo-Service] Failed to send Service Log(%v): %v", logg, err)
			}
		}
	}
}
