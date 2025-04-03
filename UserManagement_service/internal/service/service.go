package service

import (
	"UserManagement_service/internal/kafka"
	"UserManagement_service/internal/model"
	"UserManagement_service/internal/repository"
	"context"
	"time"

	"github.com/google/uuid"
)

//go:generate mockgen -source=service.go -destination=mocks/mock.go
type UserAuthentication interface {
	RegistrateAndLogin(ctx context.Context, user *model.Person) *ServiceResponse
	AuthenticateAndLogin(ctx context.Context, user *model.Person) *ServiceResponse
	//Logout(ctx context.Context, sessionID string, userId uuid.UUID) *ServiceResponse
	//DeleteAccount(ctx context.Context, sessionID string, userid uuid.UUID, password string) *ServiceResponse
}
type Service struct {
	UserAuthentication
}
type ServiceResponse struct {
	Success       bool
	UserId        uuid.UUID
	SessionId     string
	ExpireSession time.Time
	Errors        map[string]error
}

func NewService(repos *repository.Repository, kafkaProd kafka.KafkaProducer) *Service {

	return &Service{

		UserAuthentication: NewAuthService(repos.DBAuthenticateRepos, kafkaProd),
	}
}
