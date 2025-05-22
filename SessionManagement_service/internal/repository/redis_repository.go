package repository

import (
	"context"
	"fmt"
	"time"

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

func (redisrepo *AuthRedis) validateContext(ctx context.Context, place string) (string, error) {
	traceID := ctx.Value("traceID").(string)
	select {
	case <-ctx.Done():
		fmterr := fmt.Sprintf("Context error: %v", ctx.Err())
		redisrepo.KafkaProducer.NewSessionLog(kafka.LogLevelError, place, traceID, fmterr)
		return "", status.Errorf(codes.Internal, "Request timed out")
	default:
		return traceID, nil
	}
}
func (redisrepo *AuthRedis) SetSession(ctx context.Context, session model.Session) *RepositoryResponse {
	var place = "Repository-SetSession"
	traceID, err := redisrepo.validateContext(ctx, place)
	if err != nil {
		return &RepositoryResponse{Success: false, Errors: err}
	}
	err = redisrepo.Client.RedisClient.HSet(ctx, session.SessionID, map[string]interface{}{
		"UserID":         session.UserID,
		"ExpirationTime": session.ExpirationTime.Format(time.RFC3339),
	}).Err()
	if err != nil {
		fmterr := fmt.Sprintf("Hset session error: %v", err)
		redisrepo.KafkaProducer.NewSessionLog(kafka.LogLevelError, place, traceID, fmterr)
		return &RepositoryResponse{Success: false, Errors: status.Errorf(codes.Internal, "Session-Service is unavailable")}
	}
	expiration := time.Until(session.ExpirationTime)
	err = redisrepo.Client.RedisClient.Expire(ctx, session.SessionID, expiration).Err()
	if err != nil {
		fmterr := fmt.Sprintf("Expire session error: %v", err)
		redisrepo.KafkaProducer.NewSessionLog(kafka.LogLevelError, place, traceID, fmterr)
		return &RepositoryResponse{Success: false, Errors: status.Errorf(codes.Internal, "Session-Service is unavailable")}
	}
	redisrepo.KafkaProducer.NewSessionLog(kafka.LogLevelInfo, place, traceID, "Successful session installation")
	return &RepositoryResponse{Success: true, SessionId: session.SessionID, ExpirationTime: session.ExpirationTime}
}

func (redisrepo *AuthRedis) GetSession(ctx context.Context, sessionID string) *RepositoryResponse {
	var place = "Repository-GetSession"
	traceID, err := redisrepo.validateContext(ctx, place)
	if err != nil {
		return &RepositoryResponse{Success: false, Errors: err}
	}
	flag := ctx.Value("flagvalidate").(string)
	flagValidate := flag == "true"
	return redisrepo.getSessionData(ctx, sessionID, traceID, place, flagValidate)
}
func (redisrepo *AuthRedis) getSessionData(ctx context.Context, sessionID string, traceID string, place string, flagvalidate bool) *RepositoryResponse {
	result, err := redisrepo.Client.RedisClient.HGetAll(ctx, sessionID).Result()
	if err != nil {
		if err == redis.Nil {
			if flagvalidate {
				redisrepo.KafkaProducer.NewSessionLog(kafka.LogLevelWarn, place, traceID, "Session not found")
				return &RepositoryResponse{Success: false, Errors: status.Errorf(codes.InvalidArgument, "Session not found")}
			}
			redisrepo.KafkaProducer.NewSessionLog(kafka.LogLevelInfo, place, traceID, "Successful verification of missing session")
			return &RepositoryResponse{Success: true}
		}
		fmterr := fmt.Sprintf("HGetAll session error: %v", err)
		redisrepo.KafkaProducer.NewSessionLog(kafka.LogLevelError, place, traceID, fmterr)
		return &RepositoryResponse{Success: false, Errors: status.Errorf(codes.Internal, "Session-Service is unavailable")}
	}
	if len(result) == 0 {
		if flagvalidate {
			redisrepo.KafkaProducer.NewSessionLog(kafka.LogLevelWarn, place, traceID, "Session is empty or invalid")
			return &RepositoryResponse{Success: false, Errors: status.Errorf(codes.InvalidArgument, "Session is empty or invalid")}
		}
		redisrepo.KafkaProducer.NewSessionLog(kafka.LogLevelInfo, place, traceID, "Successful verification of missing session")
		return &RepositoryResponse{Success: true}
	}
	if !flagvalidate {
		redisrepo.KafkaProducer.NewSessionLog(kafka.LogLevelWarn, place, traceID, "Authorized-request for unauthorized users")
		return &RepositoryResponse{Success: false, Errors: status.Errorf(codes.InvalidArgument, "Already Authorized")}
	}
	userIDString, ok := result["UserID"]
	if !ok {
		redisrepo.KafkaProducer.NewSessionLog(kafka.LogLevelError, place, traceID, "Get UserID from session error")
		return &RepositoryResponse{Success: false, Errors: status.Errorf(codes.InvalidArgument, "Session not found")}
	}
	expirationTimeString, ok := result["ExpirationTime"]
	if !ok {
		redisrepo.KafkaProducer.NewSessionLog(kafka.LogLevelError, place, traceID, "Get ExpirationTime from session error")
		return &RepositoryResponse{Success: false, Errors: status.Errorf(codes.InvalidArgument, "Session not found")}
	}
	expirationTime, err := time.Parse(time.RFC3339, expirationTimeString)
	if err != nil {
		fmterr := fmt.Sprintf("Time-parse error: %v", err)
		redisrepo.KafkaProducer.NewSessionLog(kafka.LogLevelError, place, traceID, fmterr)
		return &RepositoryResponse{Success: false, Errors: status.Errorf(codes.Internal, "Session-Service is unavailable")}
	}
	userID, err := uuid.Parse(userIDString)
	if err != nil {
		fmterr := fmt.Sprintf("UUID-parse error: %v", err)
		redisrepo.KafkaProducer.NewSessionLog(kafka.LogLevelError, place, traceID, fmterr)
		return &RepositoryResponse{Success: false, Errors: status.Errorf(codes.Internal, "Session-Service is unavailable")}
	}
	redisrepo.KafkaProducer.NewSessionLog(kafka.LogLevelInfo, place, traceID, "Successful session receiving")
	return &RepositoryResponse{Success: true, UserID: userID.String(), ExpirationTime: expirationTime}
}
func (redisrepo *AuthRedis) DeleteSession(ctx context.Context, sessionID string) *RepositoryResponse {
	var place = "Repository-DeleteSession"
	traceID, err := redisrepo.validateContext(ctx, place)
	if err != nil {
		return &RepositoryResponse{Success: false, Errors: err}
	}
	err = redisrepo.Client.RedisClient.Del(ctx, sessionID).Err()
	if err != nil {
		if err == redis.Nil {
			redisrepo.KafkaProducer.NewSessionLog(kafka.LogLevelWarn, place, traceID, "Session not found")
			return &RepositoryResponse{
				Success: false,
				Errors:  status.Errorf(codes.InvalidArgument, "Session not found"),
			}
		}
		fmterr := fmt.Sprintf("Del session error: %v", err)
		redisrepo.KafkaProducer.NewSessionLog(kafka.LogLevelError, place, traceID, fmterr)
		return &RepositoryResponse{Success: false, Errors: status.Errorf(codes.Internal, "Session-Service is unavailable")}
	}
	redisrepo.KafkaProducer.NewSessionLog(kafka.LogLevelInfo, place, traceID, "Successful session deleted")
	return &RepositoryResponse{Success: true}
}
