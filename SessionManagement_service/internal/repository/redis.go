package repository

import (
	"context"
	"fmt"

	"github.com/niktin06sash/MicroserviceProject/SessionManagement_service/internal/configs"
	"github.com/niktin06sash/MicroserviceProject/SessionManagement_service/internal/logger"

	"github.com/redis/go-redis/v9"
)

type RedisInterface interface {
	Open(host string, port int, password string, db int) (*redis.Client, error)
	Ping(client *redis.Client) error
	Close(client *redis.Client) error
	GetLogger() *logger.SessionLogger
}

type RedisObject struct {
	Logger *logger.SessionLogger
}

func (r *RedisObject) Open(host string, port int, password string, db int) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", host, port),
		Password: password,
		DB:       db,
	})
	return client, nil
}

func (r *RedisObject) Ping(client *redis.Client) error {
	_, err := client.Ping(context.Background()).Result()
	return err
}

func (r *RedisObject) Close(client *redis.Client) error {
	return client.Close()
}
func (r *RedisObject) GetLogger() *logger.SessionLogger {
	return r.Logger
}
func ConnectToRedis(cfg configs.Config, redisInterface RedisInterface) (*redis.Client, error) {
	logger := redisInterface.GetLogger()
	client, err := redisInterface.Open(cfg.Redis.Host, cfg.Redis.Port, cfg.Redis.Password, cfg.Redis.DB)
	if err != nil {
		return nil, err
	}
	err = redisInterface.Ping(client)
	if err != nil {
		redisInterface.Close(client)
		return nil, err
	}
	logger.Info("SessionManagement: Successful connect to Redis-Client!")
	return client, nil
}
