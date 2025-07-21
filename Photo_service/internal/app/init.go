package app

import (
	"context"
	"log"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/niktin06sash/MicroserviceProject/Photo_service/internal/brokers/kafka"
	"github.com/niktin06sash/MicroserviceProject/Photo_service/internal/brokers/rabbitmq"
	"github.com/niktin06sash/MicroserviceProject/Photo_service/internal/configs"
	"github.com/niktin06sash/MicroserviceProject/Photo_service/internal/handlers"
	"github.com/niktin06sash/MicroserviceProject/Photo_service/internal/repository/cache"
	"github.com/niktin06sash/MicroserviceProject/Photo_service/internal/repository/cloud"
	"github.com/niktin06sash/MicroserviceProject/Photo_service/internal/repository/database"
	"github.com/niktin06sash/MicroserviceProject/Photo_service/internal/server"
	"github.com/niktin06sash/MicroserviceProject/Photo_service/internal/service"
)

type PhotoApplication struct {
	config configs.Config
	server *server.GrpcServer
}

func NewPhotoApplication(config configs.Config) *PhotoApplication {
	return &PhotoApplication{config: config}
}

func (a *PhotoApplication) Start() error {
	defer func() {
		buf := make([]byte, 1<<20)
		n := runtime.Stack(buf, true)
		log.Printf("[DEBUG] [Photo-Service] Count of active goroutines: %v", runtime.NumGoroutine())
		log.Printf("[DEBUG] [Photo-Service] Active goroutines:\n%s", buf[:n])
	}()
	pg, err := database.NewPostgresConnection(a.config.Database)
	if err != nil {
		return err
	}
	defer pg.Close()
	mega, err := cloud.NewMegaConnection(a.config.Mega)
	if err != nil {
		return err
	}
	defer mega.Close()
	redis, err := cache.NewRedisConnection(a.config.Redis)
	if err != nil {
		return err
	}
	defer redis.Close()
	kafkaProducer := kafka.NewKafkaProducer(a.config.Kafka)
	defer kafkaProducer.Close()
	defer kafkaProducer.LogClose()
	database := database.NewPhotoDatabase(pg)
	cache := cache.NewPhotoCache(redis)
	cloud := cloud.NewPhotoCloud(mega)
	photoService := service.NewPhotoServiceImplement(database, cloud, cache, kafkaProducer)
	defer photoService.Stop()
	rabbitConsumer, err := rabbitmq.NewRabbitConsumer(a.config.RabbitMQ, kafkaProducer, photoService)
	if err != nil {
		return err
	}
	defer rabbitConsumer.Close()
	api := handlers.NewPhotoAPI(photoService, kafkaProducer)
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
		log.Printf("[DEBUG] [Photo-Service] Server shutting down with signal: %v", sig)
	case err := <-serverError:
		log.Printf("[DEBUG] [Photo-Service] Server startup failed: %v", err)
		return err
	}

	return a.Stop()
}
func (a *PhotoApplication) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), a.config.Server.GracefulShutdown)
	defer cancel()
	log.Println("[DEBUG] [Photo-Service] Server is shutting down...")
	if err := a.server.Shutdown(ctx); err != nil {
		log.Printf("[DEBUG] [Photo-Service] Server shutdown error: %v", err)
		return err
	}
	log.Println("[DEBUG] [Photo-Service] Server has shutted down successfully")
	return nil
}
