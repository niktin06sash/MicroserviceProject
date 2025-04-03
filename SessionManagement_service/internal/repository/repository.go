package repository

import (
	"SessionManagement_service/internal/model"
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

//go:generate mockgen -source=repository.go -destination=mocks/mock.go
type RedisSessionRepos interface {
	SetSession(ctx context.Context, session model.Session, expiration time.Duration) *RepositoryResponse
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
	UserID         uuid.UUID
	Errors         error
}

func NewRepository(db *sql.DB, client *redis.Client) *Repository {
	return &Repository{
		RedisSessionRepos: NewAuthRedis(client),
	}
}
