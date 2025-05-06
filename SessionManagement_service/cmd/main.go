package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/niktin06sash/MicroserviceProject/SessionManagement_service/internal/configs"
	"github.com/niktin06sash/MicroserviceProject/SessionManagement_service/internal/logger"
	"github.com/niktin06sash/MicroserviceProject/SessionManagement_service/internal/repository"
	"github.com/niktin06sash/MicroserviceProject/SessionManagement_service/internal/server"
	"github.com/niktin06sash/MicroserviceProject/SessionManagement_service/internal/service"
	"go.uber.org/zap"
)

func main() {
	logger := logger.NewSessionLogger()
	config := configs.LoadConfig()
	redis, err := repository.NewRedisConnection(config.Redis, logger)
	if err != nil {
		logger.Fatal("SessionManagement: Failed to connect to database", zap.Error(err))
		return
	}
	/*brokersString := config.Kafka.BootstrapServers
	brokers := strings.Split(brokersString, ",")
	kafkaProducer, err := kafka.NewKafkaProducer(brokers)
	if err != nil {
		log.Fatalf("Failed to create Kafka producer: %v", err)
		return
	}
	defer kafkaProducer.Close()*/
	repository := repository.NewRepository(redis, logger)
	service := service.NewSessionAPI(repository, logger)
	srv := server.NewGrpcServer(service, logger)
	serverError := make(chan error, 1)
	go func() {
		if err := srv.Run(config.Server.Port); err != nil {
			serverError <- fmt.Errorf("server run failed: %w", err)
			return
		}
		close(serverError)
	}()
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	select {
	case sig := <-quit:
		logger.Info("SessionManagement: Service shutting down with signal", zap.String("signal", sig.String()))
	case err := <-serverError:
		logger.Fatal("SessionManagement: Service startup failed", zap.Error(err))
	}
	shutdownTimeout := 5 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	logger.Info("SessionManagement: Service is shutting down...")
	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("SessionManagement: Server shutdown error", zap.Error(err))
	}
	logger.Info("SessionManagement: Service has shut down successfully")
	defer func() {
		redis.Close()
		buf := make([]byte, 10<<20)
		n := runtime.Stack(buf, true)
		log.Printf("Active goroutines:\n%s", buf[:n])
	}()
}
