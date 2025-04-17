package repository

import (
	"context"
	"time"

	"github.com/niktin06sash/MicroserviceProject/SessionManagement_service/internal/erro"
	"github.com/niktin06sash/MicroserviceProject/SessionManagement_service/internal/logger"
	"github.com/niktin06sash/MicroserviceProject/SessionManagement_service/internal/model"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type RedisClientInterface interface {
	HSet(ctx context.Context, key string, values ...interface{}) *redis.IntCmd
	Expire(ctx context.Context, key string, expiration time.Duration) *redis.BoolCmd
	HGetAll(ctx context.Context, key string) *redis.MapStringStringCmd
	Del(ctx context.Context, key ...string) *redis.IntCmd
}
type AuthRedis struct {
	Client RedisClientInterface
	logger *logger.SessionLogger
}

func (redisrepo *AuthRedis) SetSession(ctx context.Context, session model.Session) *RepositoryResponse {
	if ctx.Err() != nil {
		return &RepositoryResponse{Success: false, Errors: erro.ErrorContextTimeOut}
	}
	err := redisrepo.Client.HSet(ctx, session.SessionID, map[string]interface{}{
		"UserID":         session.UserID,
		"ExpirationTime": session.ExpirationTime.Format(time.RFC3339),
	}).Err()

	if err != nil {
		return &RepositoryResponse{Success: false, Errors: erro.ErrorSetSession}
	}
	expiration := time.Until(session.ExpirationTime)
	err = redisrepo.Client.Expire(ctx, session.SessionID, expiration).Err()
	if err != nil {
		return &RepositoryResponse{Success: false, Errors: erro.ErrorSetSession}
	}
	if ctx.Err() != nil {
		return &RepositoryResponse{Success: false, Errors: erro.ErrorContextTimeOut}
	}
	redisrepo.logger.Info("Successful session installation")
	return &RepositoryResponse{Success: true, SessionId: session.SessionID, ExpirationTime: session.ExpirationTime}
}

func (redisrepo *AuthRedis) GetSession(ctx context.Context, sessionID string) *RepositoryResponse {
	if ctx.Err() != nil {
		return &RepositoryResponse{Success: false, Errors: erro.ErrorContextTimeOut}
	}
	result, err := redisrepo.Client.HGetAll(ctx, sessionID).Result()
	if err != nil {
		return &RepositoryResponse{Success: false, Errors: erro.ErrorGetSession}
	}

	if len(result) == 0 {
		return &RepositoryResponse{Success: false, Errors: erro.ErrorInvalidSessionID}
	}

	userIDString, ok := result["UserID"]
	if !ok {
		return &RepositoryResponse{Success: false, Errors: erro.ErrorGetUserIdSession}
	}

	expirationTimeString, ok := result["ExpirationTime"]
	if !ok {
		return &RepositoryResponse{Success: false, Errors: erro.ErrorGetExpirationTimeSession}
	}

	expirationTime, err := time.Parse(time.RFC3339, expirationTimeString)
	if err != nil {
		return &RepositoryResponse{Success: false, Errors: erro.ErrorSessionParse}
	}

	userID, err := uuid.Parse(userIDString)
	if err != nil {
		return &RepositoryResponse{Success: false, Errors: erro.ErrorSessionParse}
	}
	if ctx.Err() != nil {
		return &RepositoryResponse{Success: false, Errors: erro.ErrorContextTimeOut}
	}
	redisrepo.logger.Info("Successful session receiving")
	return &RepositoryResponse{Success: true, UserID: userID.String(), ExpirationTime: expirationTime}
}
func (redisrepo *AuthRedis) DeleteSession(ctx context.Context, sessionID string) *RepositoryResponse {
	if ctx.Err() != nil {
		return &RepositoryResponse{Success: false, Errors: erro.ErrorContextTimeOut}
	}
	err := redisrepo.Client.Del(ctx, sessionID).Err()
	if err != nil {
		return &RepositoryResponse{Success: false, Errors: erro.ErrorInternalServer}
	}
	if ctx.Err() != nil {
		return &RepositoryResponse{Success: false, Errors: erro.ErrorContextTimeOut}
	}
	redisrepo.logger.Info("Successful session deleted")
	return &RepositoryResponse{Success: true}
}
func NewAuthRedis(client *redis.Client, log *logger.SessionLogger) *AuthRedis {
	return &AuthRedis{Client: client, logger: log}
}
