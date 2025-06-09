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

	"github.com/niktin06sash/MicroserviceProject/API_service/internal/brokers/kafka"
	"github.com/niktin06sash/MicroserviceProject/API_service/internal/client"
	"github.com/niktin06sash/MicroserviceProject/API_service/internal/configs"
	"github.com/niktin06sash/MicroserviceProject/API_service/internal/handlers"
	"github.com/niktin06sash/MicroserviceProject/API_service/internal/handlers/middleware"
	_ "github.com/niktin06sash/MicroserviceProject/API_service/internal/handlers/middleware"
	_ "github.com/niktin06sash/MicroserviceProject/API_service/internal/handlers/response"
	"github.com/niktin06sash/MicroserviceProject/API_service/internal/metrics"
	"github.com/niktin06sash/MicroserviceProject/API_service/internal/server"
)

// @title API-Gateway
// @version 1.0
// @description This is a sample server for managing users and sessions.
// @host localhost:8083
// @BasePath /
// @schemes http

func main() {
	config := configs.LoadConfig()
	metrics.Start()
	kafkaprod := kafka.NewKafkaProducer(config.Kafka)
	grpcclient, err := client.NewGrpcClient(config.SessionService)
	if err != nil {
		return
	}
	middleware := middleware.NewMiddleware(grpcclient, kafkaprod)
	handler := handlers.NewHandler(middleware, kafkaprod, config.Routes)
	srv := &server.Server{}
	serverError := make(chan error, 1)
	go func() {
		if err := srv.Run(config.Server.Port, handler.InitRoutes()); err != nil {
			serverError <- fmt.Errorf("server run failed: %w", err)
			return
		}
		close(serverError)
	}()
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	select {
	case sig := <-quit:
		log.Printf("[DEBUG] [API-Service] Service shutting down with signal: %v", sig)
	case err := <-serverError:
		log.Printf("[DEBUG] [API-Service] Service startup failed: %v", err)
		return
	}
	shutdownTimeout := 5 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	log.Println("[DEBUG] [API-Service] Service is shutting down...")
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("[DEBUG] [API-Service] Server shutdown error: %v", err)
		return
	}
	log.Println("[DEBUG] [API-Service] Service has shutted down successfully")
	defer func() {
		metrics.Stop()
		middleware.Stop()
		grpcclient.Close()
		kafkaprod.Close()
		buf := make([]byte, 10<<20)
		n := runtime.Stack(buf, true)
		log.Printf("[DEBUG] [API-Service] Active goroutines:\n%s", buf[:n])
	}()
}
