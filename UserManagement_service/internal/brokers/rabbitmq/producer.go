package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/brokers/kafka"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/configs"
	"github.com/streadway/amqp"
)

type RabbitProducer struct {
	conn      *amqp.Connection
	channel   *amqp.Channel
	config    configs.RabbitMQConfig
	kafkaprod kafka.KafkaProducerService
}

func NewRabbitProducer(config configs.RabbitMQConfig, kafkaprod kafka.KafkaProducerService) (*RabbitProducer, error) {
	connString := fmt.Sprintf("amqp://%s:%s@%s:%s/", config.Name, config.Password, config.Host, strconv.Itoa(config.Port))
	conn, err := amqp.Dial(connString)
	if err != nil {
		log.Printf("[DEBUG] [User-Service] Failed to connect to Rabbit-Producer: %v", err)
		return nil, err
	}
	channel, err := conn.Channel()
	if err != nil {
		log.Printf("[DEBUG] [User-Service] Failed to open a channel to Rabbit-Producer: %v", err)
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
		log.Printf("[DEBUG] [User-Service] Failed to declare an exchange to Rabbit-Producer: %v", err)
		return nil, err
	}

	return &RabbitProducer{conn: conn, channel: channel, config: config}, nil
}

type Event struct {
	Userid  string `json:"userid"`
	Traceid string `json:"traceid"`
}

func (rp *RabbitProducer) NewUserEvent(ctx context.Context, routingKey string, userid string, place string, traceid string) error {
	body, err := json.Marshal(Event{Userid: userid, Traceid: traceid})
	if err != nil {
		rp.kafkaprod.NewUserLog(kafka.LogLevelError, place, traceid, fmt.Sprintf("Failed to marshal message: %v", err))
		return err
	}
	for attempt := 1; attempt <= 3; attempt++ {
		err = rp.channel.Publish(
			rp.config.Exchange,
			routingKey,
			false,
			false,
			amqp.Publishing{
				ContentType: "application/json",
				Body:        body,
			},
		)
		if err == nil {
			rp.kafkaprod.NewUserLog(kafka.LogLevelInfo, place, traceid, fmt.Sprintf("User Event with routing key: %s was published on attempt %d", routingKey, attempt))
			return nil
		}
		rp.kafkaprod.NewUserLog(kafka.LogLevelWarn, place, traceid, fmt.Sprintf("Attempt %d failed to User Event: %v", attempt, err))
		time.Sleep(time.Second)
	}
	return err
}
