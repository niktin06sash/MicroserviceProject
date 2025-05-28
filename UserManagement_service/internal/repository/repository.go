package repository

import (
	"context"
	"database/sql"

	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/erro"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/kafka"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/model"

	"github.com/google/uuid"
)

//go:generate mockgen -source=repository.go -destination=mocks/mock.go
type DBUserRepos interface {
	CreateUser(ctx context.Context, tx *sql.Tx, user *model.User) *RepositoryResponse
	AuthenticateUser(ctx context.Context, useremail, password string) *RepositoryResponse
	DeleteUser(ctx context.Context, tx *sql.Tx, userId uuid.UUID, password string) *RepositoryResponse
	UpdateUserName(ctx context.Context, userId uuid.UUID, name string) *RepositoryResponse
	UpdateUserEmail(ctx context.Context, userId uuid.UUID, email string, password string) *RepositoryResponse
	UpdateUserPassword(ctx context.Context, userId uuid.UUID, lastpassword string, newpassword string) *RepositoryResponse
	GetMyProfile(ctx context.Context, userid uuid.UUID) *RepositoryResponse
	GetProfileById(ctx context.Context, getid uuid.UUID) *RepositoryResponse
}
type DBTransactionManager interface {
	BeginTx(ctx context.Context) (*sql.Tx, error)
	RollbackTx(tx *sql.Tx) error
	CommitTx(tx *sql.Tx) error
}
type RedisUserRepos interface {
	AddProfile(ctx context.Context, id uuid.UUID, data map[string]any) *RepositoryResponse
	DeleteProfile(ctx context.Context, id uuid.UUID) *RepositoryResponse
	GetProfile(ctx context.Context, id uuid.UUID) *RepositoryResponse
}

const CreateUser = "Repository-CreateUser"
const AuthenticateUser = "Repository-AuthenticateUser"
const DeleteUser = "Repository-DeleteUser"
const UpdateName = "Repository-UpdateName"
const UpdatePassword = "Repository-UpdatePassword"
const UpdateEmail = "Repository-UpdateEmail"
const GetMyProfile = "Repository-GetMyProfile"
const GetProfileById = "Repository-GetProfileById"

type Repository struct {
	DBUserRepos
	DBTransactionManager
	RedisUserRepos
}
type RepositoryResponse struct {
	Success bool
	Data    map[string]any
	Errors  *erro.ErrorResponse
}

func NewRepository(db *DBObject, redis *RedisObject, kafka kafka.KafkaProducerService) *Repository {
	return &Repository{
		DBUserRepos:          NewAuthPostgresRepo(db, kafka),
		DBTransactionManager: NewTxManagerRepo(db),
		RedisUserRepos:       NewAuthRedisRepo(redis, kafka),
	}
}
