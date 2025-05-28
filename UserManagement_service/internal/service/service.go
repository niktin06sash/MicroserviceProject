package service

import (
	"context"

	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/client"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/kafka"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/model"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/repository"
)

//go:generate mockgen -source=service.go -destination=mocks/mock.go
type UserService interface {
	RegistrateAndLogin(ctx context.Context, req *model.RegistrationRequest) *ServiceResponse
	AuthenticateAndLogin(ctx context.Context, req *model.AuthenticationRequest) *ServiceResponse
	DeleteAccount(ctx context.Context, req *model.DeletionRequest, sessionID string, useridstr string) *ServiceResponse
	Logout(ctx context.Context, sessionID string) *ServiceResponse
	UpdateAccount(ctx context.Context, req *model.UpdateRequest, useridstr string, updateType string) *ServiceResponse
	GetMyProfile(ctx context.Context, useridstr string) *ServiceResponse
	GetProfileById(ctx context.Context, useridstr string, findidstr string) *ServiceResponse
}

const RegistrateAndLogin = "UseCase-RegistrateAndLogin"
const AuthenticateAndLogin = "UseCase-AuthenticateAndLogin"
const DeleteAccount = "UseCase-DeleteAccount"
const Logout = "UseCase-Logout"
const UpdateAccount = "UseCase-UpdateAccount"
const GetMyProfile = "UseCase-GetMyProfile"
const GetProfileById = "UseCase-GetProfileById"

type Service struct {
	UserService
}
type ServiceResponse struct {
	Success   bool
	Data      map[string]any
	Errors    map[string]string
	ErrorType string
}

func NewService(repos *repository.Repository, kafkaProd kafka.KafkaProducerService, clientgrpc *client.GrpcClient) *Service {

	return &Service{

		UserService: NewAuthService(repos.DBUserRepos, repos.DBTransactionManager, repos.RedisUserRepos, kafkaProd, clientgrpc),
	}
}
