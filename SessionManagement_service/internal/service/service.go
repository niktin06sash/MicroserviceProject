package service

import (
	"context"

	"github.com/niktin06sash/MicroserviceProject/SessionManagement_service/internal/erro"
	"github.com/niktin06sash/MicroserviceProject/SessionManagement_service/internal/model"
	"github.com/niktin06sash/MicroserviceProject/SessionManagement_service/internal/repository"
)

//go:generate mockgen -source=service.go -destination=mocks/mock.go
type SessionRepos interface {
	SetSession(ctx context.Context, session *model.Session) *repository.RepositoryResponse
	GetSession(ctx context.Context, sessionID string, flag string) *repository.RepositoryResponse
	DeleteSession(ctx context.Context, sessionID string) *repository.RepositoryResponse
}
type LogProducer interface {
	NewSessionLog(level, place, traceid, msg string)
}

const CreateSession = "UseCase-CreateSession"
const ValidateSession = "UseCase-ValidateSession"
const DeleteSession = "UseCase-DeleteSession"

type ServiceResponse struct {
	Success bool
	Data    Data
	Errors  *erro.CustomError
}
type Data struct {
	SessionID      string
	UserID         string
	ExpirationTime int64
}
