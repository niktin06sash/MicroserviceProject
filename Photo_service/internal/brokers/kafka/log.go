package kafka

import (
	"context"
	"encoding/json"
	"log"
	"strings"
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
			return
		case logg, ok := <-kf.logchan:
			if !ok {
				log.Printf("[INFO] [Photo-Service] [Worker: %v] Log channel closed, stopping worker", num)
				return
			}
			ctx, cancel := context.WithTimeout(kf.context, 5*time.Second)
			defer cancel()
			topic := "photo-" + strings.ToLower(logg.Level) + "-log-topic"
			data, err := json.Marshal(logg)
			if err != nil {
				log.Printf("[ERROR] [Photo-Service] [Worker: %v] Failed to marshal log: %v", num, err)
				continue
			}
		label:
			for i := 0; i < 3; i++ {
				select {
				case <-ctx.Done():
					log.Printf("[WARN] [Photo-Service] [Worker: %v] Context canceled or expired, dropping log: %v", num, err)
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
