package service

import (
	"context"
	"time"

	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/client"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/kafka"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/model"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/repository"

	"github.com/google/uuid"
)

//go:generate mockgen -source=service.go -destination=mocks/mock.go
type UserAuthentication interface {
	RegistrateAndLogin(ctx context.Context, user *model.Person) *ServiceResponse
	AuthenticateAndLogin(ctx context.Context, user *model.Person) *ServiceResponse
	DeleteAccount(ctx context.Context, sessionID string, userid uuid.UUID, password string) *ServiceResponse
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

func NewService(repos *repository.Repository, kafkaProd kafka.KafkaProducer, clientgrpc *client.GrpcClient) *Service {

	return &Service{

		UserAuthentication: NewAuthService(repos.DBAuthenticateRepos, kafkaProd, clientgrpc),
	}
}
