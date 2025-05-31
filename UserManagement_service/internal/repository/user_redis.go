package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/erro"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/metrics"
)

type UserRedisRepo struct {
	Client *RedisObject
}

func NewUserRedisRepo(red *RedisObject) *UserRedisRepo {
	return &UserRedisRepo{Client: red}
}

func (redisrepo *UserRedisRepo) AddProfileCache(ctx context.Context, id string, data map[string]any) *RepositoryResponse {
	const place = AddProfileCache
	start := time.Now()
	defer CacheMetrics(place, start)
	key := fmt.Sprintf("user:%s:profile", id)
	err := redisrepo.Client.RedisClient.HSet(ctx, key, map[string]interface{}{
		"UserID":    data[KeyUserID],
		"UserEmail": data[KeyUserEmail],
		"UserName":  data[KeyUserName],
	}).Err()
	if err != nil {
		metrics.UserDBErrorsTotal.WithLabelValues("HSET").Inc()
		return &RepositoryResponse{Success: false, Errors: &ErrorResponse{Message: fmt.Sprintf("Hset profiles-cache error: %v", err), Type: erro.ServerErrorType, Place: place}}
	}
	err = redisrepo.Client.RedisClient.Expire(ctx, key, time.Until(time.Now().Add(1*time.Hour))).Err()
	if err != nil {
		metrics.UserDBErrorsTotal.WithLabelValues("EXPIRE").Inc()
		return &RepositoryResponse{Success: false, Errors: &ErrorResponse{Message: fmt.Sprintf("Expire profiles-cache error: %v", err), Type: erro.ServerErrorType, Place: place}}
	}
	return &RepositoryResponse{Success: true}
}
func (redisrepo *UserRedisRepo) DeleteProfileCache(ctx context.Context, id string) *RepositoryResponse {
	const place = DeleteProfileCache
	start := time.Now()
	defer CacheMetrics(place, start)
	key := fmt.Sprintf("user:%s:profile", id)
	num, err := redisrepo.Client.RedisClient.Del(ctx, key).Result()
	if err != nil {
		metrics.UserDBErrorsTotal.WithLabelValues("DEL").Inc()
		return &RepositoryResponse{Success: false, Errors: &ErrorResponse{Message: fmt.Sprintf("Del profiles-cache error: %v", err), Type: erro.ServerErrorType, Place: place}}
	}
	if num == 0 {
		return &RepositoryResponse{
			Success: false,
		}
	}
	return &RepositoryResponse{Success: true}
}
func (redisrepo *UserRedisRepo) GetProfileCache(ctx context.Context, id string) *RepositoryResponse {
	const place = GetProfileCache
	start := time.Now()
	defer CacheMetrics(place, start)
	key := fmt.Sprintf("user:%s:profile", id)
	result, err := redisrepo.Client.RedisClient.HGetAll(ctx, key).Result()
	if err != nil {
		metrics.UserDBErrorsTotal.WithLabelValues("HGETALL").Inc()
		return &RepositoryResponse{Success: false, Errors: &ErrorResponse{Message: fmt.Sprintf("HGetAll profiles-cache error: %v", err), Type: erro.ServerErrorType, Place: place}}
	}
	if len(result) == 0 {
		return &RepositoryResponse{Success: false}
	}
	userIDstr := result["UserID"]
	userEmail := result["UserEmail"]
	userName := result["UserName"]
	return &RepositoryResponse{Success: true, Data: map[string]any{KeyUserID: userIDstr, KeyUserEmail: userEmail, KeyUserName: userName}}
}
