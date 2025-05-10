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

	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/client"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/configs"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/handlers"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/handlers/middleware"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/kafka"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/repository"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/server"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/service"
)

func main() {
	config := configs.LoadConfig()
	db, err := repository.NewDatabaseConnection(config.Database)
	if err != nil {
		return
	}
	kafkaProducer := kafka.NewKafkaProducer(config.Kafka)
	repositories := repository.NewRepository(db)
	grpcclient := client.NewGrpcClient(config.SessionService)
	service := service.NewService(repositories, kafkaProducer, grpcclient)
	middleware := middleware.NewMiddleware(kafkaProducer)
	handlers := handlers.NewHandler(service, middleware, kafkaProducer)
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
		log.Printf("[ERROR] [UserManagement] Service startup failed: %v", err)
		return
	}
	shutdownTimeout := 5 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	log.Println("[INFO] [UserManagement] Service is shutting down...")
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("[ERROR] [UserManagement] Server shutdown error: %v", err)
		return
	}
	log.Println("[INFO] [UserManagement] Service has shutted down successfully")
	defer func() {
		kafkaProducer.Close()
		db.Close()
		grpcclient.Close()
		buf := make([]byte, 10<<20)
		n := runtime.Stack(buf, true)
		log.Printf("Active goroutines:\n%s", buf[:n])
	}()
}
