package service

import (
	"context"

	"github.com/niktin06sash/MicroserviceProject/SessionManagement_service/internal/model"
	"github.com/niktin06sash/MicroserviceProject/SessionManagement_service/internal/repository"
)

type SessionRepos interface {
	SetSession(ctx context.Context, session model.Session) *repository.RepositoryResponse
	GetSession(ctx context.Context, sessionID string) *repository.RepositoryResponse
	DeleteSession(ctx context.Context, sessionID string) *repository.RepositoryResponse
}
type LogProducer interface {
	NewSessionLog(level, place, traceid, msg string)
}

const UseCase_CreateSession = "UseCase-CreateSession"
const UseCase_ValidateSession = "UseCase-ValidateSession"
const UseCase_DeleteSession = "UseCase-DeleteSession"

type ServiceResponse struct {
	Success bool
	Data    Data
	Errors  map[string]string
}
type Data struct {
	SessionID      string
	UserID         string
	ExpirationTime int64
}
