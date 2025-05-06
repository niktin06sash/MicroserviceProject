package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/client"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/configs"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/handlers"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/kafka"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/repository"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/server"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/service"
)

func main() {
	config := configs.LoadConfig()
	db, err := repository.NewDatabaseConnection(config.Database)
	if err != nil {
		log.Fatalf("[ERROR] [UserManagement] Failed to connect to database: %v", err)
		return
	}
	brokersString := config.Kafka.BootstrapServers
	brokers := strings.Split(brokersString, ",")
	kafkaProducer, err := kafka.NewKafkaProducer(brokers)
	if err != nil {
		log.Fatalf("[ERROR] [UserManagement] Failed to create Kafka producer: %v", err)
		return
	}
	defer kafkaProducer.Close()
	repositories := repository.NewRepository(db)
	grpcclient := client.NewGrpcClient(config.SessionService)
	service := service.NewService(repositories, kafkaProducer, grpcclient)
	handlers := handlers.NewHandler(service)
	srv := &server.Server{}
	serverError := make(chan error, 1)
	go func() {
		if err := srv.Run(config.Server.Port, handlers.InitRoutes()); err != nil {
			serverError <- fmt.Errorf("server run failed: %w", err)
			return
		}
		close(serverError)
	}()
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	select {
	case sig := <-quit:
		log.Printf("[INFO] [UserManagement] Service shutting down with signal: %v", sig)
	case err := <-serverError:
		log.Fatalf("[ERROR] [UserManagement] Service startup failed: %v", err)
	}
	shutdownTimeout := 5 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	log.Println("[INFO] [UserManagement] Service is shutting down...")
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("[ERROR] [UserManagement] Server shutdown error: %v", err)
	}
	log.Println("[INFO] [UserManagement] Service has shutted down successfully")
	defer func() {
		db.Close()
		grpcclient.Close()
		buf := make([]byte, 10<<20)
		n := runtime.Stack(buf, true)
		log.Printf("Active goroutines:\n%s", buf[:n])
	}()
}
