package kafka

import (
	"context"
	"errors"
	"log"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	configs "github.com/niktin06sash/MicroserviceProject/Kafka_service/internal/configs"
	"github.com/niktin06sash/MicroserviceProject/Kafka_service/internal/logs"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

type KafkaConsumer struct {
	reader     *kafka.Reader
	wg         *sync.WaitGroup
	cancelchan chan struct{}
	logger     *logs.Logger
	counter    int64
}

func NewKafkaConsumer(config configs.KafkaConfig, logger *logs.Logger, topic string) *KafkaConsumer {
	brokersString := config.BootstrapServers
	brokers := strings.Split(brokersString, ",")
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:           brokers,
		Topic:             topic,
		GroupID:           config.GroupId,
		SessionTimeout:    config.SessionTimeout,
		HeartbeatInterval: config.HearbeatInterval,
	})
	consumer := &KafkaConsumer{
		reader:     r,
		wg:         &sync.WaitGroup{},
		cancelchan: make(chan struct{}),
		logger:     logger,
	}
	consumer.wg.Add(1)
	go func() {
		defer consumer.wg.Done()
		consumer.startLogs()
	}()
	log.Printf("[INFO] [Kafka-Service] [KafkaConsumer:%s] Successful connect to Kafka-Consumer", strings.ToUpper(consumer.reader.Config().Topic))
	consumer.logger.ZapLogger.Info("Start services...", zap.Time("start_time", time.Now()))
	return consumer
}
func (kf *KafkaConsumer) Close() {
	close(kf.cancelchan)
	kf.wg.Wait()
	kf.logger.Sync()
	kf.reader.Close()
	log.Printf("[INFO] [Kafka-Service] [KafkaConsumer:%s] Successful close Kafka-Consumer[%v logs received]", strings.ToUpper(kf.reader.Config().Topic), kf.counter)
}
func (kf *KafkaConsumer) startLogs() {
	for {
		select {
		case <-kf.cancelchan:
			return
		default:
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer func() {
				cancel()
			}()
			msg, err := kf.reader.ReadMessage(ctx)
			if err != nil {
				if !errors.Is(err, context.Canceled) && !errors.Is(err, context.DeadlineExceeded) {
					log.Printf("[ERROR] [Kafka-Service] [KafkaConsumer:%s] Failed to read log: %v", strings.ToUpper(kf.reader.Config().Topic), err)
					continue
				}
			}
			atomic.AddInt64(&kf.counter, 1)
			kf.logger.ZapLogger.Info(string(msg.Value), zap.Int64("number", kf.counter))
			if err := kf.reader.CommitMessages(ctx, msg); err != nil {
				log.Printf("[ERROR] [Kafka-Service] [KafkaConsumer:%s] Failed to commit log: %v", strings.ToUpper(kf.reader.Config().Topic), err)
			}
		}
	}
}
