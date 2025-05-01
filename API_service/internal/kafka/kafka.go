package kafka

import (
	"log"
	"strings"
	"time"

	"github.com/niktin06sash/MicroserviceProject/API_service/internal/configs"
	"github.com/segmentio/kafka-go"
)

type KafkaProducer struct {
	writer *kafka.Writer
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
	log.Println("[INFO] [API-Service] Successful connect to Kafka-Producer")
	return &KafkaProducer{writer: w}
}

func (kf *KafkaProducer) Close() {
	kf.writer.Close()
	log.Println("[INFO] [API-Service] Successful close Kafka-Producer")
}
