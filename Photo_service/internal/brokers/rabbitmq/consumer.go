package rabbitmq

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"sync"

	"github.com/niktin06sash/MicroserviceProject/Photo_service/internal/configs"
	"github.com/niktin06sash/MicroserviceProject/Photo_service/internal/repository"
	"github.com/streadway/amqp"
)

type LogProducer interface {
	NewPhotoLog(level, place, traceid, msg string)
}
type DBUserRepos interface {
	AddUserId(ctx context.Context, userid string) *repository.RepositoryResponse
	DeleteUserData(ctx context.Context, userid string) *repository.RepositoryResponse
	GetPhotos(ctx context.Context, userid string) *repository.RepositoryResponse
}
type PhotoCloudRepos interface {
	DeleteFile(id, ext string) *repository.RepositoryResponse
}
type RabbitConsumer struct {
	conn          *amqp.Connection
	channel       *amqp.Channel
	queue         amqp.Queue
	userrepo      DBUserRepos
	kafkaproducer LogProducer
	photocloud    PhotoCloudRepos
	wg            *sync.WaitGroup
	ctx           context.Context
	cancel        context.CancelFunc
}

func NewRabbitConsumer(config configs.RabbitMQConfig, kafkaprod LogProducer, dbrepo DBUserRepos, photocloud PhotoCloudRepos) (*RabbitConsumer, error) {
	connstr := fmt.Sprintf("amqp://%s:%s@%s:%s/", config.Name, config.Password, config.Host, strconv.Itoa(config.Port))
	conn, err := amqp.Dial(connstr)
	if err != nil {
		log.Printf("[DEBUG] [Photo-Service] Failed to connect to RabbitMQ: %v", err)
		return nil, err
	}
	channel, err := conn.Channel()
	if err != nil {
		log.Printf("[DEBUG] [Photo-Service] Failed to open to channel: %v", err)
		return nil, err
	}
	err = channel.ExchangeDeclare(
		config.Exchange,
		"topic",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Printf("[DEBUG] [Photo-Service] Failed to declare an exchange: %v", err)
		return nil, err
	}
	queue, err := channel.QueueDeclare(
		config.Queue,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Printf("[DEBUG] [Photo-Service] Failed to declare an queue: %v", err)
		return nil, err
	}
	routingKeys := []string{"user.registration", "user.delete"}
	for _, routingKey := range routingKeys {
		err = channel.QueueBind(
			queue.Name,
			routingKey,
			config.Exchange,
			false,
			nil,
		)
		if err != nil {
			log.Printf("[DEBUG] [Photo-Service] Failed to bind queue with routing key %s: %v", routingKey, err)
			return nil, err
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	rc := &RabbitConsumer{
		conn:          conn,
		channel:       channel,
		queue:         queue,
		userrepo:      dbrepo,
		ctx:           ctx,
		cancel:        cancel,
		wg:            &sync.WaitGroup{},
		kafkaproducer: kafkaprod,
	}
	rc.wg.Add(1)
	go rc.readEvent()
	log.Printf("[DEBUG] [Logs-Service] Successful connect to Rabbit-Consumer")
	return rc, nil
}
func (rc *RabbitConsumer) Close() {
	rc.cancel()
	rc.wg.Wait()
	rc.channel.Close()
	rc.conn.Close()
	log.Printf("[DEBUG] [Logs-Service] Successful close Rabbit-Consumer")
}
