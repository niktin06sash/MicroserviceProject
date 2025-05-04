package kafka

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"sync"
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
	Service   string `json:"service"`
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
	wg      *sync.WaitGroup
}
type KafkaProducerService interface {
	NewAPILog(c *http.Request, level, place, traceid, msg string)
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
		Balancer:        &kafka.Hash{},
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
		wg:      &sync.WaitGroup{},
	}
	startmsg := "Successful connect to Kafka-Producer"
	startlog := APILog{
		Level:     LogLevelInfo,
		Service:   "API-Service",
		Timestamp: time.Now().Format(time.RFC3339),
		Message:   startmsg,
	}
	producer.wg.Add(1)
	go func() {
		defer producer.wg.Done()
		producer.sendLogs()
	}()
	producer.logchan <- startlog
	log.Println("[INFO] [API-Service] [KafkaProducer] Successful connect to Kafka-Producer")
	return producer
}

func (kf *KafkaProducer) Close() {
	close(kf.logchan)
	kf.wg.Wait()
	kf.writer.Close()
	log.Println("[INFO] [API-Service] [KafkaProducer] Successful close Kafka-Producer")
}
func (kf *KafkaProducer) NewAPILog(c *http.Request, level, place, traceid, msg string) {
	if err := c.Context().Err(); err != nil {
		log.Printf("[WARN] [API-Service] [KafkaProducer] Context canceled or expired, dropping log: %v", err)
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
	default:
		log.Printf("[WARN] [API-Service] [KafkaProducer] Log channel is full, dropping log: %+v", newlog)
	}
}
func (kf *KafkaProducer) sendLogs() {
	for logg := range kf.logchan {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		topic := strings.ToLower(logg.Level) + "-log-topic"
		data, err := json.Marshal(logg)
		if err != nil {
			log.Printf("[ERROR] [API-Service] [KafkaProducer] Failed to marshal log: %v", err)
			cancel()
			continue
		}
		for i := 0; i < 3; i++ {
			err = kf.writer.WriteMessages(ctx, kafka.Message{
				Topic: topic,
				Key:   []byte(logg.TraceID),
				Value: data,
			})
			if err == nil {
				break
			}
			log.Printf("[WARN] [API-Service] [KafkaProducer] Retry %d failed to send log: %v", i+1, err)
			time.Sleep(1 * time.Second)
		}
		if err := ctx.Err(); err != nil {
			log.Printf("[WARN] [API-Service] [KafkaProducer] Context canceled or expired, dropping log: %v", err)
			cancel()
			continue
		}
		if err != nil {
			log.Printf("[ERROR] [API-Service] [KafkaProducer] Failed to send log after all retries: %v", err)
			cancel()
		}
		cancel()
	}
}
