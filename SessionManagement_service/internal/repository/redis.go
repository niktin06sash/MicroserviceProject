package repository

import (
	"context"
	"fmt"

	"github.com/niktin06sash/MicroserviceProject/SessionManagement_service/internal/configs"
	"github.com/niktin06sash/MicroserviceProject/SessionManagement_service/internal/logger"

	"github.com/redis/go-redis/v9"
)

type RedisInterface interface {
	Open(host string, port int, password string, db int) error
	Ping() error
	Close()
}

type RedisObject struct {
	RedisClient *redis.Client
	Logger      *logger.SessionLogger
}

func NewRedisConnection(cfg configs.RedisConfig, logger *logger.SessionLogger) (*RedisObject, error) {
	redisobject := &RedisObject{Logger: logger}
	redisobject.Open(cfg.Host, cfg.Port, cfg.Password, cfg.DB)
	err := redisobject.Ping()
	if err != nil {
		redisobject.Close()
		return nil, fmt.Errorf("failed to establish database connection: %w", err)
	}
	redisobject.Logger.Info("SessionManagement: Successful connect to Redis-Client!")
	return redisobject, nil
}
func (r *RedisObject) Open(host string, port int, password string, db int) {
	r.RedisClient = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", host, port),
		Password: password,
		DB:       db,
	})
}

func (r *RedisObject) Ping() error {
	_, err := r.RedisClient.Ping(context.Background()).Result()
	return err
}

func (r *RedisObject) Close() {
	r.RedisClient.Close()
	r.Logger.Info("SessionManagement: Successful close Redis-Client!")
}
