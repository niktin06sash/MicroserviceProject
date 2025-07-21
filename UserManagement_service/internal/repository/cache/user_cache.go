package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/erro"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/metrics"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/model"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/repository"
	"github.com/redis/go-redis/v9"
)

type UserCache struct {
	cacheclient *CacheObject
}

const GetProfileCache = "Repository-GetProfileCache"
const DeleteProfileCache = "Repository-DeleteProfileCache"
const AddProfileCache = "Repository-AddProfileCache"

func NewUserCache(red *CacheObject) *UserCache {
	return &UserCache{cacheclient: red}
}
func CacheMetrics(place string, start time.Time) {
	metrics.UserCacheQueriesTotal.WithLabelValues(place).Inc()
	duration := time.Since(start).Seconds()
	metrics.UserCacheQueryDuration.WithLabelValues(place).Observe(duration)
}
func (redisrepo *UserCache) AddProfileCache(ctx context.Context, user *model.User) *repository.RepositoryResponse {
	const place = AddProfileCache
	start := time.Now()
	jsondata, err := json.Marshal(user)
	if err != nil {
		return repository.BadResponse(erro.ServerError(fmt.Sprintf(erro.ErrorMarshal, err)), place)
	}
	defer CacheMetrics(place, start)
	err = redisrepo.cacheclient.connect.Set(ctx, user.Id.String(), jsondata, 1*time.Hour).Err()
	if err != nil {
		metrics.UserCacheErrorsTotal.WithLabelValues("SET").Inc()
		return repository.BadResponse(erro.ServerError(fmt.Sprintf(erro.ErrorSetProfiles, err)), place)
	}
	return repository.SuccessResponse(nil, place, "Successful add profile in cache")
}
func (redisrepo *UserCache) DeleteProfileCache(ctx context.Context, id string) *repository.RepositoryResponse {
	const place = DeleteProfileCache
	start := time.Now()
	defer CacheMetrics(place, start)
	num, err := redisrepo.cacheclient.connect.Del(ctx, id).Result()
	if err != nil {
		metrics.UserCacheErrorsTotal.WithLabelValues("DEL").Inc()
		return repository.BadResponse(erro.ServerError(fmt.Sprintf(erro.ErrorDelProfiles, err)), place)
	}
	if num == 0 {
		return repository.SuccessResponse(nil, place, "Profile was not found in the cache")
	}
	return repository.SuccessResponse(nil, place, "Successful delete profile from cache")
}
func (redisrepo *UserCache) GetProfileCache(ctx context.Context, id string) *repository.RepositoryResponse {
	const place = GetProfileCache
	start := time.Now()
	defer CacheMetrics(place, start)
	result, err := redisrepo.cacheclient.connect.Get(ctx, id).Result()
	if err != nil {
		if err == redis.Nil {
			return &repository.RepositoryResponse{Success: false, SuccessMessage: "Profile was not found in the cache", Place: place}
		}
		metrics.UserCacheErrorsTotal.WithLabelValues("GET").Inc()
		return repository.BadResponse(erro.ServerError(fmt.Sprintf(erro.ErrorGetProfiles, err)), place)
	}
	var user model.User
	err = json.Unmarshal([]byte(result), &user)
	if err != nil {
		return repository.BadResponse(erro.ServerError(fmt.Sprintf(erro.ErrorUnmarshal, err)), place)
	}
	return repository.SuccessResponse(map[string]any{repository.KeyUser: &user}, place, "Successful add profile in cache")
}
