package main

import (
	"log"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/niktin06sash/MicroserviceProject/Logs_service/internal/brokers/kafka"
	configs "github.com/niktin06sash/MicroserviceProject/Logs_service/internal/configs"
	"github.com/niktin06sash/MicroserviceProject/Logs_service/internal/logs"
)

func main() {
	config := configs.LoadConfig()
	logger, err := logs.NewLogger(config.Logger, config.Kafka.Topics)
	if err != nil {
		return
	}
	groupconsumer := kafka.NewKafkaConsumerGroup(config.Kafka, logger)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	sig := <-quit
	log.Printf("[DEBUG] [Logs-Service] Service shutting down with signal: %v", sig)
	log.Println("[DEBUG] [Logs-Service] Shutting down Kafka consumers...")
	defer func() {
		groupconsumer.Close()
		logger.Sync()
		buf := make([]byte, 10<<20)
		n := runtime.Stack(buf, true)
		log.Printf("[DEBUG] [Logs-Service] Active goroutines:\n%s", buf[:n])
	}()
}
