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
	ctx, cancel := context.WithCancel(context.Background())
	log.Println("[DEBUG] [User-Service] Successful connect to Rabbit-Producer")
	return &RabbitProducer{conn: conn, channel: channel, config: config, logProducer: kafkaprod, context: ctx, cancel: cancel}, nil
}
func (rp *RabbitProducer) Close() {
	rp.channel.Close()
	rp.cancel()
	rp.conn.Close()
	log.Println("[DEBUG] [User-Service] Successful close Rabbit-Producer")
}
