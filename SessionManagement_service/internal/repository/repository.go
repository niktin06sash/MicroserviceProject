package repository

import (
	"context"
	"time"

	"github.com/niktin06sash/MicroserviceProject/SessionManagement_service/internal/model"

	"github.com/redis/go-redis/v9"
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

func NewRepository(client *redis.Client) *Repository {
	return &Repository{
		RedisSessionRepos: NewAuthRedis(client),
	}
}
