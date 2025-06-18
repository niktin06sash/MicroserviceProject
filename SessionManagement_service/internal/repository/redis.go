package repository

import (
	"context"
	"fmt"
	"log"

	"github.com/niktin06sash/MicroserviceProject/SessionManagement_service/internal/configs"

	"github.com/redis/go-redis/v9"
)

type RedisObject struct {
	RedisClient *redis.Client
}

func NewRedisConnection(cfg configs.RedisConfig) (*RedisObject, error) {
	redisobject := &RedisObject{}
	redisobject.Open(cfg.Host, cfg.Port, cfg.Password, cfg.DB)
	err := redisobject.Ping()
	if err != nil {
		redisobject.Close()
		log.Printf("[DEBUG] [Session-Service] Failed to establish Redis-Client connection: %v", err)
		return nil, err
	}
	log.Println("[DEBUG] [Session-Service] Successful connect to Redis-Client")
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
	log.Println("[DEBUG] [Session-Service] Successful close Redis-Client")
}
