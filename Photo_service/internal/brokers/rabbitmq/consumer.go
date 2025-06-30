package rabbitmq

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"sync"

	"github.com/niktin06sash/MicroserviceProject/Photo_service/internal/configs"
	"github.com/niktin06sash/MicroserviceProject/Photo_service/internal/service"
	"github.com/streadway/amqp"
)

type LogProducer interface {
	NewPhotoLog(level, place, traceid, msg string)
}
type UserIDService interface {
	DeleteAllUserData(ctx context.Context, userid string, traceid string) *service.ServiceResponse
	AddUserId(ctx context.Context, userid string, traceid string) *service.ServiceResponse
}
type RabbitConsumer struct {
	conn        *amqp.Connection
	channel     *amqp.Channel
	queue       amqp.Queue
	userservice UserIDService
	logproducer LogProducer
	wg          *sync.WaitGroup
	ctx         context.Context
	cancel      context.CancelFunc
}

func NewRabbitConsumer(config configs.RabbitMQConfig, logproducer LogProducer, userservice UserIDService) (*RabbitConsumer, error) {
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
	routingKeys := []string{userRegistrationKey, userDeleteKey}
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
		conn:        conn,
		channel:     channel,
		queue:       queue,
		userservice: userservice,
		ctx:         ctx,
		cancel:      cancel,
		wg:          &sync.WaitGroup{},
		logproducer: logproducer,
	}
	rc.wg.Add(1)
	go func() {
		defer rc.wg.Done()
		rc.readEvent()
	}()
	log.Printf("[DEBUG] [Photo-Service] Successful connect to Rabbit-Consumer")
	return rc, nil
}
func (rc *RabbitConsumer) Close() {
	rc.cancel()
	rc.channel.Close()
	rc.wg.Wait()
	rc.conn.Close()
	log.Printf("[DEBUG] [Photo-Service] Successful close Rabbit-Consumer")
}
