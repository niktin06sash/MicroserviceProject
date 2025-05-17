package kafka

import (
	"context"
	"errors"
	"log"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	configs "github.com/niktin06sash/MicroserviceProject/Logs_service/internal/configs"
	"github.com/niktin06sash/MicroserviceProject/Logs_service/internal/logs"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
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
	go consumer.startLogs()
	log.Printf("[INFO] [Logs-Service] [KafkaConsumer:%s] Successful connect to Kafka-Consumer", consumer.reader.Config().Topic)
	return consumer
}
func (kf *KafkaConsumer) Close() {
	kf.cancel()
	kf.wg.Wait()
	kf.logger.Sync()
	kf.reader.Close()
	log.Printf("[INFO] [Logs-Service] [KafkaConsumer:%s] Successful close Kafka-Consumer[%v logs received]", kf.reader.Config().Topic, kf.counter)
}
func (kf *KafkaConsumer) startLogs() {
	defer kf.wg.Done()
	for {
		select {
		case <-kf.ctx.Done():
			return
		default:
			ctx, cancel := context.WithTimeout(kf.ctx, 2*time.Second)
			defer cancel()
			msg, err := kf.reader.FetchMessage(ctx)
			if err != nil {
				if !errors.Is(err, context.Canceled) && !errors.Is(err, context.DeadlineExceeded) {
					log.Printf("[ERROR] [Logs-Service] [KafkaConsumer:%s] Failed to read log: %v", kf.reader.Config().Topic, err)
				}
				continue
			}
			atomic.AddInt64(&kf.counter, 1)
			parts := strings.Split(kf.reader.Config().Topic, "-")
			level := parts[1]
			switch level {
			case "info":
				kf.logger.ZapLogger.Info(string(msg.Value), zap.Int64("number", kf.counter))
			case "error":
				kf.logger.ZapLogger.Error(string(msg.Value), zap.Int64("number", kf.counter))
			case "warn":
				kf.logger.ZapLogger.Warn(string(msg.Value), zap.Int64("number", kf.counter))
			}
			if err := kf.reader.CommitMessages(ctx, msg); err != nil {
				log.Printf("[ERROR] [Logs-Service] [KafkaConsumer:%s] Failed to commit offset: %v", kf.reader.Config().Topic, err)
			}
		}
	}
}
