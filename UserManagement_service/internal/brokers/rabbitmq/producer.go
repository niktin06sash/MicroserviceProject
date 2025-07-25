package rabbitmq

import (
	"context"
	"fmt"
	"log"
	"strconv"

	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/configs"
	"github.com/streadway/amqp"
)

type RabbitProducer struct {
	conn        *amqp.Connection
	channel     *amqp.Channel
	config      configs.RabbitMQConfig
	logProducer LogProducer
	context     context.Context
	cancel      context.CancelFunc
	confirmchan chan amqp.Confirmation
}
type LogProducer interface {
	NewUserLog(level, place, traceid, msg string)
}

func NewRabbitProducer(config configs.RabbitMQConfig, kafkaprod LogProducer) (*RabbitProducer, error) {
	connString := fmt.Sprintf("amqp://%s:%s@%s:%s/", config.Name, config.Password, config.Host, strconv.Itoa(config.Port))
	conn, err := amqp.Dial(connString)
	if err != nil {
		log.Printf("[DEBUG] [User-Service] Failed to connect to Rabbit-Producer: %v", err)
		return nil, err
	}
	channel, err := conn.Channel()
	if err != nil {
		conn.Close()
		log.Printf("[DEBUG] [User-Service] Failed to open a channel to Rabbit-Producer: %v", err)
		return nil, err
	}
	if err := channel.Confirm(false); err != nil {
		conn.Close()
		log.Printf("[DEBUG] [User-Service] Failed to create confirm-mode in Rabbit-Producer: %v", err)
		return nil, fmt.Errorf("failed to put channel in confirm mode: %v", err)
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
		conn.Close()
		log.Printf("[DEBUG] [User-Service] Failed to declare an exchange to Rabbit-Producer: %v", err)
		return nil, err
	}
	ctx, cancel := context.WithCancel(context.Background())
	confirmchan := channel.NotifyPublish(make(chan amqp.Confirmation, 100))
	log.Println("[DEBUG] [User-Service] Successful connect to Rabbit-Producer")
	return &RabbitProducer{conn: conn, channel: channel, config: config, logProducer: kafkaprod, context: ctx, cancel: cancel, confirmchan: confirmchan}, nil
}
func (rp *RabbitProducer) Close() {
	rp.cancel()
	rp.channel.Close()
	rp.conn.Close()
	log.Println("[DEBUG] [User-Service] Successful close Rabbit-Producer")
}
