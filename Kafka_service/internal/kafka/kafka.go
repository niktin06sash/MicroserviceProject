package kafka

import (
	"context"
	"log"
	"strings"
	"sync"

	configs "github.com/niktin06sash/MicroserviceProject/Kafka_service/internal/configs"
	"github.com/segmentio/kafka-go"
)

type KafkaConsumer struct {
	reader *kafka.Reader
	wg     *sync.WaitGroup
}

func NewKafkaConsumer(config configs.KafkaConfig, topic string) *KafkaConsumer {
	brokersString := config.BootstrapServers
	brokers := strings.Split(brokersString, ",")
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:           brokers,
		Topic:             topic,
		GroupID:           config.GroupId,
		SessionTimeout:    config.SessionTimeout,
		HeartbeatInterval: config.HearbeatInterval,
	})
	producer := &KafkaConsumer{
		reader: r,
		wg:     &sync.WaitGroup{},
	}
	producer.wg.Add(1)
	go func() {
		producer.startLogs()
	}()
	log.Printf("[INFO] [Kafka-Service] [KafkaConsumer:%s] Successful connect to Kafka-Consumer", topic)
	return producer
}
func (kf *KafkaConsumer) Close() {
	kf.wg.Wait()
	kf.reader.Close()
	log.Printf("[INFO] [Kafka-Service] [KafkaConsumer:%s] Successful close Kafka-Consumer", strings.ToUpper(kf.reader.Config().Topic))
}
func (kf *KafkaConsumer) startLogs() {
	defer kf.wg.Done()
	for {
		msg, err := kf.reader.ReadMessage(context.Background())
		if err != nil {
			log.Printf("[ERROR] [Kafka-Service] [KafkaConsumer:%s] Failed to read log: %v", strings.ToUpper(kf.reader.Config().Topic), err)
		}
		log.Printf("[INFO] [Kafka-Service] [KafkaConsumer:%s] Received log: %s", msg.Topic, string(msg.Value))
		if err := kf.reader.CommitMessages(context.Background(), msg); err != nil {
			log.Printf("[ERROR] [Kafka-Service] [KafkaConsumer:%s] Failed to commit log: %v", msg.Topic, err)
		}
	}
}
