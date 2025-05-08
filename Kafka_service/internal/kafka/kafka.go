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
	ctx        context.Context
	cancel     context.CancelFunc
	counter    int64
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
		reader:     r,
		wg:         &sync.WaitGroup{},
		cancelchan: make(chan struct{}),
		logger:     logger,
		ctx:        ctx,
		cancel:     cancel,
	}
	consumer.wg.Add(1)
	go consumer.startLogs()
	log.Printf("[INFO] [Kafka-Service] [KafkaConsumer:%s] Successful connect to Kafka-Consumer", strings.ToUpper(consumer.reader.Config().Topic))
	return consumer
}
func (kf *KafkaConsumer) Close() {
	kf.cancel()
	close(kf.cancelchan)
	kf.wg.Wait()
	kf.reader.Close()
	log.Printf("[INFO] [Kafka-Service] [KafkaConsumer:%s] Successful close Kafka-Consumer[%v logs received]", strings.ToUpper(kf.reader.Config().Topic), kf.counter)
}
func (kf *KafkaConsumer) startLogs() {
	defer kf.wg.Done()
	for {
		select {
		case <-kf.cancelchan:
			return
		case <-kf.ctx.Done():
			return
		default:
			ctx, cancel := context.WithTimeout(kf.ctx, 2*time.Second)
			defer cancel()
			msg, err := kf.reader.FetchMessage(ctx)
			if err != nil {
				if !errors.Is(err, context.Canceled) && !errors.Is(err, context.DeadlineExceeded) {
					log.Printf("[ERROR] [Kafka-Service] [KafkaConsumer:%s] Failed to read log: %v", strings.ToUpper(kf.reader.Config().Topic), err)
				}
				continue
			}
			atomic.AddInt64(&kf.counter, 1)
			switch strings.ToUpper(kf.reader.Config().Topic) {
			case "INFO-LOG-TOPIC":
				kf.logger.ZapLogger.Info(string(msg.Value), zap.Int64("number", kf.counter))
			case "ERROR-LOG-TOPIC":
				kf.logger.ZapLogger.Error(string(msg.Value), zap.Int64("number", kf.counter))
			case "WARN-LOG-TOPIC":
				kf.logger.ZapLogger.Warn(string(msg.Value), zap.Int64("number", kf.counter))
			}
			if err := kf.reader.CommitMessages(ctx, msg); err != nil {
				log.Printf("[ERROR] [Kafka-Service] [KafkaConsumer:%s] Failed to commit offset: %v", strings.ToUpper(kf.reader.Config().Topic), err)
			}
		}
	}
}
