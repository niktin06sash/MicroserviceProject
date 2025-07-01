package kafka

import (
	"context"
	"log"
	"strings"
	"sync"
	"time"

	configs "github.com/niktin06sash/MicroserviceProject/Logs_service/internal/configs"
	"github.com/niktin06sash/MicroserviceProject/Logs_service/internal/logs"
	"github.com/segmentio/kafka-go"
)

const (
	LogLevelInfo  = "info"
	LogLevelWarn  = "warn"
	LogLevelError = "error"
)

type KafkaConsumer struct {
	reader  *kafka.Reader
	wg      *sync.WaitGroup
	logger  *logs.Logger
	ctx     context.Context
	cancel  context.CancelFunc
	counter int64
}

func NewKafkaConsumer(config configs.KafkaConfig, logger *logs.Logger, topic string) *KafkaConsumer {
	brokersString := config.BootstrapServers
	brokers := strings.Split(brokersString, ",")
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:           brokers,
		Topic:             topic,
		GroupID:           config.GroupId,
		SessionTimeout:    30 * time.Second,
		HeartbeatInterval: 10 * time.Second,
	})
	ctx, cancel := context.WithCancel(context.Background())
	consumer := &KafkaConsumer{
		reader: r,
		wg:     &sync.WaitGroup{},
		logger: logger,
		ctx:    ctx,
		cancel: cancel,
	}
	consumer.wg.Add(1)
	go consumer.readLogs()
	log.Printf("[DEBUG] [Logs-Service] [KafkaConsumer:%s] Successful connect to Kafka-Consumer", consumer.reader.Config().Topic)
	return consumer
}
func (kf *KafkaConsumer) Close() {
	kf.cancel()
	kf.wg.Wait()
	kf.logger.Sync()
	kf.reader.Close()
	log.Printf("[DEBUG] [Logs-Service] [KafkaConsumer:%s] Successful close Kafka-Consumer[%v logs received]", kf.reader.Config().Topic, kf.counter)
}
