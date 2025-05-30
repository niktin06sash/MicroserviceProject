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
	Redisrepo     repository.CacheFriendshipRepos
}

func NewFriendshipServiceImplement(dbrepo repository.DBFriendshipRepos, dbtxmanager repository.DBTransactionManager, redis repository.CacheFriendshipRepos, kafkaProd kafka.KafkaProducerService) *FriendshipServiceImplement {
	return &FriendshipServiceImplement{Dbrepo: dbrepo, Dbtxmanager: dbtxmanager, KafkaProducer: kafkaProd, Redisrepo: redis}
}
func (as *FriendshipServiceImplement) GetMyFriends(ctx context.Context, useridstr string) *ServiceResponse {
	const place = GetMyFriends
	getMyFriendsMap := make(map[string]string)
	traceid := ctx.Value("traceID").(string)
	userid, err := parsingUserId(useridstr, getMyFriendsMap, traceid, place, as.KafkaProducer)
	if err != nil {
		return &ServiceResponse{Success: false, Errors: getMyFriendsMap, ErrorType: erro.ServerErrorType}
	}
	redisresponse, serviceresponse := requestToDB(as.Redisrepo.GetFriendsCache(ctx, useridstr), getMyFriendsMap)
	if serviceresponse != nil {
		return serviceresponse
	}
	if redisresponse.Success {
		friendslice := redisresponse.Data[repository.KeyFriends].([]string)
		msg := fmt.Sprintf("You have %v friends", len(friendslice))
		return &ServiceResponse{Success: true, Data: map[string]any{"message": msg, repository.KeyFriends: friendslice}}
	}
	bdresponse, serviceresponse := requestToDB(as.Dbrepo.GetMyFriends(ctx, userid), getMyFriendsMap)
	if serviceresponse != nil {
		return serviceresponse
	}
	_, serviceresponse = requestToDB(as.Redisrepo.AddFriendsCache(ctx, useridstr, bdresponse.Data), getMyFriendsMap)
	if serviceresponse != nil {
		return serviceresponse
	}
	friendslice := bdresponse.Data[repository.KeyFriends].([]string)
	msg := fmt.Sprintf("You have %v friends", len(friendslice))
	return &ServiceResponse{Success: true, Data: map[string]any{"message": msg, repository.KeyFriends: friendslice}}
}
