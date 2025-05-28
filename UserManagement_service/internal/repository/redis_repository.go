package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/kafka"
)

type AuthRedisRepo struct {
	Db            *RedisObject
	KafkaProducer kafka.KafkaProducerService
}

func NewAuthRedisRepo(db *RedisObject, kafkaprod kafka.KafkaProducerService) *AuthRedisRepo {
	return &AuthRedisRepo{Db: db, KafkaProducer: kafkaprod}
}

func (redis *AuthRedisRepo) AddProfile(ctx context.Context, id uuid.UUID, data map[string]any) *RepositoryResponse {

}
func (redis *AuthRedisRepo) DeleteProfile(ctx context.Context, id uuid.UUID) *RepositoryResponse {

}
func (redis *AuthRedisRepo) GetProfile(ctx context.Context, id uuid.UUID) *RepositoryResponse {

}
