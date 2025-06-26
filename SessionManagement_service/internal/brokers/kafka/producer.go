package kafka

import (
	"context"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/niktin06sash/MicroserviceProject/SessionManagement_service/internal/configs"
	"github.com/segmentio/kafka-go"
)

const (
	LogLevelInfo    = "info"
	LogLevelWarn    = "warn"
	LogLevelError   = "error"
	logCloseService = "----------CLOSE SERVICE----------"
	logStartService = "----------START SERVICE----------"
)

type KafkaProducer struct {
	writer  *kafka.Writer
	logchan chan SessionLog
	wg      *sync.WaitGroup
	context context.Context
	cancel  context.CancelFunc
	topics  map[string]string
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
	logs := make(chan SessionLog, 1000)
	producer := &KafkaProducer{
		writer:  w,
		logchan: logs,
		wg:      &sync.WaitGroup{},
		context: ctx,
		cancel:  cancel,
		topics:  config.Topics,
	}
	for i := 1; i <= 3; i++ {
		producer.wg.Add(1)
		go producer.sendLogs(i)
	}
	log.Println("[DEBUG] [Session-Service] Successful connect to Kafka-Producer")
	return producer
}
func (kf *KafkaProducer) Close() {
	close(kf.logchan)
	kf.cancel()
	kf.wg.Wait()
	kf.writer.Close()
	log.Println("[DEBUG] [Session-Service] Successful close Kafka-Producer")
}
