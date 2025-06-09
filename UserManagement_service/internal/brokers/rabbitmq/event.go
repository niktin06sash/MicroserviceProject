package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/brokers/kafka"
	"github.com/streadway/amqp"
)

type UserEvent struct {
	Userid  string `json:"userid"`
	Traceid string `json:"traceid"`
}

func (rp *RabbitProducer) NewUserEvent(ctx context.Context, routingKey string, userid string, place string, traceid string) error {
	_ = rp.channel.Confirm(false)
	confirms := rp.channel.NotifyPublish(make(chan amqp.Confirmation, 1))
	body, err := json.Marshal(&UserEvent{Userid: userid, Traceid: traceid})
	if err != nil {
		rp.logProducer.NewUserLog(kafka.LogLevelError, place, traceid, fmt.Sprintf("Failed to marshal message: %v", err))
		return err
	}
	for attempt := 1; attempt <= 3; attempt++ {
		select {
		case <-rp.context.Done():
			rp.logProducer.NewUserLog(kafka.LogLevelError, place, traceid, "RabbitProducer's context was canceled")
			return err
		default:
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
				rp.logProducer.NewUserLog(kafka.LogLevelInfo, place, traceid, fmt.Sprintf("User Event with routing key: %s was published on attempt %d", routingKey, attempt))
				return nil
			}
			rp.logProducer.NewUserLog(kafka.LogLevelWarn, place, traceid, fmt.Sprintf("Attempt %d failed to User Event: %v", attempt, err))
			time.Sleep(time.Second)
		}
	}
	select {
	case <-confirms:
		return nil
	case <-ctx.Done():
		rp.logProducer.NewUserLog(kafka.LogLevelError, place, traceid, "Request context was canceled")
		return err
	}
}
