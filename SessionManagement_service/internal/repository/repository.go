package repository

import (
	"context"

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
	SuccessMessage string
	Data           map[string]any
	Errors         error
	Place          string
}

const KeySessionId = "sessionID"
const KeyUserId = "userID"
const ExpiryTime = "expirytime"
const SetSession = "Repository-SetSession"
const GetSession = "Repository-GetSession"
const DeleteSession = "Repository-DeleteSession"

func NewRepository(client *RedisObject) *Repository {
	return &Repository{
		RedisSessionRepos: NewSessionRedis(client),
	}
}
