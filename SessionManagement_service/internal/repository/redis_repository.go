package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/niktin06sash/MicroserviceProject/SessionManagement_service/internal/erro"
	"github.com/niktin06sash/MicroserviceProject/SessionManagement_service/internal/model"
)

type SessionRedis struct {
	Client *RedisObject
}

func NewSessionRepos(client *RedisObject) *SessionRedis {
	return &SessionRedis{Client: client}
}

func (redisrepo *SessionRedis) SetSession(ctx context.Context, session *model.Session) *RepositoryResponse {
	const place = SetSession
	err := redisrepo.Client.RedisClient.HSet(ctx, session.SessionID, map[string]interface{}{
		KeyUserId:    session.UserID,
		KeySessionId: session.ExpirationTime.Format(time.RFC3339),
	}).Err()
	if err != nil {
		return &RepositoryResponse{Success: false, Errors: erro.ServerError(fmt.Sprintf(erro.ErrorHset, err)), Place: place}
	}
	expiration := time.Until(session.ExpirationTime)
	err = redisrepo.Client.RedisClient.Expire(ctx, session.SessionID, expiration).Err()
	if err != nil {
		return &RepositoryResponse{Success: false, Errors: erro.ServerError(fmt.Sprintf(erro.ErrorExpire, err)), Place: place}
	}
	return &RepositoryResponse{Success: true, SuccessMessage: "Successfull session installation", Place: place}
}

func (redisrepo *SessionRedis) GetSession(ctx context.Context, sessionID string, flag string) *RepositoryResponse {
	const place = GetSession
	return redisrepo.getSessionData(ctx, sessionID, place, flag)
}
func (redisrepo *SessionRedis) getSessionData(ctx context.Context, sessionID string, place string, flag string) *RepositoryResponse {
	result, err := redisrepo.Client.RedisClient.HGetAll(ctx, sessionID).Result()
	if err != nil {
		return &RepositoryResponse{Success: false, Errors: erro.ServerError(fmt.Sprintf(erro.ErrorHgetAll, err)), Place: place}
	}
	if len(result) == 0 {
		if flag == "true" {
			return &RepositoryResponse{Success: false, Errors: erro.ClientError(erro.InvalidSession), Place: place}
		}
		return &RepositoryResponse{Success: true, SuccessMessage: "Request for an unauthorized page with invalid session", Place: place}
	}
	if flag == "false" {
		return &RepositoryResponse{Success: false, Errors: erro.ClientError(erro.AlreadyAuthorized), Place: place}
	}
	userIDString := result[KeyUserId]
	return &RepositoryResponse{Success: true, Data: Data{UserID: userIDString}, SuccessMessage: "Successfull get session", Place: place}
}
func (redisrepo *SessionRedis) DeleteSession(ctx context.Context, sessionID string) *RepositoryResponse {
	const place = DeleteSession
	num, err := redisrepo.Client.RedisClient.Del(ctx, sessionID).Result()
	if err != nil {
		return &RepositoryResponse{Success: false, Errors: erro.ServerError(fmt.Sprintf(erro.ErrorDelSession, err)), Place: place}
	}
	if num == 0 {
		return &RepositoryResponse{Success: false, Errors: erro.ClientError(erro.InvalidSession), Place: place}
	}
	return &RepositoryResponse{Success: true, SuccessMessage: "Successfull delete session", Place: place}
}
