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

func (redisrepo *SessionRedis) SetSession(ctx context.Context, session model.Session) *RepositoryResponse {
	const place = SetSession
	err := redisrepo.Client.RedisClient.HSet(ctx, session.SessionID, map[string]interface{}{
		"UserID":         session.UserID,
		"ExpirationTime": session.ExpirationTime.Format(time.RFC3339),
	}).Err()
	if err != nil {
		return &RepositoryResponse{Success: false, Errors: map[string]string{erro.ErrorType: erro.ServerErrorType, erro.ErrorMessage: fmt.Sprintf("Hset session error: %v", err)}, Place: place}
	}
	expiration := time.Until(session.ExpirationTime)
	err = redisrepo.Client.RedisClient.Expire(ctx, session.SessionID, expiration).Err()
	if err != nil {
		return &RepositoryResponse{Success: false, Errors: map[string]string{erro.ErrorType: erro.ServerErrorType, erro.ErrorMessage: fmt.Sprintf("Expire session error: %v", err)}, Place: place}
	}
	return &RepositoryResponse{Success: true, Data: map[string]any{KeySessionId: session.SessionID, KeyExpiryTime: session.ExpirationTime}, SuccessMessage: "Successfull session installation", Place: place}
}

func (redisrepo *SessionRedis) GetSession(ctx context.Context, sessionID string) *RepositoryResponse {
	const place = GetSession
	flag := ctx.Value("flagvalidate").(string)
	flagValidate := flag == "true"
	return redisrepo.getSessionData(ctx, sessionID, place, flagValidate)
}
func (redisrepo *SessionRedis) getSessionData(ctx context.Context, sessionID string, place string, flagvalidate bool) *RepositoryResponse {
	result, err := redisrepo.Client.RedisClient.HGetAll(ctx, sessionID).Result()
	if err != nil {
		return &RepositoryResponse{Success: false, Errors: map[string]string{erro.ErrorType: erro.ServerErrorType, erro.ErrorMessage: fmt.Sprintf("HGetAll session error: %v", err)}, Place: place}
	}
	if len(result) == 0 {
		if flagvalidate {
			return &RepositoryResponse{Success: false, Errors: map[string]string{erro.ErrorType: erro.ClientErrorType, erro.ErrorMessage: "Session is empty or invalid"}, Place: place}
		}
		return &RepositoryResponse{Success: true}
	}
	if !flagvalidate {
		return &RepositoryResponse{Success: false, Errors: map[string]string{erro.ErrorType: erro.ClientErrorType, erro.ErrorMessage: "Already Authorized"}, Place: place}
	}
	userIDString := result[KeyUserId]
	return &RepositoryResponse{Success: true, Data: map[string]any{KeyUserId: userIDString}, SuccessMessage: "Successfull get session", Place: place}
}
func (redisrepo *SessionRedis) DeleteSession(ctx context.Context, sessionID string) *RepositoryResponse {
	const place = DeleteSession
	num, err := redisrepo.Client.RedisClient.Del(ctx, sessionID).Result()
	if err != nil {
		return &RepositoryResponse{Success: false, Errors: map[string]string{erro.ErrorType: erro.ServerErrorType, erro.ErrorMessage: fmt.Sprintf("Del session error: %v", err)}, Place: place}
	}
	if num == 0 {
		return &RepositoryResponse{Success: false, Errors: map[string]string{erro.ErrorType: erro.ClientErrorType, erro.ErrorMessage: "Session is empty or invalid"}, Place: place}
	}
	return &RepositoryResponse{Success: true, SuccessMessage: "Successfull delete session", Place: place}
}
