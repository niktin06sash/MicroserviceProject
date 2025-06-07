package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/niktin06sash/MicroserviceProject/Photo_service/internal/brokers/kafka"
	"github.com/niktin06sash/MicroserviceProject/Photo_service/internal/configs"
	"github.com/niktin06sash/MicroserviceProject/Photo_service/internal/erro"
	"github.com/niktin06sash/MicroserviceProject/Photo_service/internal/repository"
	"github.com/streadway/amqp"
)

type DBUserIDRepos interface {
	AddUserId(ctx context.Context, userid string) *repository.RepositoryResponse
	DeleteUserData(ctx context.Context, userid string) *repository.RepositoryResponse
}
type LogProducer interface {
	NewPhotoLog(level, place, traceid, msg string)
}
type RabbitConsumer struct {
	conn          *amqp.Connection
	channel       *amqp.Channel
	queue         amqp.Queue
	userrepo      DBUserIDRepos
	kafkaproducer LogProducer
	wg            *sync.WaitGroup
	ctx           context.Context
	cancel        context.CancelFunc
}

func NewRabbitConsumer(config configs.RabbitMQConfig, kafkaprod LogProducer, dbrepo DBUserIDRepos) (*RabbitConsumer, error) {
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
	go rc.startConsuming()
	log.Printf("[DEBUG] [Logs-Service] Successful connect to Rabbit-Consumer")
	return rc, nil
}

type UserMessage struct {
	UserID  string `json:"userid"`
	Traceid string `json:"traceid"`
}

func (rc *RabbitConsumer) Close() {
	rc.cancel()
	rc.wg.Wait()
	rc.channel.Close()
	rc.conn.Close()
	log.Printf("[DEBUG] [Logs-Service] Successful close Rabbit-Consumer")
}
func (rc *RabbitConsumer) startConsuming() {
	const place = "RabbitConsumer"
	msgs, _ := rc.channel.Consume(
		rc.queue.Name,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	defer rc.wg.Done()
	for {
		select {
		case <-rc.ctx.Done():
			return
		case msg, ok := <-msgs:
			if !ok {
				log.Printf("[INFO] [Photo-Service] Rabbit's channel closed, stopping worker")
				return
			}
			var newmsg UserMessage
			err := json.Unmarshal(msg.Body, &newmsg)
			if err != nil {
				rc.kafkaproducer.NewPhotoLog(kafka.LogLevelError, place, newmsg.Traceid, fmt.Sprintf("Failed to unmarshal message: %v", err))
				msg.Nack(false, false)
				continue
			}
			ctx, cancel := context.WithTimeout(rc.ctx, 5*time.Second)
			defer cancel()
			switch msg.RoutingKey {
			case "user.registration":
				resp := rc.userrepo.AddUserId(ctx, newmsg.UserID)
				if resp.Errors != nil {
					rc.kafkaproducer.NewPhotoLog(kafka.LogLevelError, place, newmsg.Traceid, resp.Errors[erro.ErrorMessage])
					continue
				}
				rc.kafkaproducer.NewPhotoLog(kafka.LogLevelInfo, place, newmsg.Traceid, fmt.Sprintf("Received user registration event for userID: %s", newmsg.UserID))

			case "user.delete":
				resp := rc.userrepo.DeleteUserData(ctx, newmsg.UserID)
				if resp.Errors != nil {
					rc.kafkaproducer.NewPhotoLog(kafka.LogLevelError, place, newmsg.Traceid, resp.Errors[erro.ErrorMessage])
					continue
				}
				rc.kafkaproducer.NewPhotoLog(kafka.LogLevelInfo, place, newmsg.Traceid, fmt.Sprintf("Received user delete account event for userID: %s", newmsg.UserID))
			}
			err = msg.Ack(false)
			if err != nil {
				rc.kafkaproducer.NewPhotoLog(kafka.LogLevelError, place, newmsg.Traceid, fmt.Sprintf("Failed to acknowledge message: %v", err))
			}
		}
	}
}
