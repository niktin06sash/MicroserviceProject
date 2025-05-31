package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/erro"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/metrics"
)

type FriendshipRedisRepo struct {
	Client *RedisObject
}

func NewFriendshipRedisRepo(red *RedisObject) *FriendshipRedisRepo {
	return &FriendshipRedisRepo{Client: red}
}
func (redisrepo *FriendshipRedisRepo) AddFriendsCache(ctx context.Context, id string, data map[string]any) *RepositoryResponse {
	const place = AddFriendsCache
	start := time.Now()
	defer CacheMetrics(place, start)
	key := fmt.Sprintf("user:%s:friends", id)
	traceID := ctx.Value("traceID").(string)
	friends := data[KeyFriends].([]string)
	jsonfriends, err := json.Marshal(friends)
	if err != nil {
		metrics.UserDBErrorsTotal.WithLabelValues("MARSHAL").Inc()
		return &RepositoryResponse{Success: false, Errors: &erro.ErrorResponse{Message: erro.UserServiceUnavalaible, Type: erro.ServerErrorType}}
	}
	err = redisrepo.Client.RedisClient.HSet(ctx, key, map[string]interface{}{
		"Friends": string(jsonfriends),
	}).Err()
	if err != nil {
		metrics.UserDBErrorsTotal.WithLabelValues("HSET").Inc()
		return &RepositoryResponse{Success: false, Errors: &erro.ErrorResponse{Message: erro.UserServiceUnavalaible, Type: erro.ServerErrorType}}
	}
	err = redisrepo.Client.RedisClient.Expire(ctx, key, time.Until(time.Now().Add(1*time.Hour))).Err()
	if err != nil {
		metrics.UserDBErrorsTotal.WithLabelValues("EXPIRE").Inc()
		return &RepositoryResponse{Success: false, Errors: &erro.ErrorResponse{Message: erro.UserServiceUnavalaible, Type: erro.ServerErrorType}}
	}
	return &RepositoryResponse{Success: true}
}
func (redisrepo *FriendshipRedisRepo) DeleteFriendsCache(ctx context.Context, id string) *RepositoryResponse {
	const place = DeleteFriendsCache
	start := time.Now()
	defer CacheMetrics(place, start)
	key := fmt.Sprintf("user:%s:friends", id)
	traceID := ctx.Value("traceID").(string)
	num, err := redisrepo.Client.RedisClient.Del(ctx, key).Result()
	if err != nil {
		metrics.UserDBErrorsTotal.WithLabelValues("DEL").Inc()
		return &RepositoryResponse{Success: false, Errors: &erro.ErrorResponse{Message: erro.UserServiceUnavalaible, Type: erro.ServerErrorType}}
	}
	if num == 0 {
		return &RepositoryResponse{
			Success: false,
		}
	}
	return &RepositoryResponse{Success: true}
}
func (redisrepo *FriendshipRedisRepo) GetFriendsCache(ctx context.Context, id string) *RepositoryResponse {
	const place = GetFriendsCache
	start := time.Now()
	defer CacheMetrics(place, start)
	key := fmt.Sprintf("user:%s:friends", id)
	traceID := ctx.Value("traceID").(string)
	result, err := redisrepo.Client.RedisClient.HGetAll(ctx, key).Result()
	if err != nil {
		metrics.UserDBErrorsTotal.WithLabelValues("HGETALL").Inc()
		return &RepositoryResponse{Success: false, Errors: &erro.ErrorResponse{Message: erro.UserServiceUnavalaible, Type: erro.ServerErrorType}}
	}
	if len(result) == 0 {
		return &RepositoryResponse{Success: false}
	}
	var friends []string
	err = json.Unmarshal([]byte(result["Friends"]), &friends)
	if err != nil {
		metrics.UserDBErrorsTotal.WithLabelValues("UNMARSAL").Inc()
		return &RepositoryResponse{Success: false, Errors: &erro.ErrorResponse{Message: erro.UserServiceUnavalaible, Type: erro.ServerErrorType}}
	}
	return &RepositoryResponse{Success: true, Data: map[string]any{KeyFriends: friends}}
}
