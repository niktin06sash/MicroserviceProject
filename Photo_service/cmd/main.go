package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/niktin06sash/MicroserviceProject/Photo_service/internal/brokers/kafka"
	"github.com/niktin06sash/MicroserviceProject/Photo_service/internal/brokers/rabbitmq"
	"github.com/niktin06sash/MicroserviceProject/Photo_service/internal/configs"
	"github.com/niktin06sash/MicroserviceProject/Photo_service/internal/handlers"
	"github.com/niktin06sash/MicroserviceProject/Photo_service/internal/repository"
	"github.com/niktin06sash/MicroserviceProject/Photo_service/internal/server"
	"github.com/niktin06sash/MicroserviceProject/Photo_service/internal/service"
)

func main() {
	config := configs.LoadConfig()
	db, err := repository.NewDatabaseConnection(config.Database)
	if err != nil {
		return
	}
	mega, err := repository.NewMegaClient(config.Mega)
	if err != nil {
		return
	}
	cache, err := repository.NewRedisConnection(config.Redis)
	if err != nil {
		return
	}
	kafkaProducer := kafka.NewKafkaProducer(config.Kafka)
	postgres := repository.NewPhotoPostgresRepo(db)
	redis := repository.NewPhotoRedisRepo(cache)
	service := service.NewPhotoServiceImplement(postgres, mega, redis, kafkaProducer)
	rabbitconsumer, err := rabbitmq.NewRabbitConsumer(config.RabbitMQ, kafkaProducer, service)
	if err != nil {
		return
	}
	api := handlers.NewPhotoAPI(service, kafkaProducer)
	srv := server.NewGrpcServer(config.Server, api)
	kafkaProducer.LogStart()
	serverError := make(chan error, 1)
	go func() {
		if err := srv.Run(config.Server); err != nil {
			serverError <- fmt.Errorf("server run failed: %w", err)
			return
		}
		close(serverError)
	}()
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	select {
	case sig := <-quit:
		log.Printf("[DEBUG] [Photo-Service] Service shutting down with signal: %v", sig)
	case err := <-serverError:
		log.Printf("[DEBUG] [Photo-Service] Service startup failed: %v", err)
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), config.Server.GracefulShutdown)
	defer cancel()
	log.Println("[DEBUG] [Photo-Service] Service is shutting down...")
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("[DEBUG] [Photo-Service] Server shutdown error: %v", err)
		return
	}
	log.Println("[DEBUG] [Photo-Service] Service has shutted down successfully")
	defer func() {
		service.StopWorkers()
		rabbitconsumer.Close()
		db.Close()
		mega.Close()
		kafkaProducer.LogClose()
		kafkaProducer.Close()
		buf := make([]byte, 10<<20)
		n := runtime.Stack(buf, true)
		log.Printf("[DEBUG] [Photo-Service] Active goroutines:\n%s", buf[:n])
	}()
}
