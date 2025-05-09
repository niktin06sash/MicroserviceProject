package main

import (
	"log"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"

	configs "github.com/niktin06sash/MicroserviceProject/Kafka_service/internal/configs"
	"github.com/niktin06sash/MicroserviceProject/Kafka_service/internal/kafka"
	"github.com/niktin06sash/MicroserviceProject/Kafka_service/internal/logs"
)

func main() {
	config := configs.LoadConfig()
	topics := config.GetAllTopics()
	consumers := make([]*kafka.KafkaConsumer, 0)
	var wg sync.WaitGroup
	var mux sync.Mutex
	for _, topic := range topics {
		wg.Add(1)
		go func() {
			defer wg.Done()
			logger := logs.NewLogger(config.Logger, topic)
			consumer := kafka.NewKafkaConsumer(config.Kafka, logger, topic)
			mux.Lock()
			consumers = append(consumers, consumer)
			mux.Unlock()
		}()
	}
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	sig := <-quit
	log.Printf("[INFO] [Kafka-Service] Service shutting down with signal: %v", sig)
	log.Println("[INFO] [Kafka-Service] Shutting down Kafka consumers...")
	for _, consumer := range consumers {
		consumer.Close()
	}
	wg.Wait()
	log.Println("[INFO] [Kafka-Service] All Kafka consumers have been successfully closed")
	defer func() {
		buf := make([]byte, 10<<20)
		n := runtime.Stack(buf, true)
		log.Printf("Active goroutines:\n%s", buf[:n])
	}()
}
