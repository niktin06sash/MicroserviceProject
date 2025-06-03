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

	"github.com/niktin06sash/MicroserviceProject/Photo_service/internal/brokers/kafka"
	"github.com/niktin06sash/MicroserviceProject/Photo_service/internal/brokers/rabbitmq"
	"github.com/niktin06sash/MicroserviceProject/Photo_service/internal/configs"
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
	kafkaProducer := kafka.NewKafkaProducer(config.Kafka)
	postgres := repository.NewPhotoPostgresRepo(db)
	rabbitconsumer, err := rabbitmq.NewRabbitConsumer(config.RabbitMQ, kafkaProducer, postgres)
	if err != nil {
		return
	}
	api := service.NewPhotoAPI(postgres, mega, kafkaProducer)
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
		log.Printf("[DEBUG] [Photo-Service] Service shutting down with signal: %v", sig)
	case err := <-serverError:
		log.Printf("[DEBUG] [Photo-Service] Service startup failed: %v", err)
		return
	}
	shutdownTimeout := 5 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	log.Println("[DEBUG] [Photo-Service] Service is shutting down...")
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("[DEBUG] [Photo-Service] Server shutdown error: %v", err)
		return
	}
	log.Println("[DEBUG] [Photo-Service] Service has shutted down successfully")
	defer func() {
		db.Close()
		rabbitconsumer.Close()
		kafkaProducer.Close()
		buf := make([]byte, 10<<20)
		n := runtime.Stack(buf, true)
		log.Printf("[DEBUG] [Photo-Service] Active goroutines:\n%s", buf[:n])
	}()
}
