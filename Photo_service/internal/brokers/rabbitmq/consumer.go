package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/niktin06sash/MicroserviceProject/Photo_service/internal/brokers/kafka"
	"github.com/niktin06sash/MicroserviceProject/Photo_service/internal/erro"
	"github.com/niktin06sash/MicroserviceProject/Photo_service/internal/model"
	"github.com/niktin06sash/MicroserviceProject/Photo_service/internal/repository"
)

type UserMessage struct {
	UserID  string `json:"userid"`
	Traceid string `json:"traceid"`
}
type DBUserRepos interface {
	AddUserId(ctx context.Context, userid string) *repository.RepositoryResponse
	DeleteUserData(ctx context.Context, userid string) *repository.RepositoryResponse
	GetPhotos(ctx context.Context, userid string) *repository.RepositoryResponse
}
type PhotoCloudRepos interface {
	DeleteFile(id, ext string) *repository.RepositoryResponse
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
				rc.kafkaproducer.NewPhotoLog(kafka.LogLevelInfo, place, "", "Rabbit's channel closed, stopping worker")
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
				rc.kafkaproducer.NewPhotoLog(kafka.LogLevelInfo, place, newmsg.Traceid, fmt.Sprintf("Received user registration event for userID: %s", newmsg.UserID))
				resp := rc.userrepo.AddUserId(ctx, newmsg.UserID)
				if resp.Errors != nil {
					rc.kafkaproducer.NewPhotoLog(kafka.LogLevelError, resp.Place, newmsg.Traceid, resp.Errors[erro.ErrorMessage])
					continue
				}
				rc.kafkaproducer.NewPhotoLog(kafka.LogLevelInfo, resp.Place, newmsg.Traceid, resp.SuccessMessage)

			case "user.delete":
				rc.kafkaproducer.NewPhotoLog(kafka.LogLevelInfo, place, newmsg.Traceid, fmt.Sprintf("Received user delete account event for userID: %s", newmsg.UserID))
				go rc.deleteAllUserData(ctx, newmsg.UserID, newmsg.Traceid)
			}
			err = msg.Ack(false)
			if err != nil {
				rc.kafkaproducer.NewPhotoLog(kafka.LogLevelError, place, newmsg.Traceid, fmt.Sprintf("Failed to acknowledge message: %v", err))
			}
		}
	}
}
func (rc *RabbitConsumer) deleteAllUserData(ctx context.Context, userid string, traceid string) {
	const place = "RabbitConsumer-DeleteAllUserData"
	getresp := rc.userrepo.GetPhotos(ctx, userid)
	if getresp.Errors != nil {
		rc.kafkaproducer.NewPhotoLog(kafka.LogLevelError, getresp.Place, traceid, getresp.Errors[erro.ErrorMessage])
		return
	}
	rc.kafkaproducer.NewPhotoLog(kafka.LogLevelInfo, getresp.Place, traceid, getresp.SuccessMessage)
	photos := getresp.Data[repository.KeyPhoto].([]*model.Photo)
	delresp := rc.userrepo.DeleteUserData(ctx, userid)
	if delresp.Errors != nil {
		rc.kafkaproducer.NewPhotoLog(kafka.LogLevelError, getresp.Place, traceid, getresp.Errors[erro.ErrorMessage])
		return
	}
	rc.kafkaproducer.NewPhotoLog(kafka.LogLevelInfo, delresp.Place, traceid, delresp.SuccessMessage)
	select {
	case <-ctx.Done():
		rc.kafkaproducer.NewPhotoLog(kafka.LogLevelError, place, traceid, "Context canceled or timeout")
	default:
		for _, photo := range photos {
			photodelresp := rc.photocloud.DeleteFile(photo.ID, photo.ContentType)
			if photodelresp.Errors != nil {
				rc.kafkaproducer.NewPhotoLog(kafka.LogLevelError, photodelresp.Place, traceid, photodelresp.Errors[erro.ErrorMessage])
			}
			rc.kafkaproducer.NewPhotoLog(kafka.LogLevelInfo, photodelresp.Place, traceid, photodelresp.SuccessMessage)
		}
	}
}
