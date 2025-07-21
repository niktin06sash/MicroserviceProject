package app

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime"
	"syscall"

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

type UserApplication struct {
	config configs.Config
	server *server.Server
}

func NewUserApplication(config configs.Config) *UserApplication {
	return &UserApplication{config: config}
}

func (a *UserApplication) Start() error {
	defer func() {
		buf := make([]byte, 1<<20)
		n := runtime.Stack(buf, true)
		log.Printf("[DEBUG] [User-Service] Count of active goroutines: %v", runtime.NumGoroutine())
		log.Printf("[DEBUG] [User-Service] Active goroutines:\n%s", buf[:n])
	}()
	db, err := repository.NewDatabaseConnection(a.config.Database)
	if err != nil {
		return err
	}
	defer db.Close()
	redis, err := repository.NewRedisConnection(a.config.Redis)
	if err != nil {
		return err
	}
	defer redis.Close()
	tx := repository.NewTxManagerRepo(db)
	redisdb := repository.NewUserRedisRepo(redis)
	postgredb := repository.NewUserPostgresRepo(db)
	metrics.Start()
	defer metrics.Stop()
	kafkaProducer := kafka.NewKafkaProducer(a.config.Kafka)
	defer kafkaProducer.Close()
	defer kafkaProducer.LogClose()
	rabbitproducer, err := rabbitmq.NewRabbitProducer(a.config.RabbitMQ, kafkaProducer)
	if err != nil {
		return err
	}
	defer rabbitproducer.Close()
	grpcclient, err := client.NewGrpcClient(a.config.SessionService)
	if err != nil {
		return err
	}
	defer grpcclient.Close()
	service := service.NewUserService(postgredb, tx, redisdb, kafkaProducer, rabbitproducer, grpcclient)
	middleware := middleware.NewMiddleware(kafkaProducer)
	handlers := handlers.NewHandler(service, middleware, kafkaProducer)
	a.server = server.NewServer(a.config.Server, handlers.InitRoutes())
	kafkaProducer.LogStart()
	serverError := make(chan error, 1)
	go func() {
		if err := a.server.Run(); err != nil {
			serverError <- fmt.Errorf("server run failed: %w", err)
			return
		}
		close(serverError)
	}()
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	select {
	case sig := <-quit:
		log.Printf("[DEBUG] [User-Service] Server shutting down with signal: %v", sig)
	case err := <-serverError:
		log.Printf("[DEBUG] [User-Service] Server startup failed: %v", err)
		return err
	}
	return a.Stop()
}
func (a *UserApplication) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), a.config.Server.GracefulShutdown)
	defer cancel()
	log.Println("[DEBUG] [User-Service] Server is shutting down...")
	if err := a.server.Shutdown(ctx); err != nil {
		log.Printf("[DEBUG] [User-Service] Server shutdown error: %v", err)
		return err
	}
	log.Println("[DEBUG] [User-Service] Server has shutted down successfully")
	return nil
}
