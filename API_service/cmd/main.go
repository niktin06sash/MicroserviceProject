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

	"github.com/niktin06sash/MicroserviceProject/API_service/internal/client"
	"github.com/niktin06sash/MicroserviceProject/API_service/internal/configs"
	"github.com/niktin06sash/MicroserviceProject/API_service/internal/handlers"
	"github.com/niktin06sash/MicroserviceProject/API_service/internal/handlers/middleware"
	_ "github.com/niktin06sash/MicroserviceProject/API_service/internal/handlers/middleware"
	_ "github.com/niktin06sash/MicroserviceProject/API_service/internal/handlers/response"
	"github.com/niktin06sash/MicroserviceProject/API_service/internal/kafka"
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
	kafkaprod := kafka.NewKafkaProducer(config.Kafka)
	grpcclient := client.NewGrpcClient(config.SessionService)
	middleware := middleware.NewMiddleware(grpcclient, kafkaprod)
	handler := handlers.NewHandler(middleware, kafkaprod, config.Routes)
	srv := &server.Server{}
	port := config.Server.Port
	if port == "" {
		port = "8083"
	}
	serverError := make(chan error, 1)
	go func() {
		if err := srv.Run(port, handler.InitRoutes()); err != nil {
			serverError <- fmt.Errorf("server run failed: %w", err)
			return
		}
		close(serverError)
	}()
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	select {
	case sig := <-quit:
		log.Printf("[INFO] [API-Service] Service shutting down with signal: %v", sig)
	case err := <-serverError:
		log.Fatalf("[ERROR] [API-Service] Service startup failed: %v", err)
	}
	shutdownTimeout := 5 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	log.Println("[INFO] [API-Service] Service is shutting down...")
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("[INFO] [API-Service] Server shutdown error: %v", err)
	}
	log.Println("[INFO] [API-Service] Service has shutted down successfully")
	defer func() {
		middleware.Stop()
		grpcclient.Close()
		kafkaprod.Close()
		buf := make([]byte, 10<<20)
		n := runtime.Stack(buf, true)
		log.Printf("Active goroutines:\n%s", buf[:n])
	}()
}
