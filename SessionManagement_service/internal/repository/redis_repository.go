package repository

import (
	"context"
	"log"
	"time"

	"github.com/niktin06sash/MicroserviceProject/SessionManagement_service/internal/erro"
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
}

func (redisrepo *AuthRedis) SetSession(ctx context.Context, session model.Session) *RepositoryResponse {
	if ctx.Err() != nil {
		log.Printf("Context Error:%v", ctx.Err())
		return &RepositoryResponse{Success: false, Errors: erro.ErrorContextTimeOut}
	}
	err := redisrepo.Client.HSet(ctx, session.SessionID, map[string]interface{}{
		"UserID":         session.UserID,
		"ExpirationTime": session.ExpirationTime.Format(time.RFC3339),
	}).Err()

	if err != nil {
		log.Printf("Hset error: %v", err)
		return &RepositoryResponse{Success: false, Errors: erro.ErrorSetSession}
	}
	expiration := time.Until(session.ExpirationTime)
	err = redisrepo.Client.Expire(ctx, session.SessionID, expiration).Err()
	if err != nil {
		log.Printf("Expire error: %v", err)
		return &RepositoryResponse{Success: false, Errors: erro.ErrorSetSession}
	}
	if ctx.Err() != nil {
		log.Printf("Context Error:%v", ctx.Err())
		return &RepositoryResponse{Success: false, Errors: erro.ErrorContextTimeOut}
	}
	log.Printf("Successful session id = %v installation!", session)
	return &RepositoryResponse{Success: true, SessionId: session.SessionID, ExpirationTime: session.ExpirationTime}
}

func (redisrepo *AuthRedis) GetSession(ctx context.Context, sessionID string) *RepositoryResponse {
	if ctx.Err() != nil {
		log.Printf("Context Error:%v", ctx.Err())
		return &RepositoryResponse{Success: false, Errors: erro.ErrorContextTimeOut}
	}
	result, err := redisrepo.Client.HGetAll(ctx, sessionID).Result()
	if err != nil {
		log.Printf("HGetAll error:%v", err)
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
		log.Printf("Time-parse error: %v", err)
		return &RepositoryResponse{Success: false, Errors: erro.ErrorSessionParse}
	}

	userID, err := uuid.Parse(userIDString)
	if err != nil {
		log.Printf("UUID-parse error: %v", err)
		return &RepositoryResponse{Success: false, Errors: erro.ErrorSessionParse}
	}
	if ctx.Err() != nil {
		log.Printf("Context Error:%v", ctx.Err())
		return &RepositoryResponse{Success: false, Errors: erro.ErrorContextTimeOut}
	}
	log.Printf("Successful session id = %v receiving!", sessionID)
	return &RepositoryResponse{Success: true, UserID: userID.String(), ExpirationTime: expirationTime}
}
func (redisrepo *AuthRedis) DeleteSession(ctx context.Context, sessionID string) *RepositoryResponse {
	if ctx.Err() != nil {
		log.Printf("Context Error:%v", ctx.Err())
		return &RepositoryResponse{Success: false, Errors: erro.ErrorContextTimeOut}
	}
	err := redisrepo.Client.Del(ctx, sessionID).Err()
	if err != nil {
		log.Printf("Error deleting session %s: %v", sessionID, err)
		return &RepositoryResponse{Success: false, Errors: erro.ErrorInternalServer}
	}
	if ctx.Err() != nil {
		log.Printf("Context Error:%v", ctx.Err())
		return &RepositoryResponse{Success: false, Errors: erro.ErrorContextTimeOut}
	}
	log.Printf("Session %s deleted successfully", sessionID)
	return &RepositoryResponse{Success: true}
}
func NewAuthRedis(client *redis.Client) *AuthRedis {
	return &AuthRedis{Client: client}
}
