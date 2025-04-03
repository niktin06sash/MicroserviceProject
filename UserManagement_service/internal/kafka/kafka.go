package kafka

import (
	"context"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
)

type KafkaProducer interface {
	SendMessage(topic string, key string, value []byte) error
	Close() error
}

type kafkaProducer struct {
	writer *kafka.Writer
}

func NewKafkaProducer(brokers []string) (KafkaProducer, error) {
	w := &kafka.Writer{
		Addr:         kafka.TCP(brokers...), // Pass brokers as a slice of strings
		Topic:        "",                    // Topic is specified in SendMessage
		Balancer:     &kafka.LeastBytes{},   // Use LeastBytes balancer for even distribution
		WriteTimeout: 10 * time.Second,      // Set a write timeout
	}

	return &kafkaProducer{writer: w}, nil
}

func (kp *kafkaProducer) SendMessage(topic string, key string, value []byte) error {
	msg := kafka.Message{
		Topic: topic,
		Key:   []byte(key),
		Value: value,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second) // Add context
	defer cancel()

	err := kp.writer.WriteMessages(ctx, msg) // Use WriteMessages with context
	if err != nil {
		return fmt.Errorf("failed to write messages: %w", err)
	}

	return nil
}

func (kp *kafkaProducer) Close() error {
	err := kp.writer.Close()
	if err != nil {
		return fmt.Errorf("failed to close writer: %w", err)
	}
	return nil
}
