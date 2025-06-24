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
		return &RepositoryResponse{Success: false, Errors: erro.ServerError(fmt.Sprintf(erro.ErrorHsetProfiles, err)), Place: place}
	}
	err = redisrepo.Client.RedisClient.Expire(ctx, id, time.Until(time.Now().Add(1*time.Hour))).Err()
	if err != nil {
		metrics.UserCacheErrorsTotal.WithLabelValues("EXPIRE").Inc()
		return &RepositoryResponse{Success: false, Errors: erro.ServerError(fmt.Sprintf(erro.ErrorExpireProfiles, err)), Place: place}
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
		return &RepositoryResponse{Success: false, Errors: erro.ServerError(fmt.Sprintf(erro.ErrorDelProfiles, err)), Place: place}
	}
	if num == 0 {
		return &RepositoryResponse{Success: false, SuccessMessage: "Profile was not found in the cache", Place: place}
	}
	return &RepositoryResponse{Success: true, SuccessMessage: "Successful delete profile from cache", Place: place}
}
func (redisrepo *UserRedisRepo) GetProfileCache(ctx context.Context, id string) *RepositoryResponse {
	const place = GetProfileCache
	start := time.Now()
	defer CacheMetrics(place, start)
	result, err := redisrepo.Client.RedisClient.HGetAll(ctx, id).Result()
	if err != nil {
		metrics.UserCacheErrorsTotal.WithLabelValues("HGETALL").Inc()
		return &RepositoryResponse{Success: false, Errors: erro.ServerError(fmt.Sprintf(erro.ErrorHgetAllProfiles, err)), Place: place}
	}
	if len(result) == 0 {
		return &RepositoryResponse{Success: false, SuccessMessage: "Profile was not found in the cache", Place: place}
	}
	userIDstr := result[KeyUserID]
	userEmail := result[KeyUserEmail]
	userName := result[KeyUserName]
	return &RepositoryResponse{Success: true, Data: map[string]any{KeyUserID: userIDstr, KeyUserEmail: userEmail, KeyUserName: userName}, SuccessMessage: "Successful get profile from cache", Place: place}
}
