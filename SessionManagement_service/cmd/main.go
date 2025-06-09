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

	"github.com/niktin06sash/MicroserviceProject/SessionManagement_service/internal/brokers/kafka"
	"github.com/niktin06sash/MicroserviceProject/SessionManagement_service/internal/configs"
	"github.com/niktin06sash/MicroserviceProject/SessionManagement_service/internal/handlers"
	"github.com/niktin06sash/MicroserviceProject/SessionManagement_service/internal/repository"
	"github.com/niktin06sash/MicroserviceProject/SessionManagement_service/internal/server"
	"github.com/niktin06sash/MicroserviceProject/SessionManagement_service/internal/service"
)

func main() {
	config := configs.LoadConfig()
	redis, err := repository.NewRedisConnection(config.Redis)
	if err != nil {
		return
	}
	kafkaProducer := kafka.NewKafkaProducer(config.Kafka)
	repository := repository.NewSessionRepos(redis)
	service := service.NewSessionService(repository, kafkaProducer)
	api := handlers.NewSessionAPI(service, kafkaProducer)
	srv := server.NewGrpcServer(api)
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
		log.Printf("[DEBUG] [Session-Service] Service shutting down with signal: %v", sig)
	case err := <-serverError:
		log.Printf("[DEBUG] [Session-Service] Service startup failed: %v", err)
		return
	}
	shutdownTimeout := 5 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	log.Println("[DEBUG] [Session-Service] Service is shutting down...")
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("[DEBUG] [Session-Service] Server shutdown error: %v", err)
		return
	}
	log.Println("[DEBUG] [Session-Service] Service has shutted down successfully")
	defer func() {
		redis.Close()
		kafkaProducer.Close()
		buf := make([]byte, 10<<20)
		n := runtime.Stack(buf, true)
		log.Printf("[DEBUG] [Session-Service] Active goroutines:\n%s", buf[:n])
	}()
}
