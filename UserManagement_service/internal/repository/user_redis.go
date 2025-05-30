package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/erro"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/kafka"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/metrics"
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
	start := time.Now()
	defer CacheMetrics(place, start)
	key := fmt.Sprintf("user:%s:profile", id)
	traceID := ctx.Value("traceID").(string)
	err := redisrepo.Client.RedisClient.HSet(ctx, key, map[string]interface{}{
		"UserID":    data[KeyUserID],
		"UserEmail": data[KeyUserEmail],
		"UserName":  data[KeyUserName],
	}).Err()
	if err != nil {
		metrics.UserDBErrorsTotal.WithLabelValues("HSET").Inc()
		redisrepo.KafkaProducer.NewUserLog(kafka.LogLevelError, place, traceID, fmt.Sprintf("Hset profiles-cache error: %v", err))
		return &RepositoryResponse{Success: false, Errors: &erro.ErrorResponse{Message: erro.UserServiceUnavalaible, Type: erro.ServerErrorType}}
	}
	err = redisrepo.Client.RedisClient.Expire(ctx, key, time.Until(time.Now().Add(1*time.Hour))).Err()
	if err != nil {
		metrics.UserDBErrorsTotal.WithLabelValues("EXPIRE").Inc()
		redisrepo.KafkaProducer.NewUserLog(kafka.LogLevelError, place, traceID, fmt.Sprintf("Expire profiles-cache error: %v", err))
		return &RepositoryResponse{Success: false, Errors: &erro.ErrorResponse{Message: erro.UserServiceUnavalaible, Type: erro.ServerErrorType}}
	}
	redisrepo.KafkaProducer.NewUserLog(kafka.LogLevelInfo, place, traceID, "Successful profiles-cache installation")
	return &RepositoryResponse{Success: true}
}
func (redisrepo *UserRedisRepo) DeleteProfileCache(ctx context.Context, id string) *RepositoryResponse {
	const place = DeleteProfileCache
	start := time.Now()
	defer CacheMetrics(place, start)
	key := fmt.Sprintf("user:%s:profile", id)
	traceID := ctx.Value("traceID").(string)
	num, err := redisrepo.Client.RedisClient.Del(ctx, key).Result()
	if err != nil {
		metrics.UserDBErrorsTotal.WithLabelValues("DEL").Inc()
		redisrepo.KafkaProducer.NewUserLog(kafka.LogLevelError, place, traceID, fmt.Sprintf("Del profiles-cache error: %v", err))
		return &RepositoryResponse{Success: false, Errors: &erro.ErrorResponse{Message: erro.UserServiceUnavalaible, Type: erro.ServerErrorType}}
	}
	if num == 0 {
		return &RepositoryResponse{
			Success: false,
		}
	}
	redisrepo.KafkaProducer.NewUserLog(kafka.LogLevelInfo, place, traceID, "Successful profiles-cache deleted")
	return &RepositoryResponse{Success: true}
}
func (redisrepo *UserRedisRepo) GetProfileCache(ctx context.Context, id string) *RepositoryResponse {
	const place = GetProfileCache
	start := time.Now()
	defer CacheMetrics(place, start)
	key := fmt.Sprintf("user:%s:profile", id)
	traceID := ctx.Value("traceID").(string)
	result, err := redisrepo.Client.RedisClient.HGetAll(ctx, key).Result()
	if err != nil {
		metrics.UserDBErrorsTotal.WithLabelValues("HGETALL").Inc()
		redisrepo.KafkaProducer.NewUserLog(kafka.LogLevelError, place, traceID, fmt.Sprintf("HGetAll profiles-cache error: %v", err))
		return &RepositoryResponse{Success: false, Errors: &erro.ErrorResponse{Message: erro.UserServiceUnavalaible, Type: erro.ServerErrorType}}
	}
	if len(result) == 0 {
		return &RepositoryResponse{Success: false}
	}
	userIDstr, ok := result["UserID"]
	if !ok {
		metrics.UserDBErrorsTotal.WithLabelValues("GETMAP").Inc()
		redisrepo.KafkaProducer.NewUserLog(kafka.LogLevelError, place, traceID, "Get UserID from profiles-cache error")
		return &RepositoryResponse{Success: false, Errors: &erro.ErrorResponse{Message: erro.UserServiceUnavalaible, Type: erro.ServerErrorType}}
	}
	userEmail, ok := result["UserEmail"]
	if !ok {
		metrics.UserDBErrorsTotal.WithLabelValues("GETMAP").Inc()
		redisrepo.KafkaProducer.NewUserLog(kafka.LogLevelError, place, traceID, "Get UserEmail from profiles-cache error")
		return &RepositoryResponse{Success: false, Errors: &erro.ErrorResponse{Message: erro.UserServiceUnavalaible, Type: erro.ServerErrorType}}
	}
	userName, ok := result["UserName"]
	if !ok {
		metrics.UserDBErrorsTotal.WithLabelValues("GETMAP").Inc()
		redisrepo.KafkaProducer.NewUserLog(kafka.LogLevelError, place, traceID, "Get UserName from profiles-cache error")
		return &RepositoryResponse{Success: false, Errors: &erro.ErrorResponse{Message: erro.UserServiceUnavalaible, Type: erro.ServerErrorType}}
	}
	redisrepo.KafkaProducer.NewUserLog(kafka.LogLevelInfo, place, traceID, "Successful profiles-cache got")
	return &RepositoryResponse{Success: true, Data: map[string]any{KeyUserID: userIDstr, KeyUserEmail: userEmail, KeyUserName: userName}}
}
