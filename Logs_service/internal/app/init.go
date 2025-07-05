package app

import (
	"log"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/niktin06sash/MicroserviceProject/Logs_service/internal/brokers/kafka"
	"github.com/niktin06sash/MicroserviceProject/Logs_service/internal/configs"
	"github.com/niktin06sash/MicroserviceProject/Logs_service/internal/logs"
)

type LogsApplication struct {
	config        configs.Config
	kafkaConsumer *kafka.KafkaConsumerGroup
	logger        *logs.Logger
}

func NewLogsApplication(config configs.Config) *LogsApplication {
	return &LogsApplication{config: config}
}
func (a *LogsApplication) Start() error {
	defer func() {
		buf := make([]byte, 1<<20)
		n := runtime.Stack(buf, true)
		log.Printf("[DEBUG] [Logs-Service] Count of active goroutines: %v", runtime.NumGoroutine())
		log.Printf("[DEBUG] [Logs-Service] Active goroutines:\n%s", buf[:n])
	}()
	var err error
	a.logger, err = logs.NewLogger(a.config.Logger, a.config.Kafka.Topics)
	if err != nil {
		return err
	}
	defer a.logger.Sync()
	a.kafkaConsumer = kafka.NewKafkaConsumerGroup(a.config.Kafka, a.logger)
	defer a.kafkaConsumer.Close()
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	sig := <-quit
	log.Printf("[DEBUG] [Logs-Service] Server shutting down with signal: %v", sig)
	return nil
}
