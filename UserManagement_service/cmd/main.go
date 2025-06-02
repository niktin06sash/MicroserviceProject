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

	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/brokers/kafka"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/brokers/rabbitmq"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/client"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/configs"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/handlers"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/handlers/middleware"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/metrics"
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
	redis, err := repository.NewRedisConnection(config.Redis)
	if err != nil {
		return
	}
	tx := repository.NewTxManagerRepo(db)
	redisdb := repository.NewUserRedisRepo(redis)
	postgredb := repository.NewUserPostgresRepo(db)
	metrics.Start()
	kafkaProducer := kafka.NewKafkaProducer(config.Kafka)
	rabbitproducer, err := rabbitmq.NewRabbitProducer(config.RabbitMQ, kafkaProducer)
	if err != nil {
		return
	}
	grpcclient, err := client.NewGrpcClient(config.SessionService)
	if err != nil {
		return
	}
	service := service.NewUserService(postgredb, tx, redisdb, kafkaProducer, rabbitproducer, grpcclient)
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
		log.Printf("[DEBUG] [User-Service] Service shutting down with signal: %v", sig)
	case err := <-serverError:
		log.Printf("[DEBUG] [User-Service] Service startup failed: %v", err)
		return
	}
	shutdownTimeout := 5 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	log.Println("[DEBUG] [User-Service] Service is shutting down...")
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("[DEBUG] [User-Service] Server shutdown error: %v", err)
		return
	}
	log.Println("[DEBUG] [User-Service] Service has shutted down successfully")
	defer func() {
		metrics.Stop()
		kafkaProducer.Close()
		db.Close()
		grpcclient.Close()
		buf := make([]byte, 10<<20)
		n := runtime.Stack(buf, true)
		log.Printf("[DEBUG] [User-Service] Active goroutines:\n%s", buf[:n])
	}()
}
