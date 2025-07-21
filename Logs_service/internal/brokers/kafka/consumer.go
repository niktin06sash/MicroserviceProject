package kafka

import (
	"context"
	"log"
	"strings"
	"sync"

	configs "github.com/niktin06sash/MicroserviceProject/Logs_service/internal/configs"
	"github.com/segmentio/kafka-go"
)

const (
	LogLevelInfo  = "info"
	LogLevelWarn  = "warn"
	LogLevelError = "error"
)

type Logger interface {
	Log(msg kafka.Message) error
}
type KafkaConsumerGroup struct {
	reader  *kafka.Reader
	wg      *sync.WaitGroup
	logger  Logger
	ctx     context.Context
	cancel  context.CancelFunc
	counter int64
}

func NewKafkaConsumerGroup(config configs.KafkaConfig, logger Logger) *KafkaConsumerGroup {
	brokersString := config.BootstrapServers
	brokers := strings.Split(brokersString, ",")
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:           brokers,
		GroupTopics:       config.GetAllTopics(),
		GroupID:           config.GroupId,
		SessionTimeout:    config.SessionTimeout,
		HeartbeatInterval: config.HearbeatInterval,
	})
	ctx, cancel := context.WithCancel(context.Background())
	consumer := &KafkaConsumerGroup{
		reader: r,
		wg:     &sync.WaitGroup{},
		logger: logger,
		ctx:    ctx,
		cancel: cancel,
	}
	consumer.wg.Add(1)
	go consumer.readLogs()
	log.Println("[DEBUG] [Logs-Service] Successful connect to KafkaConsumer-Group")
	return consumer
}
func (kf *KafkaConsumerGroup) Close() {
	kf.cancel()
	kf.wg.Wait()
	kf.reader.Close()
	log.Printf("[DEBUG] [Logs-Service] Successful close KafkaConsumer-Group[%v logs received]", kf.counter)
}
