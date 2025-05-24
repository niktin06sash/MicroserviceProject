package service

import (
	"context"
	"time"

	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/client"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/erro"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/kafka"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/model"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/repository"

	"github.com/google/uuid"
)

//go:generate mockgen -source=service.go -destination=mocks/mock.go
type UserAuthentication interface {
	RegistrateAndLogin(ctx context.Context, user *model.Person) *ServiceResponse
	AuthenticateAndLogin(ctx context.Context, user *model.Person) *ServiceResponse
	DeleteAccount(ctx context.Context, sessionID string, userid string, password string) *ServiceResponse
	Logout(ctx context.Context, sessionID string) *ServiceResponse
	UpdateAccount(ctx context.Context, data map[string]string, updateType string) *ServiceResponse
}

const RegistrateAndLogin = "UseCase-RegistrateAndLogin"
const AuthenticateAndLogin = "UseCase-AuthenticateAndLogin"
const DeleteAccount = "UseCase-DeleteAccount"
const Logout = "UseCase-Logout"
const UpdateAccount = "UseCase-UpdateAccount"

type Service struct {
	UserAuthentication
}
type ServiceResponse struct {
	Success       bool
	UserId        uuid.UUID
	SessionId     string
	ExpireSession time.Time
	Errors        map[string]error
	Type          erro.ErrorType
}

func NewService(repos *repository.Repository, kafkaProd kafka.KafkaProducerService, clientgrpc *client.GrpcClient) *Service {

	return &Service{

		UserAuthentication: NewAuthService(repos.DBAuthenticateRepos, repos.DBTransactionManager, kafkaProd, clientgrpc),
	}
}
