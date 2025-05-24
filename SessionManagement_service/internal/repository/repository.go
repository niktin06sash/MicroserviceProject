package repository

import (
	"context"
	"time"

	"github.com/niktin06sash/MicroserviceProject/SessionManagement_service/internal/kafka"
	"github.com/niktin06sash/MicroserviceProject/SessionManagement_service/internal/model"
)

//go:generate mockgen -source=repository.go -destination=mocks/mock.go
type RedisSessionRepos interface {
	SetSession(ctx context.Context, session model.Session) *RepositoryResponse
	GetSession(ctx context.Context, sessionID string) *RepositoryResponse
	DeleteSession(ctx context.Context, sessionID string) *RepositoryResponse
}
type Repository struct {
	RedisSessionRepos
}
type RepositoryResponse struct {
	Success        bool
	SessionId      string
	ExpirationTime time.Time
	UserID         string
	Errors         error
}

const SetSession = "Repository-SetSession"
const GetSession = "Repository-GetSession"
const DeleteSession = "Repository-DeleteSession"

func NewRepository(client *RedisObject, kafkaprod kafka.KafkaProducerService) *Repository {
	return &Repository{
		RedisSessionRepos: NewAuthRedis(client, kafkaprod),
	}
}
