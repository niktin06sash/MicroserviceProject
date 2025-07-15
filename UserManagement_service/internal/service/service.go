package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	pb "github.com/niktin06sash/MicroserviceProject/SessionManagement_service/proto"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/model"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/repository"
)

const RegistrateAndLogin = "UseCase-RegistrateAndLogin"
const AuthenticateAndLogin = "UseCase-AuthenticateAndLogin"
const DeleteAccount = "UseCase-DeleteAccount"
const Logout = "UseCase-Logout"
const UpdateAccount = "UseCase-UpdateAccount"
const GetMyProfile = "UseCase-GetMyProfile"
const GetProfileById = "UseCase-GetProfileById"
const GetMyFriends = "UseCase-GetMyFriends"
const KeyExpirySession = "expirysession"
const KeySessionID = "sessionid"

type ServiceResponse struct {
	Success bool
	Data    map[string]any
	Errors  error
}

//go:generate mockgen -source=service.go -destination=mocks/mock.go
type DBUserRepos interface {
	CreateUser(ctx context.Context, tx pgx.Tx, user *model.User) *repository.RepositoryResponse
	GetUser(ctx context.Context, useremail, password string) *repository.RepositoryResponse
	DeleteUser(ctx context.Context, tx pgx.Tx, userId uuid.UUID, password string) *repository.RepositoryResponse
	UpdateUserData(ctx context.Context, tx pgx.Tx, userId uuid.UUID, updateType string, args ...interface{}) *repository.RepositoryResponse
	GetProfileById(ctx context.Context, userid uuid.UUID) *repository.RepositoryResponse
}
type DBTransactionManager interface {
	BeginTx(ctx context.Context) (pgx.Tx, error)
	RollbackTx(ctx context.Context, tx pgx.Tx) error
	CommitTx(ctx context.Context, tx pgx.Tx) error
}
type CacheUserRepos interface {
	AddProfileCache(ctx context.Context, user *model.User) *repository.RepositoryResponse
	DeleteProfileCache(ctx context.Context, id string) *repository.RepositoryResponse
	GetProfileCache(ctx context.Context, id string) *repository.RepositoryResponse
}
type EventProducer interface {
	NewUserEvent(ctx context.Context, routingKey string, userid string, place string, traceid string) error
}
type LogProducer interface {
	NewUserLog(level, place, traceid, msg string)
}
type SessionClient interface {
	CreateSession(ctx context.Context, userID string) (*pb.CreateSessionResponse, error)
	DeleteSession(ctx context.Context, sessionID string) (*pb.DeleteSessionResponse, error)
}
