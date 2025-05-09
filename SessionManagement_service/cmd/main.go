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
	"github.com/niktin06sash/MicroserviceProject/SessionManagement_service/internal/kafka"
	"github.com/niktin06sash/MicroserviceProject/SessionManagement_service/internal/repository"
	"github.com/niktin06sash/MicroserviceProject/SessionManagement_service/internal/server"
	"github.com/niktin06sash/MicroserviceProject/SessionManagement_service/internal/service"
)

func main() {
	config := configs.LoadConfig()
	kafkaProducer := kafka.NewKafkaProducer(config.Kafka)
	redis, err := repository.NewRedisConnection(config.Redis)
	if err != nil {
		log.Fatalf("[ERROR] [Session-Service] Failed to connect to database: %v", err)
		return
	}
	repository := repository.NewRepository(redis)
	service := service.NewSessionAPI(repository)
	srv := server.NewGrpcServer(service)
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
		log.Printf("[INFO] [Session-Service] Service shutting down with signal: %v", sig)
	case err := <-serverError:
		log.Fatalf("[ERROR] [Session-Service] Service startup failed: %v", err)
	}
	shutdownTimeout := 5 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	log.Println("[INFO] [Session-Service] Service is shutting down...")
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("[ERROR] [Session-Service] Server shutdown error: %v", err)
	}
	log.Println("[INFO] [Session-Service] Service has shutted down successfully")
	defer func() {
		redis.Close()
		kafkaProducer.Close()
		buf := make([]byte, 10<<20)
		n := runtime.Stack(buf, true)
		log.Printf("Active goroutines:\n%s", buf[:n])
	}()
}
