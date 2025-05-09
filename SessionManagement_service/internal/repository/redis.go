package repository

import (
	"context"
	"fmt"
	"log"

	"github.com/niktin06sash/MicroserviceProject/SessionManagement_service/internal/configs"

	"github.com/redis/go-redis/v9"
)

type RedisInterface interface {
	Open(host string, port int, password string, db int) error
	Ping() error
	Close()
}

type RedisObject struct {
	RedisClient *redis.Client
}

func NewRedisConnection(cfg configs.RedisConfig) *RedisObject {
	redisobject := &RedisObject{}
	redisobject.Open(cfg.Host, cfg.Port, cfg.Password, cfg.DB)
	err := redisobject.Ping()
	if err != nil {
		redisobject.Close()
		log.Fatalf("[INFO] [Session-Service] Failed to establish Redis-Client connection: %v", err)
	}
	log.Println("[INFO] [Session-Service] Successful connect to Redis-Client")
	return redisobject
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
	log.Println("[INFO] [Session-Service] Successful close Redis-Client")
}
