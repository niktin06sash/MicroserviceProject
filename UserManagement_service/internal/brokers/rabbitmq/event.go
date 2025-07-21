package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/brokers/kafka"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/metrics"
	"github.com/streadway/amqp"
)

type UserEvent struct {
	Userid  string `json:"userid"`
	Traceid string `json:"traceid"`
}

const (
	UserRegistrationKey = "user.registration"
	UserDeleteKey       = "user.delete"
)

func (rp *RabbitProducer) NewUserEvent(ctx context.Context, routingKey string, userid string, place string, traceid string) error {
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
		case <-ctx.Done():
			rp.logProducer.NewUserLog(kafka.LogLevelWarn, place, traceid, "Request context was canceled")
			return ctx.Err()
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
				select {
				case confirmed := <-rp.confirmchan:
					if confirmed.Ack {
						rp.logProducer.NewUserLog(kafka.LogLevelInfo, place, traceid, fmt.Sprintf("User Event with routing key: %s was published on attempt %d", routingKey, attempt))
						metrics.UserRabbitProducerEventsSent.WithLabelValues(routingKey)
						return nil
					}
					rp.logProducer.NewUserLog(kafka.LogLevelWarn, place, traceid, fmt.Sprintf("Attempt %d failed to confirm message", attempt))
				case <-time.After(5 * time.Second):
					rp.logProducer.NewUserLog(kafka.LogLevelWarn, place, traceid, fmt.Sprintf("Confirmation timeout (attempt %d)", attempt))
					metrics.UserRabbitProducerErrorsTotal.WithLabelValues(routingKey).Inc()
				case <-ctx.Done():
					rp.logProducer.NewUserLog(kafka.LogLevelError, place, traceid, "Request context was canceled")
					return ctx.Err()
				}
				time.Sleep(time.Duration(attempt) * time.Second)
			}
		}
	}
	rp.logProducer.NewUserLog(kafka.LogLevelError, place, traceid, "All attempts failed to User Event")
	return err
}
