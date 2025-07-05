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
	"github.com/niktin06sash/MicroserviceProject/Photo_service/internal/repository"
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
	db, err := repository.NewDatabaseConnection(a.config.Database)
	if err != nil {
		return err
	}
	defer db.Close()
	cloud, err := repository.NewCloudConnection(a.config.Mega)
	if err != nil {
		return err
	}
	defer cloud.Close()
	cache, err := repository.NewRedisConnection(a.config.Redis)
	if err != nil {
		return err
	}
	defer cache.Close()
	kafkaProducer := kafka.NewKafkaProducer(a.config.Kafka)
	defer kafkaProducer.Close()
	defer kafkaProducer.LogClose()
	postgresRepo := repository.NewPhotoPostgresRepo(db)
	redisRepo := repository.NewPhotoRedisRepo(cache)
	mega := repository.NewMegaClientRepo(cloud)
	photoService := service.NewPhotoServiceImplement(postgresRepo, mega, redisRepo, kafkaProducer)
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
	log.Println("Server is shutting down...")
	if err := a.server.Shutdown(ctx); err != nil {
		log.Printf("[DEBUG] [Photo-Service] Server shutdown error: %v", err)
		return err
	}
	log.Println("[DEBUG] [Photo-Service] Server has shutted down successfully")
	return nil
}
