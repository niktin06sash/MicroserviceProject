package service

import (
	"context"
	"fmt"

	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/erro"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/kafka"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/repository"
)

type FriendshipServiceImplement struct {
	Dbrepo        repository.DBFriendshipRepos
	Dbtxmanager   repository.DBTransactionManager
	KafkaProducer kafka.KafkaProducerService
}

func NewFriendshipServiceImplement(dbrepo repository.DBFriendshipRepos, dbtxmanager repository.DBTransactionManager, kafkaProd kafka.KafkaProducerService) *FriendshipServiceImplement {
	return &FriendshipServiceImplement{Dbrepo: dbrepo, Dbtxmanager: dbtxmanager, KafkaProducer: kafkaProd}
}
func (as *FriendshipServiceImplement) GetMyFriends(ctx context.Context, useridstr string) *ServiceResponse {
	const place = GetMyFriends
	getMyFriendsMap := make(map[string]string)
	traceid := ctx.Value("traceID").(string)
	userid, err := parsingUserId(useridstr, getMyFriendsMap, traceid, place, as.KafkaProducer)
	if err != nil {
		return &ServiceResponse{Success: false, Errors: getMyFriendsMap, ErrorType: erro.ServerErrorType}
	}
	bdresponse, serviceresponse := requestToDB(as.Dbrepo.GetMyFriends(ctx, userid), getMyFriendsMap)
	if serviceresponse != nil {
		return serviceresponse
	}
	msg := fmt.Sprintf("You have %v friends", len(bdresponse.Data))
	return &ServiceResponse{Success: true, Data: map[string]any{"message": msg, "friendsID": bdresponse.Data}}
}
