package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"
	"time"

	"github.com/niktin06sash/MicroserviceProject/SessionManagement_service/internal/configs"
	"github.com/niktin06sash/MicroserviceProject/SessionManagement_service/internal/repository"
	"github.com/niktin06sash/MicroserviceProject/SessionManagement_service/internal/server"
	"github.com/niktin06sash/MicroserviceProject/SessionManagement_service/internal/service"

	"github.com/spf13/viper"
)

func main() {

	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		log.Fatalf("Failed to get current file path")
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

			log.Printf("Config file not found; using defaults or environment variables")
		} else {
			log.Fatalf("Error reading config file: %s", err)
		}
	}
	var config configs.Config

	err = viper.Unmarshal(&config)
	if err != nil {
		log.Fatalf("Unable to decode into struct, %v", err)
	}

	redis, redisInterface, err := repository.ConnectToRedis(config)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
		return
	}
	defer redisInterface.Close(redis)
	/*brokersString := config.Kafka.BootstrapServers
	brokers := strings.Split(brokersString, ",")
	kafkaProducer, err := kafka.NewKafkaProducer(brokers)
	if err != nil {
		log.Fatalf("Failed to create Kafka producer: %v", err)
		return
	}
	defer kafkaProducer.Close()*/
	repository := repository.NewRepository(redis)

	service := service.NewService(repository)
	srv := server.NewGrpcServer(service)

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
		log.Printf("Service shutting down with signal: %v", sig)
	case err := <-serverError:
		log.Fatalf("Service startup failed: %v", err)
	}

	shutdownTimeout := 5 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	log.Println("Service is shutting down...")
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}

	log.Println("Service has shut down successfully")
}
