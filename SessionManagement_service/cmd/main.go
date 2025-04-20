package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"
	"time"

	"github.com/niktin06sash/MicroserviceProject/SessionManagement_service/internal/configs"
	"github.com/niktin06sash/MicroserviceProject/SessionManagement_service/internal/logger"
	"github.com/niktin06sash/MicroserviceProject/SessionManagement_service/internal/repository"
	"github.com/niktin06sash/MicroserviceProject/SessionManagement_service/internal/server"
	"github.com/niktin06sash/MicroserviceProject/SessionManagement_service/internal/service"
	"go.uber.org/zap"

	"github.com/spf13/viper"
)

func main() {
	logger := logger.NewSessionLogger()
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		logger.Fatal("SessionManagement: Failed to get current file path")
	}

	cmdDir := filepath.Dir(filename)

	projectRoot := filepath.Dir(filepath.Dir(cmdDir))

	configDir := filepath.Join(projectRoot, "SessionManagement_service/internal/configs")

	viper.SetConfigName("config")
	viper.SetConfigType("yml")
	viper.AddConfigPath(configDir)

	err := viper.ReadInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			logger.Warn("SessionManagement: Config file not found; using defaults or environment variables")
		} else {
			logger.Fatal("SessionManagement: Failed to get current file path", zap.Error(err))
		}
	}
	var config configs.Config

	err = viper.Unmarshal(&config)
	if err != nil {
		logger.Fatal("SessionManagement: Failed to get current file path", zap.Error(err))
	}

	redisobject := &repository.RedisObject{
		Logger: logger,
	}
	redis, err := repository.ConnectToRedis(config, redisobject)
	if err != nil {
		logger.Fatal("SessionManagement: Failed to connect to database", zap.Error(err))
		return
	}
	defer redisobject.Close(redis)
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

	port := viper.GetString("server.port")
	if port == "" {
		port = "50051"
	}
	serverError := make(chan error, 1)
	go func() {
		if err := srv.Run(port); err != nil {
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
}
