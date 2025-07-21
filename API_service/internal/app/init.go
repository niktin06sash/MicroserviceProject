package app

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/niktin06sash/MicroserviceProject/API_service/internal/brokers/kafka"
	"github.com/niktin06sash/MicroserviceProject/API_service/internal/client"
	"github.com/niktin06sash/MicroserviceProject/API_service/internal/configs"
	"github.com/niktin06sash/MicroserviceProject/API_service/internal/handlers"
	"github.com/niktin06sash/MicroserviceProject/API_service/internal/handlers/middleware"
	"github.com/niktin06sash/MicroserviceProject/API_service/internal/metrics"
	"github.com/niktin06sash/MicroserviceProject/API_service/internal/server"
)

type APIApplication struct {
	config configs.Config
	server *server.Server
}

func NewAPIApplication(config configs.Config) *APIApplication {
	return &APIApplication{config: config}
}

func (a *APIApplication) Start() error {
	defer func() {
		buf := make([]byte, 1<<20)
		n := runtime.Stack(buf, true)
		log.Printf("[DEBUG] [API-Service] Count of active goroutines: %v", runtime.NumGoroutine())
		log.Printf("[DEBUG] [API-Service] Active goroutines:\n%s", buf[:n])
	}()
	metrics.Start()
	defer metrics.Stop()
	sessionclient, err := client.NewGrpcSessionClient(a.config.SessionService)
	if err != nil {
		return err
	}
	defer sessionclient.Close()
	photoclient, err := client.NewGrpcPhotoClient(a.config.PhotoService)
	if err != nil {
		return err
	}
	defer photoclient.Close()
	kafkaprod := kafka.NewKafkaProducer(a.config.Kafka)
	defer kafkaprod.Close()
	defer kafkaprod.LogClose()
	middleware := middleware.NewMiddleware(sessionclient, kafkaprod)
	defer middleware.Stop()
	handler := handlers.NewHandler(middleware, photoclient, kafkaprod, a.config.Routes)
	srv := server.NewServer(a.config.Server, handler.InitRoutes())
	kafkaprod.LogStart()
	serverError := make(chan error, 1)
	go func() {
		if err := srv.Run(); err != nil {
			serverError <- fmt.Errorf("server run failed: %w", err)
			return
		}
		close(serverError)
	}()
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	select {
	case sig := <-quit:
		log.Printf("[DEBUG] [API-Service] Server shutting down with signal: %v", sig)
	case err := <-serverError:
		log.Printf("[DEBUG] [API-Service] Server startup failed: %v", err)
		return err
	}
	return a.Stop()
}
func (a *APIApplication) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), a.config.Server.GracefulShutdown)
	defer cancel()
	log.Println("[DEBUG] [API-Service] Server is shutting down...")
	if err := a.server.Shutdown(ctx); err != nil {
		log.Printf("[DEBUG] [API-Service] Server shutdown error: %v", err)
		return err
	}
	log.Println("[DEBUG] [API-Service] Server has shutted down successfully")
	return nil
}
