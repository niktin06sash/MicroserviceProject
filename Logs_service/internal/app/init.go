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
	"github.com/niktin06sash/MicroserviceProject/Logs_service/internal/storage/elastic"
)

type LogsApplication struct {
	config configs.Config
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
	elastic, err := elastic.NewElasticClient(a.config.Elastic)
	if err != nil {
		return err
	}
	defer elastic.Close()
	logger, err := logs.NewLogger(a.config.Logger, a.config.Kafka.Topics, elastic)
	if err != nil {
		return err
	}
	defer logger.Sync()
	kafkaConsumer := kafka.NewKafkaConsumerGroup(a.config.Kafka, logger)
	defer kafkaConsumer.Close()
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	sig := <-quit
	log.Printf("[DEBUG] [Logs-Service] Server shutting down with signal: %v", sig)
	return nil
}
