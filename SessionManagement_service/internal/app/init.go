package app

import (
	"context"
	"log"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/niktin06sash/MicroserviceProject/SessionManagement_service/internal/brokers/kafka"
	"github.com/niktin06sash/MicroserviceProject/SessionManagement_service/internal/configs"
	"github.com/niktin06sash/MicroserviceProject/SessionManagement_service/internal/handlers"
	"github.com/niktin06sash/MicroserviceProject/SessionManagement_service/internal/repository"
	"github.com/niktin06sash/MicroserviceProject/SessionManagement_service/internal/server"
	"github.com/niktin06sash/MicroserviceProject/SessionManagement_service/internal/service"
)

type SessionApplication struct {
	config configs.Config
	server *server.GrpcServer
}

func NewSessionApplication(config configs.Config) *SessionApplication {
	return &SessionApplication{config: config}
}
func (a *SessionApplication) Start() error {
	defer func() {
		buf := make([]byte, 1<<20)
		n := runtime.Stack(buf, true)
		log.Printf("[DEBUG] [Session-Service] Count of active goroutines: %v", runtime.NumGoroutine())
		log.Printf("[DEBUG] [Session-Service] Active goroutines:\n%s", buf[:n])
	}()
	cache, err := repository.NewRedisConnection(a.config.Redis)
	if err != nil {
		return err
	}
	defer cache.Close()
	kafkaProducer := kafka.NewKafkaProducer(a.config.Kafka)
	defer kafkaProducer.Close()
	defer kafkaProducer.LogClose()
	repository := repository.NewSessionRepos(cache)
	service := service.NewSessionService(repository, kafkaProducer)
	api := handlers.NewSessionAPI(service, kafkaProducer)
	a.server = server.NewGrpcServer(a.config.Server, api)
	kafkaProducer.LogStart()
	serverError := make(chan error, 1)
	go func() {
		if err := a.server.Run(a.config.Server); err != nil {
			serverError <- err
		}
		close(serverError)
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	select {
	case sig := <-quit:
		log.Printf("[DEBUG] [Session-Service] Server shutting down with signal: %v", sig)
	case err := <-serverError:
		log.Printf("[DEBUG] [Session-Service] Server startup failed: %v", err)
		return err
	}

	return a.Stop()
}
func (a *SessionApplication) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), a.config.Server.GracefulShutdown)
	defer cancel()
	log.Println("Server is shutting down...")
	if err := a.server.Shutdown(ctx); err != nil {
		log.Printf("[DEBUG] [Session-Service] Server shutdown error: %v", err)
		return err
	}
	log.Println("[DEBUG] [Session-Service] Server has shutted down successfully")
	return nil
}
