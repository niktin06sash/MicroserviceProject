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
func CacheMetrics(place string, start time.Time) {
	metrics.UserCacheQueriesTotal.WithLabelValues(place).Inc()
	duration := time.Since(start).Seconds()
	metrics.UserCacheQueryDuration.WithLabelValues(place).Observe(duration)
}
func (redisrepo *UserRedisRepo) AddProfileCache(ctx context.Context, id string, data map[string]any) *RepositoryResponse {
	const place = AddProfileCache
	start := time.Now()
	defer CacheMetrics(place, start)
	err := redisrepo.Client.RedisClient.HSet(ctx, id, map[string]interface{}{
		KeyUserID:    data[KeyUserID],
		KeyUserEmail: data[KeyUserEmail],
		KeyUserName:  data[KeyUserName],
	}).Err()
	if err != nil {
		metrics.UserCacheErrorsTotal.WithLabelValues("HSET").Inc()
		return &RepositoryResponse{Success: false, Errors: map[string]string{erro.ErrorType: erro.ServerErrorType, erro.ErrorMessage: fmt.Sprintf("Hset profiles-cache error: %v", err)}, Place: place}
	}
	err = redisrepo.Client.RedisClient.Expire(ctx, id, time.Until(time.Now().Add(1*time.Hour))).Err()
	if err != nil {
		metrics.UserCacheErrorsTotal.WithLabelValues("EXPIRE").Inc()
		return &RepositoryResponse{Success: false, Errors: map[string]string{erro.ErrorType: erro.ServerErrorType, erro.ErrorMessage: fmt.Sprintf("Expire profiles-cache error: %v", err)}, Place: place}
	}
	return &RepositoryResponse{Success: true, SuccessMessage: "Successful add profile in cache", Place: place}
}
func (redisrepo *UserRedisRepo) DeleteProfileCache(ctx context.Context, id string) *RepositoryResponse {
	const place = DeleteProfileCache
	start := time.Now()
	defer CacheMetrics(place, start)
	num, err := redisrepo.Client.RedisClient.Del(ctx, id).Result()
	if err != nil {
		metrics.UserCacheErrorsTotal.WithLabelValues("DEL").Inc()
		return &RepositoryResponse{Success: false, Errors: map[string]string{erro.ErrorType: erro.ServerErrorType, erro.ErrorMessage: fmt.Sprintf("Del profiles-cache error: %v", err)}, Place: place}
	}
	if num == 0 {
		return &RepositoryResponse{
			Success: false,
		}
	}
	return &RepositoryResponse{Success: true, SuccessMessage: "Successful deleted profile from cache", Place: place}
}
func (redisrepo *UserRedisRepo) GetProfileCache(ctx context.Context, id string) *RepositoryResponse {
	const place = GetProfileCache
	start := time.Now()
	defer CacheMetrics(place, start)
	result, err := redisrepo.Client.RedisClient.HGetAll(ctx, id).Result()
	if err != nil {
		metrics.UserCacheErrorsTotal.WithLabelValues("HGETALL").Inc()
		return &RepositoryResponse{Success: false, Errors: map[string]string{erro.ErrorType: erro.ServerErrorType, erro.ErrorMessage: fmt.Sprintf("HGetAll profiles-cache error: %v", err)}, Place: place}
	}
	if len(result) == 0 {
		return &RepositoryResponse{Success: false}
	}
	userIDstr := result[KeyUserID]
	userEmail := result[KeyUserEmail]
	userName := result[KeyUserName]
	return &RepositoryResponse{Success: true, Data: map[string]any{KeyUserID: userIDstr, KeyUserEmail: userEmail, KeyUserName: userName}, SuccessMessage: "Successful get profile from cache", Place: place}
}
