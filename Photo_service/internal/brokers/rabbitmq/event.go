package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/niktin06sash/MicroserviceProject/Photo_service/internal/brokers/kafka"
	"github.com/niktin06sash/MicroserviceProject/Photo_service/internal/erro"
	"github.com/niktin06sash/MicroserviceProject/Photo_service/internal/model"
	"github.com/niktin06sash/MicroserviceProject/Photo_service/internal/repository"
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
	defer rc.wg.Done()
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
				resp := rc.userrepo.AddUserId(ctx, newmsg.UserID)
				if resp.Errors != nil {
					rc.logproducer.NewPhotoLog(kafka.LogLevelError, resp.Place, newmsg.Traceid, resp.Errors[erro.ErrorMessage])
					continue
				}
				rc.logproducer.NewPhotoLog(kafka.LogLevelInfo, resp.Place, newmsg.Traceid, resp.SuccessMessage)

			case userDeleteKey:
				rc.logproducer.NewPhotoLog(kafka.LogLevelInfo, place, newmsg.Traceid, fmt.Sprintf("Received user delete account event for userID: %s", newmsg.UserID))
				rc.wg.Add(1)
				go func() {
					rc.deleteAllUserData(ctx, newmsg.UserID, newmsg.Traceid)
				}()
			}
			err = msg.Ack(false)
			if err != nil {
				rc.logproducer.NewPhotoLog(kafka.LogLevelError, place, newmsg.Traceid, fmt.Sprintf("Failed to acknowledge message: %v", err))
			}
		}
	}
}
func (rc *RabbitConsumer) deleteAllUserData(ctx context.Context, userid string, traceid string) {
	defer rc.wg.Done()
	const place = "RabbitConsumer-DeleteAllUserData"
	getresp := rc.userrepo.GetPhotos(ctx, userid)
	if getresp.Errors != nil {
		rc.logproducer.NewPhotoLog(kafka.LogLevelError, getresp.Place, traceid, getresp.Errors[erro.ErrorMessage])
		return
	}
	rc.logproducer.NewPhotoLog(kafka.LogLevelInfo, getresp.Place, traceid, getresp.SuccessMessage)
	photos := getresp.Data[repository.KeyPhoto].([]*model.Photo)
	delresp := rc.userrepo.DeleteUserData(ctx, userid)
	if delresp.Errors != nil {
		rc.logproducer.NewPhotoLog(kafka.LogLevelError, getresp.Place, traceid, getresp.Errors[erro.ErrorMessage])
		return
	}
	rc.logproducer.NewPhotoLog(kafka.LogLevelInfo, delresp.Place, traceid, delresp.SuccessMessage)
	select {
	case <-ctx.Done():
		rc.logproducer.NewPhotoLog(kafka.LogLevelError, place, traceid, "Context canceled or timeout")
	default:
		for _, photo := range photos {
			photodelresp := rc.photocloud.DeleteFile(photo.ID, photo.ContentType)
			if photodelresp.Errors != nil {
				rc.logproducer.NewPhotoLog(kafka.LogLevelError, photodelresp.Place, traceid, photodelresp.Errors[erro.ErrorMessage])
			}
			rc.logproducer.NewPhotoLog(kafka.LogLevelInfo, photodelresp.Place, traceid, photodelresp.SuccessMessage)
		}
	}
}
