package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/niktin06sash/MicroserviceProject/SessionManagement_service/internal/erro"
	"github.com/niktin06sash/MicroserviceProject/SessionManagement_service/internal/kafka"
	"github.com/niktin06sash/MicroserviceProject/SessionManagement_service/internal/model"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type AuthRedis struct {
	Client        *RedisObject
	KafkaProducer kafka.KafkaProducerService
}

func NewAuthRedis(client *RedisObject, kafka kafka.KafkaProducerService) *AuthRedis {
	return &AuthRedis{Client: client, KafkaProducer: kafka}
}

func (redisrepo *AuthRedis) SetSession(ctx context.Context, session model.Session) *RepositoryResponse {
	const place = SetSession
	traceID := ctx.Value("traceID").(string)
	err := redisrepo.Client.RedisClient.HSet(ctx, session.SessionID, map[string]interface{}{
		"UserID":         session.UserID,
		"ExpirationTime": session.ExpirationTime.Format(time.RFC3339),
	}).Err()
	if err != nil {
		fmterr := fmt.Sprintf("Hset session error: %v", err)
		redisrepo.KafkaProducer.NewSessionLog(kafka.LogLevelError, place, traceID, fmterr)
		return &RepositoryResponse{Success: false, Errors: status.Errorf(codes.Internal, erro.SessionServiceUnavalaible)}
	}
	expiration := time.Until(session.ExpirationTime)
	err = redisrepo.Client.RedisClient.Expire(ctx, session.SessionID, expiration).Err()
	if err != nil {
		fmterr := fmt.Sprintf("Expire session error: %v", err)
		redisrepo.KafkaProducer.NewSessionLog(kafka.LogLevelError, place, traceID, fmterr)
		return &RepositoryResponse{Success: false, Errors: status.Errorf(codes.Internal, erro.SessionServiceUnavalaible)}
	}
	redisrepo.KafkaProducer.NewSessionLog(kafka.LogLevelInfo, place, traceID, "Successful session installation")
	return &RepositoryResponse{Success: true, SessionId: session.SessionID, ExpirationTime: session.ExpirationTime}
}

func (redisrepo *AuthRedis) GetSession(ctx context.Context, sessionID string) *RepositoryResponse {
	const place = GetSession
	traceID := ctx.Value("traceID").(string)
	flag := ctx.Value("flagvalidate").(string)
	flagValidate := flag == "true"
	return redisrepo.getSessionData(ctx, sessionID, traceID, place, flagValidate)
}
func (redisrepo *AuthRedis) getSessionData(ctx context.Context, sessionID string, traceID string, place string, flagvalidate bool) *RepositoryResponse {
	result, err := redisrepo.Client.RedisClient.HGetAll(ctx, sessionID).Result()
	if err != nil {
		if err == redis.Nil {
			if flagvalidate {
				redisrepo.KafkaProducer.NewSessionLog(kafka.LogLevelWarn, place, traceID, erro.SessionNotFound)
				return &RepositoryResponse{Success: false, Errors: status.Errorf(codes.InvalidArgument, erro.SessionIdRequired)}
			}
			redisrepo.KafkaProducer.NewSessionLog(kafka.LogLevelInfo, place, traceID, "Successful verification of missing session")
			return &RepositoryResponse{Success: true}
		}
		fmterr := fmt.Sprintf("HGetAll session error: %v", err)
		redisrepo.KafkaProducer.NewSessionLog(kafka.LogLevelError, place, traceID, fmterr)
		return &RepositoryResponse{Success: false, Errors: status.Errorf(codes.Internal, erro.SessionServiceUnavalaible)}
	}
	if len(result) == 0 {
		if flagvalidate {
			redisrepo.KafkaProducer.NewSessionLog(kafka.LogLevelWarn, place, traceID, "Session is empty or invalid")
			return &RepositoryResponse{Success: false, Errors: status.Errorf(codes.InvalidArgument, erro.SessionNotFound)}
		}
		redisrepo.KafkaProducer.NewSessionLog(kafka.LogLevelInfo, place, traceID, "Successful verification of missing session")
		return &RepositoryResponse{Success: true}
	}
	if !flagvalidate {
		redisrepo.KafkaProducer.NewSessionLog(kafka.LogLevelWarn, place, traceID, "Authorized-request for unauthorized users")
		return &RepositoryResponse{Success: false, Errors: status.Errorf(codes.InvalidArgument, erro.AlreadyAuthorized)}
	}
	userIDString, ok := result["UserID"]
	if !ok {
		redisrepo.KafkaProducer.NewSessionLog(kafka.LogLevelError, place, traceID, "Get UserID from session error")
		return &RepositoryResponse{Success: false, Errors: status.Errorf(codes.Internal, erro.SessionServiceUnavalaible)}
	}
	expirationTimeString, ok := result["ExpirationTime"]
	if !ok {
		redisrepo.KafkaProducer.NewSessionLog(kafka.LogLevelError, place, traceID, "Get ExpirationTime from session error")
		return &RepositoryResponse{Success: false, Errors: status.Errorf(codes.Internal, erro.SessionServiceUnavalaible)}
	}
	expirationTime, err := time.Parse(time.RFC3339, expirationTimeString)
	if err != nil {
		fmterr := fmt.Sprintf("Time-parse error: %v", err)
		redisrepo.KafkaProducer.NewSessionLog(kafka.LogLevelError, place, traceID, fmterr)
		return &RepositoryResponse{Success: false, Errors: status.Errorf(codes.Internal, erro.SessionServiceUnavalaible)}
	}
	userID, err := uuid.Parse(userIDString)
	if err != nil {
		fmterr := fmt.Sprintf("UUID-parse error: %v", err)
		redisrepo.KafkaProducer.NewSessionLog(kafka.LogLevelError, place, traceID, fmterr)
		return &RepositoryResponse{Success: false, Errors: status.Errorf(codes.Internal, erro.SessionServiceUnavalaible)}
	}
	redisrepo.KafkaProducer.NewSessionLog(kafka.LogLevelInfo, place, traceID, "Successful session receiving")
	return &RepositoryResponse{Success: true, UserID: userID.String(), ExpirationTime: expirationTime}
}
func (redisrepo *AuthRedis) DeleteSession(ctx context.Context, sessionID string) *RepositoryResponse {
	const place = DeleteSession
	traceID := ctx.Value("traceID").(string)
	err := redisrepo.Client.RedisClient.Del(ctx, sessionID).Err()
	if err != nil {
		if err == redis.Nil {
			redisrepo.KafkaProducer.NewSessionLog(kafka.LogLevelWarn, place, traceID, "Session not found")
			return &RepositoryResponse{
				Success: false,
				Errors:  status.Errorf(codes.InvalidArgument, erro.SessionNotFound),
			}
		}
		fmterr := fmt.Sprintf("Del session error: %v", err)
		redisrepo.KafkaProducer.NewSessionLog(kafka.LogLevelError, place, traceID, fmterr)
		return &RepositoryResponse{Success: false, Errors: status.Errorf(codes.Internal, erro.SessionServiceUnavalaible)}
	}
	redisrepo.KafkaProducer.NewSessionLog(kafka.LogLevelInfo, place, traceID, "Successful session deleted")
	return &RepositoryResponse{Success: true}
}
