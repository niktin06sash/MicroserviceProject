package cache

import (
	"context"
	"fmt"
	"log"

	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/configs"
	"github.com/redis/go-redis/v9"
)

type CacheObject struct {
	connect *redis.Client
}

func NewRedisConnection(cfg configs.RedisConfig) (*CacheObject, error) {
	redisobject := &CacheObject{}
	redisobject.Open(cfg.Host, cfg.Port, cfg.Password, cfg.DB)
	err := redisobject.Ping()
	if err != nil {
		redisobject.Close()
		log.Printf("[DEBUG] [User-Service] Failed to establish Redis-Client connection: %v", err)
		return nil, err
	}
	log.Println("[DEBUG] [User-Service] Successful connect to Redis-Client")
	return redisobject, nil
}
func (r *CacheObject) Open(host string, port int, password string, db int) {
	r.connect = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", host, port),
		Password: password,
		DB:       db,
	})
}

func (r *CacheObject) Ping() error {
	_, err := r.connect.Ping(context.Background()).Result()
	return err
}

func (r *CacheObject) Close() {
	r.connect.Close()
	log.Println("[DEBUG] [User-Service] Successful close Redis-Client")
}
