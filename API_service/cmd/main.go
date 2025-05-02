package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
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
	"github.com/spf13/viper"
)

// @title API-Gateway
// @version 1.0
// @description This is a sample server for managing users and sessions.
// @host localhost:8083
// @BasePath /
// @schemes http

func main() {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		log.Fatal("[ERROR] [API-Service] Failed to get current file path")
	}

	cmdDir := filepath.Dir(filename)

	projectRoot := filepath.Dir(filepath.Dir(cmdDir))

	configDir := filepath.Join(projectRoot, "API_service/internal/configs")

	viper.SetConfigName("config")
	viper.SetConfigType("yml")
	viper.AddConfigPath(configDir)

	err := viper.ReadInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Printf("[ERROR] [API-Service] Config file not found; using defaults or environment variables")
		} else {
			log.Fatalf("[ERROR] [API-Service] Error reading config file: %s", err)
		}
	}
	var config configs.Config

	err = viper.Unmarshal(&config)
	if err != nil {
		log.Fatalf("[ERROR] [API-Service] Unable to decode into struct, %v", err)
	}

	certFile := resolvePath(config.SSL.CertFile, configDir)
	keyFile := resolvePath(config.SSL.KeyFile, configDir)
	if certFile == "" || keyFile == "" {
		log.Fatalf("[ERROR] [API-Service] Required Cert/Key to initialize server")
	}
	if _, err := os.Stat(certFile); os.IsNotExist(err) {
		log.Fatalf("[ERROR] [API-Service] Cert file not found: %s", certFile)
	}
	if _, err := os.Stat(keyFile); os.IsNotExist(err) {
		log.Fatalf("[ERROR] [API-Service] Key file not found: %s", keyFile)
	}
	kafkaprod := kafka.NewKafkaProducer(config.Kafka)
	defer kafkaprod.Close()
	grpcclient := client.NewGrpcClient(config.SessionService)
	defer grpcclient.Close()
	middleware := middleware.NewMiddleware(grpcclient, kafkaprod)
	handler := handlers.NewHandler(middleware, config.Routes)
	srv := &server.Server{}
	port := config.Server.Port
	if port == "" {
		port = "8083"
	}
	serverError := make(chan error, 1)
	go func() {
		if err := srv.Run(port, handler.InitRoutes(), certFile, keyFile); err != nil {
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
	middleware.Stop()
	shutdownTimeout := 5 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	log.Println("[INFO] [API-Service] Service is shutting down...")

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("[INFO] [API-Service] Server shutdown error: %v", err)
	}

	log.Println("[INFO] [API-Service] Service has shutted down successfully")
}
func resolvePath(relativePath string, baseDir string) string {
	if relativePath == "" {
		return ""
	}
	if filepath.IsAbs(relativePath) {
		return relativePath
	}
	return filepath.Join(baseDir, relativePath)
}
