package kafka

import (
	"context"
	"encoding/json"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/configs"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/metrics"
	"github.com/segmentio/kafka-go"
)

const (
	LogLevelInfo  = "INFO"
	LogLevelWarn  = "WARN"
	LogLevelError = "ERROR"
)

type UserLog struct {
	Level     string `json:"-"`
	Service   string `json:"service"`
	Place     string `json:"place"`
	TraceID   string `json:"trace_id"`
	Timestamp string `json:"timestamp"`
	Message   string `json:"message"`
}
type KafkaProducer struct {
	writer  *kafka.Writer
	logchan chan UserLog
	wg      *sync.WaitGroup
	context context.Context
	cancel  context.CancelFunc
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
		WriteTimeout:    10 * time.Second,
		WriteBackoffMin: time.Duration(config.RetryBackoffMs) * time.Millisecond,
		WriteBackoffMax: 5 * time.Second,
		BatchSize:       config.BatchSize,
		RequiredAcks:    acks,
	}
	ctx, cancel := context.WithCancel(context.Background())
	logs := make(chan UserLog, 1000)
	producer := &KafkaProducer{
		writer:  w,
		logchan: logs,
		wg:      &sync.WaitGroup{},
		context: ctx,
		cancel:  cancel,
	}
	for i := 1; i <= 3; i++ {
		producer.wg.Add(1)
		go producer.sendLogs(i)
	}
	log.Println("[DEBUG] [User-Service] Successful connect to Kafka-Producer")
	return producer
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
	case kf.logchan <- newlog:
		metrics.UserKafkaProducerBufferSize.Set(float64(len(kf.logchan)))
	default:
		log.Printf("[WARN] [User-Service]  Log channel is full, dropping log: %+v", newlog)
	}
}
func (kf *KafkaProducer) Close() {
	close(kf.logchan)
	kf.cancel()
	kf.wg.Wait()
	kf.writer.Close()
	log.Println("[DEBUG] [User-Service] Successful close Kafka-Producer")
}
func (kf *KafkaProducer) sendLogs(num int) {
	defer kf.wg.Done()
	for {
		select {
		case <-kf.context.Done():
			return
		case logg, ok := <-kf.logchan:
			if !ok {
				log.Printf("[INFO] [User-Service] [Worker: %v] Log channel closed, stopping worker", num)
				return
			}
			metrics.UserKafkaProducerBufferSize.Set(float64(len(kf.logchan)))
			ctx, cancel := context.WithTimeout(kf.context, 5*time.Second)
			defer cancel()
			topic := "user-" + strings.ToLower(logg.Level) + "-log-topic"
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
				metrics.UserErrorsTotal.WithLabelValues("InternalServerError").Inc()
			}
		}
	}
}
