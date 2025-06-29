package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/niktin06sash/MicroserviceProject/Photo_service/internal/brokers/kafka"
)

type UserEvent struct {
	UserID  string `json:"userid"`
	Traceid string `json:"traceid"`
}

const (
	userRegistrationKey = "user.registration"
	userDeleteKey       = "user.delete"
)

func (rc *RabbitConsumer) readEvent() {
	const place = "RabbitConsumer-ReadEvent"
	msgs, err := rc.channel.Consume(
		rc.queue.Name,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Printf("[DEBUG] [Photo-Service] Failed to consume messages: %v", err)
		return
	}
	for {
		select {
		case <-rc.ctx.Done():
			return
		case msg, ok := <-msgs:
			if !ok {
				rc.logproducer.NewPhotoLog(kafka.LogLevelInfo, place, "", "Rabbit's channel closed, stopping worker")
				return
			}
			var newmsg UserEvent
			err := json.Unmarshal(msg.Body, &newmsg)
			if err != nil {
				rc.logproducer.NewPhotoLog(kafka.LogLevelError, place, newmsg.Traceid, fmt.Sprintf("Failed to unmarshal message: %v", err))
				msg.Nack(false, false)
				continue
			}
			ctx, cancel := context.WithTimeout(rc.ctx, 5*time.Second)
			defer cancel()
			switch msg.RoutingKey {
			case userRegistrationKey:
				rc.logproducer.NewPhotoLog(kafka.LogLevelInfo, place, newmsg.Traceid, fmt.Sprintf("Received user registration event for userID: %s", newmsg.UserID))
				resp := rc.userservice.AddUserId(ctx, newmsg.UserID, newmsg.Traceid)
				if resp.Errors != nil {
					continue
				}
			case userDeleteKey:
				rc.logproducer.NewPhotoLog(kafka.LogLevelInfo, place, newmsg.Traceid, fmt.Sprintf("Received user delete account event for userID: %s", newmsg.UserID))
				rc.wg.Add(1)
				go func() {
					defer rc.wg.Done()
					rc.userservice.DeleteAllUserData(ctx, newmsg.UserID, newmsg.Traceid)
				}()
			}
			err = msg.Ack(false)
			if err != nil {
				rc.logproducer.NewPhotoLog(kafka.LogLevelError, place, newmsg.Traceid, fmt.Sprintf("Failed to acknowledge message: %v", err))
			}
		}
	}
}
