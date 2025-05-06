package main

import (
	"log"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	configs "github.com/niktin06sash/MicroserviceProject/Kafka_service/internal/config"
	"github.com/niktin06sash/MicroserviceProject/Kafka_service/internal/kafka"
)

func main() {
	config := configs.LoadConfig()
	topicslice := []string{config.Kafka.Topics.InfoLog, config.Kafka.Topics.ErrorLog, config.Kafka.Topics.WarnLog}
	consumers := make([]*kafka.KafkaConsumer, 0)
	for _, topic := range topicslice {
		consumer := kafka.NewKafkaConsumer(config.Kafka, topic)
		consumers = append(consumers, consumer)
	}
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	sig := <-quit
	log.Printf("[INFO] [Kafka-Service] Service shutting down with signal: %v", sig)
	log.Println("[INFO] [Kafka-Service] Shutting down Kafka consumers...")
	for _, consumer := range consumers {
		consumer.Close()
	}
	defer func() {
		buf := make([]byte, 10<<20)
		n := runtime.Stack(buf, true)
		log.Printf("Active goroutines:\n%s", buf[:n])
	}()
}
