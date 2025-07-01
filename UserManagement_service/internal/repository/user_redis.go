package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/erro"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/metrics"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/model"
	"github.com/redis/go-redis/v9"
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
func (redisrepo *UserRedisRepo) AddProfileCache(ctx context.Context, user *model.User) *RepositoryResponse {
	const place = AddProfileCache
	start := time.Now()
	jsondata, err := json.Marshal(user)
	if err != nil {
		return &RepositoryResponse{Success: false, Errors: erro.ServerError(fmt.Sprintf(erro.ErrorMarshal, err)), Place: place}
	}
	defer CacheMetrics(place, start)
	err = redisrepo.Client.RedisClient.Set(ctx, user.Id.String(), jsondata, 1*time.Hour).Err()
	if err != nil {
		metrics.UserCacheErrorsTotal.WithLabelValues("SET").Inc()
		return &RepositoryResponse{Success: false, Errors: erro.ServerError(fmt.Sprintf(erro.ErrorSetProfiles, err)), Place: place}
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
	result, err := redisrepo.Client.RedisClient.Get(ctx, id).Result()
	if err != nil {
		if err == redis.Nil {
			return &RepositoryResponse{Success: false, SuccessMessage: "Profile was not found in the cache", Place: place}
		}
		metrics.UserCacheErrorsTotal.WithLabelValues("GET").Inc()
		return &RepositoryResponse{Success: false, Errors: erro.ServerError(fmt.Sprintf(erro.ErrorGetProfiles, err)), Place: place}
	}
	var user model.User
	err = json.Unmarshal([]byte(result), &user)
	if err != nil {
		return &RepositoryResponse{Success: false, Errors: erro.ServerError(fmt.Sprintf(erro.ErrorUnmarshal, err)), Place: place}
	}
	return &RepositoryResponse{Success: true, Data: map[string]any{KeyUser: &user}, SuccessMessage: "Successful get profile from cache", Place: place}
}
