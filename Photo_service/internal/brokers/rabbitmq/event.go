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
				resp := rc.userrepo.AddUserId(ctx, newmsg.UserID)
				if resp.Errors != nil {
					rc.logproducer.NewPhotoLog(kafka.LogLevelError, resp.Place, newmsg.Traceid, resp.Errors.Message)
					continue
				}
				rc.logproducer.NewPhotoLog(kafka.LogLevelInfo, resp.Place, newmsg.Traceid, resp.SuccessMessage)

			case userDeleteKey:
				rc.logproducer.NewPhotoLog(kafka.LogLevelInfo, place, newmsg.Traceid, fmt.Sprintf("Received user delete account event for userID: %s", newmsg.UserID))
				rc.wg.Add(1)
				go func() {
					defer rc.wg.Done()
					rc.deleteAllUserData(newmsg.UserID, newmsg.Traceid)
				}()
			}
			err = msg.Ack(false)
			if err != nil {
				rc.logproducer.NewPhotoLog(kafka.LogLevelError, place, newmsg.Traceid, fmt.Sprintf("Failed to acknowledge message: %v", err))
			}
		}
	}
}
func (rc *RabbitConsumer) deleteAllUserData(userid string, traceid string) {
	getctx, cancel := context.WithTimeout(rc.ctx, 5*time.Second)
	defer cancel()
	getresp := rc.userrepo.GetPhotos(getctx, userid)
	if getresp.Errors != nil {
		rc.logproducer.NewPhotoLog(kafka.LogLevelError, getresp.Place, traceid, getresp.Errors.Message)
		return
	}
	rc.logproducer.NewPhotoLog(kafka.LogLevelInfo, getresp.Place, traceid, getresp.SuccessMessage)
	photos := getresp.Data.Photos
	deluserctx, cancel := context.WithTimeout(rc.ctx, 15*time.Second)
	defer cancel()
	delresp := rc.userrepo.DeleteUserData(deluserctx, userid)
	if delresp.Errors != nil {
		rc.logproducer.NewPhotoLog(kafka.LogLevelError, delresp.Place, traceid, delresp.Errors.Message)
		return
	}
	cacheresp := rc.cache.DeletePhotosCache(deluserctx, userid)
	if cacheresp.Errors != nil {
		rc.logproducer.NewPhotoLog(kafka.LogLevelError, cacheresp.Place, traceid, cacheresp.Errors.Message)
		return
	}
	rc.logproducer.NewPhotoLog(kafka.LogLevelInfo, delresp.Place, traceid, delresp.SuccessMessage)
	for _, photo := range photos {
		photodelresp := rc.photocloud.DeleteFile(rc.ctx, photo.ID, photo.ContentType)
		if photodelresp.Errors != nil {
			rc.logproducer.NewPhotoLog(kafka.LogLevelError, photodelresp.Place, traceid, photodelresp.Errors.Message)
		}
		rc.logproducer.NewPhotoLog(kafka.LogLevelInfo, photodelresp.Place, traceid, photodelresp.SuccessMessage)
	}
}
