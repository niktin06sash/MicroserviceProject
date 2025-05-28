package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/erro"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/kafka"
	"github.com/redis/go-redis/v9"
)

type AuthRedisRepo struct {
	Client        *RedisObject
	KafkaProducer kafka.KafkaProducerService
}

func NewAuthRedisRepo(red *RedisObject, kafkaprod kafka.KafkaProducerService) *AuthRedisRepo {
	return &AuthRedisRepo{Client: red, KafkaProducer: kafkaprod}
}

func (redis *AuthRedisRepo) AddProfileCache(ctx context.Context, id string, data map[string]any) *RepositoryResponse {
	const place = AddProfileCache
	traceID := ctx.Value("traceID").(string)
	err := redis.Client.RedisClient.HSet(ctx, id, map[string]interface{}{
		"UserID":    data[KeyUserID],
		"UserEmail": data[KeyUserEmail],
		"UserName":  data[KeyUserName],
	}).Err()
	if err != nil {
		fmterr := fmt.Sprintf("Hset profile-cache error: %v", err)
		redis.KafkaProducer.NewUserLog(kafka.LogLevelError, place, traceID, fmterr)
		return &RepositoryResponse{Success: false, Errors: &erro.ErrorResponse{Message: erro.UserServiceUnavalaible, Type: erro.ServerErrorType}}
	}
	err = redis.Client.RedisClient.Expire(ctx, id, time.Until(time.Now().Add(1*time.Hour))).Err()
	if err != nil {
		fmterr := fmt.Sprintf("Expire profile-cache error: %v", err)
		redis.KafkaProducer.NewUserLog(kafka.LogLevelError, place, traceID, fmterr)
		return &RepositoryResponse{Success: false, Errors: &erro.ErrorResponse{Message: erro.UserServiceUnavalaible, Type: erro.ServerErrorType}}
	}
	redis.KafkaProducer.NewUserLog(kafka.LogLevelInfo, place, traceID, "Successful profile-cache installation")
	return &RepositoryResponse{Success: true}
}
func (redisrepo *AuthRedisRepo) DeleteProfileCache(ctx context.Context, id string) *RepositoryResponse {
	const place = DeleteProfileCache
	traceID := ctx.Value("traceID").(string)
	err := redisrepo.Client.RedisClient.Del(ctx, id).Err()
	if err != nil {
		if err == redis.Nil {
			return &RepositoryResponse{Success: true}
		}
		fmterr := fmt.Sprintf("Del profile-cache error: %v", err)
		redisrepo.KafkaProducer.NewUserLog(kafka.LogLevelError, place, traceID, fmterr)
		return &RepositoryResponse{Success: false, Errors: &erro.ErrorResponse{Message: erro.UserServiceUnavalaible, Type: erro.ServerErrorType}}
	}
	redisrepo.KafkaProducer.NewUserLog(kafka.LogLevelInfo, place, traceID, "Successful profile-cache deleted")
	return &RepositoryResponse{Success: true}
}
func (redisrepo *AuthRedisRepo) GetProfileCache(ctx context.Context, id string) *RepositoryResponse {
	const place = GetProfileCache
	traceID := ctx.Value("traceID").(string)
	result, err := redisrepo.Client.RedisClient.HGetAll(ctx, id).Result()
	if err != nil {
		if err == redis.Nil {
			return &RepositoryResponse{Success: true}
		}
		fmterr := fmt.Sprintf("HGetAll profile-cache error: %v", err)
		redisrepo.KafkaProducer.NewUserLog(kafka.LogLevelError, place, traceID, fmterr)
		return &RepositoryResponse{Success: false, Errors: &erro.ErrorResponse{Message: erro.UserServiceUnavalaible, Type: erro.ServerErrorType}}
	}
	userIDstr, ok := result["UserID"]
	if !ok {
		redisrepo.KafkaProducer.NewUserLog(kafka.LogLevelError, place, traceID, "Get UserID from profile-cache error")
		return &RepositoryResponse{Success: false, Errors: &erro.ErrorResponse{Message: erro.UserServiceUnavalaible, Type: erro.ServerErrorType}}
	}
	userEmail, ok := result["UserEmail"]
	if !ok {
		redisrepo.KafkaProducer.NewUserLog(kafka.LogLevelError, place, traceID, "Get UserEmail from profile-cache error")
		return &RepositoryResponse{Success: false, Errors: &erro.ErrorResponse{Message: erro.UserServiceUnavalaible, Type: erro.ServerErrorType}}
	}
	userName, ok := result["UserName"]
	if !ok {
		redisrepo.KafkaProducer.NewUserLog(kafka.LogLevelError, place, traceID, "Get UserName from profile-cache error")
		return &RepositoryResponse{Success: false, Errors: &erro.ErrorResponse{Message: erro.UserServiceUnavalaible, Type: erro.ServerErrorType}}
	}
	redisrepo.KafkaProducer.NewUserLog(kafka.LogLevelInfo, place, traceID, "Successful profile-cache got")
	return &RepositoryResponse{Success: true, Data: map[string]any{KeyUserID: userIDstr, KeyUserEmail: userEmail, KeyUserName: userName}}
}
