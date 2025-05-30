package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/erro"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/kafka"
)

type UserRedisRepo struct {
	Client        *RedisObject
	KafkaProducer kafka.KafkaProducerService
}

func NewUserRedisRepo(red *RedisObject, kafkaprod kafka.KafkaProducerService) *UserRedisRepo {
	return &UserRedisRepo{Client: red, KafkaProducer: kafkaprod}
}

func (redisrepo *UserRedisRepo) AddProfileCache(ctx context.Context, id string, data map[string]any) *RepositoryResponse {
	const place = AddProfileCache
	key := fmt.Sprintf("user:%s:profile", id)
	traceID := ctx.Value("traceID").(string)
	err := redisrepo.Client.RedisClient.HSet(ctx, key, map[string]interface{}{
		"UserID":    data[KeyUserID],
		"UserEmail": data[KeyUserEmail],
		"UserName":  data[KeyUserName],
	}).Err()
	if err != nil {
		fmterr := fmt.Sprintf("Hset profile-cache error: %v", err)
		redisrepo.KafkaProducer.NewUserLog(kafka.LogLevelError, place, traceID, fmterr)
		return &RepositoryResponse{Success: false, Errors: &erro.ErrorResponse{Message: erro.UserServiceUnavalaible, Type: erro.ServerErrorType}}
	}
	err = redisrepo.Client.RedisClient.Expire(ctx, key, time.Until(time.Now().Add(1*time.Hour))).Err()
	if err != nil {
		fmterr := fmt.Sprintf("Expire profile-cache error: %v", err)
		redisrepo.KafkaProducer.NewUserLog(kafka.LogLevelError, place, traceID, fmterr)
		return &RepositoryResponse{Success: false, Errors: &erro.ErrorResponse{Message: erro.UserServiceUnavalaible, Type: erro.ServerErrorType}}
	}
	redisrepo.KafkaProducer.NewUserLog(kafka.LogLevelInfo, place, traceID, "Successful profile-cache installation")
	return &RepositoryResponse{Success: true}
}
func (redisrepo *UserRedisRepo) DeleteProfileCache(ctx context.Context, id string) *RepositoryResponse {
	const place = DeleteProfileCache
	key := fmt.Sprintf("user:%s:profile", id)
	traceID := ctx.Value("traceID").(string)
	num, err := redisrepo.Client.RedisClient.Del(ctx, key).Result()
	if err != nil {
		fmterr := fmt.Sprintf("Del profile-cache error: %v", err)
		redisrepo.KafkaProducer.NewUserLog(kafka.LogLevelError, place, traceID, fmterr)
		return &RepositoryResponse{Success: false, Errors: &erro.ErrorResponse{Message: erro.UserServiceUnavalaible, Type: erro.ServerErrorType}}
	}
	if num == 0 {
		return &RepositoryResponse{
			Success: false,
		}
	}
	redisrepo.KafkaProducer.NewUserLog(kafka.LogLevelInfo, place, traceID, "Successful profile-cache deleted")
	return &RepositoryResponse{Success: true}
}
func (redisrepo *UserRedisRepo) GetProfileCache(ctx context.Context, id string) *RepositoryResponse {
	const place = GetProfileCache
	key := fmt.Sprintf("user:%s:profile", id)
	traceID := ctx.Value("traceID").(string)
	result, err := redisrepo.Client.RedisClient.HGetAll(ctx, key).Result()
	if err != nil {
		fmterr := fmt.Sprintf("HGetAll profile-cache error: %v", err)
		redisrepo.KafkaProducer.NewUserLog(kafka.LogLevelError, place, traceID, fmterr)
		return &RepositoryResponse{Success: false, Errors: &erro.ErrorResponse{Message: erro.UserServiceUnavalaible, Type: erro.ServerErrorType}}
	}
	if len(result) == 0 {
		return &RepositoryResponse{Success: false}
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
