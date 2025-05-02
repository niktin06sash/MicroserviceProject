package kafka

import (
	"context"
	"encoding/json"
	"log"
	"strings"
	"time"

	"github.com/niktin06sash/MicroserviceProject/API_service/internal/configs"
	"github.com/segmentio/kafka-go"
)

const (
	LogLevelInfo  = "INFO"
	LogLevelWarn  = "WARN"
	LogLevelError = "ERROR"
)

type APILog struct {
	Level     string `json:"level"`
	Place     string `json:"place"`
	TraceID   string `json:"trace_id"`
	IP        string `json:"ip"`
	Method    string `json:"method,omitempty"`
	Path      string `json:"path,omitempty"`
	Timestamp string `json:"timestamp"`
	Message   string `json:"message"`
}

type KafkaProducer struct {
	writer  *kafka.Writer
	logchan chan APILog
}
type KafkaProducerService interface {
	Close()
	NewAPILog(log APILog)
	sendLogs()
}

func NewKafkaProducer(config configs.KafkaConfig) *KafkaProducer {
	brokersString := config.BootstrapServers
	brokers := strings.Split(brokersString, ",")
	var acks kafka.RequiredAcks
	switch config.Acks {
	case "0":
		acks = kafka.RequireNone
	case "1":
		acks = kafka.RequireOne
	case "all":
		acks = kafka.RequireAll
	default:
		acks = kafka.RequireAll
	}
	w := &kafka.Writer{
		Addr:            kafka.TCP(brokers...),
		Topic:           "",
		Balancer:        &kafka.LeastBytes{},
		WriteTimeout:    10 * time.Second,
		WriteBackoffMin: time.Duration(config.RetryBackoffMs) * time.Millisecond,
		WriteBackoffMax: 5 * time.Second,
		BatchSize:       config.BatchSize,
		RequiredAcks:    acks,
	}
	logs := make(chan APILog, 1000)

	producer := &KafkaProducer{
		writer:  w,
		logchan: logs,
	}

	go producer.sendLogs()
	log.Println("[INFO] [API-Service] [KafkaProducer] Successful connect to Kafka-Producer")
	return producer
}

func (kf *KafkaProducer) Close() {
	kf.writer.Close()
	close(kf.logchan)
	log.Println("[INFO] [API-Service] [KafkaProducer] Successful close Kafka-Producer")
}
func (kf *KafkaProducer) NewAPILog(logg APILog) {
	select {
	case kf.logchan <- logg:
	default:
		log.Printf("[WARN] [API-Service] [KafkaProducer] Log channel is full, dropping log: %+v", logg)
	}
}
func (kf *KafkaProducer) sendLogs() {
	for logg := range kf.logchan {
		topic := "api-" + strings.ToLower(logg.Level) + "-log-topic"
		data, err := json.Marshal(logg)
		if err != nil {
			log.Printf("[ERROR] [API-Service] [KafkaProducer] Failed to marshal log: %v", err)
			return
		}
		retries := 3
		for i := 0; i < retries; i++ {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			err = kf.writer.WriteMessages(ctx, kafka.Message{
				Topic: topic,
				Value: data,
			})
			if err == nil {
				break
			}
			log.Printf("[WARN] [API-Service] [KafkaProducer] Retry %d/%d failed to send log: %v", i+1, retries, err)
			time.Sleep(1 * time.Second)
		}
		if err != nil {
			log.Printf("[ERROR] [API-Service] [KafkaProducer] Failed to send log after all retries: %v", err)
		}
	}
}
