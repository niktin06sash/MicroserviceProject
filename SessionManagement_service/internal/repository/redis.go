package repository

import (
	"context"
	"fmt"
	"log"

	"github.com/niktin06sash/MicroserviceProject/SessionManagement_service/internal/configs"

	"github.com/redis/go-redis/v9"
)

type RedisInterface interface {
	Open(host string, port int, password string, db int) (*redis.Client, error)
	Ping(client *redis.Client) error
	Close(client *redis.Client) error
}

type RedisObject struct{}

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

func ConnectToRedis(cfg configs.Config) (*redis.Client, RedisInterface, error) {
	redisInterface := &RedisObject{}

	client, err := redisInterface.Open(cfg.Redis.Host, cfg.Redis.Port, cfg.Redis.Password, cfg.Redis.DB)
	if err != nil {
		log.Printf("Redis-Open error %v", err)
		return nil, nil, err
	}
	err = redisInterface.Ping(client)
	if err != nil {
		redisInterface.Close(client)
		log.Printf("Redis-Ping error %v", err)
		return nil, nil, err
	}
	log.Println("RedisManagement: Successful connect to Redis-Client!")
	return client, redisInterface, nil
}
