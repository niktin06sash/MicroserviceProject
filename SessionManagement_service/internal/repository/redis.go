package repository

import (
	"context"
	"fmt"
	"log"

	"github.com/niktin06sash/MicroserviceProject/SessionManagement_service/internal/configs"

	"github.com/redis/go-redis/v9"
)

type DatabaseObject struct {
	connect *redis.Client
}

func NewRedisConnection(cfg configs.RedisConfig) (*DatabaseObject, error) {
	redisobject := &DatabaseObject{}
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
func (r *DatabaseObject) Open(host string, port int, password string, db int) {
	r.connect = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", host, port),
		Password: password,
		DB:       db,
	})
}

func (r *DatabaseObject) Ping() error {
	_, err := r.connect.Ping(context.Background()).Result()
	return err
}

func (r *DatabaseObject) Close() {
	r.connect.Close()
	log.Println("[DEBUG] [Session-Service] Successful close Redis-Client")
}
