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
	CreateUser(ctx context.Context, tx *sql.Tx, user *model.User) *DBRepositoryResponse
	AuthenticateUser(ctx context.Context, useremail, password string) *DBRepositoryResponse
	DeleteUser(ctx context.Context, tx *sql.Tx, userId uuid.UUID, password string) *DBRepositoryResponse
	UpdateUserName(ctx context.Context, userId uuid.UUID, name string) *DBRepositoryResponse
	UpdateUserEmail(ctx context.Context, userId uuid.UUID, email string, password string) *DBRepositoryResponse
	UpdateUserPassword(ctx context.Context, userId uuid.UUID, lastpassword string, newpassword string) *DBRepositoryResponse
	GetMyProfile(ctx context.Context, userid uuid.UUID) *DBRepositoryResponse
	GetProfileById(ctx context.Context, getid uuid.UUID) *DBRepositoryResponse
}
type DBTransactionManager interface {
	BeginTx(ctx context.Context) (*sql.Tx, error)
	RollbackTx(tx *sql.Tx) error
	CommitTx(tx *sql.Tx) error
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
}
type DBRepositoryResponse struct {
	Success bool
	Data    map[string]any
	Errors  *erro.ErrorResponse
}
type ErrorResponse struct {
	Message string
	Type    string
}

func NewRepository(db *DBObject, kafka kafka.KafkaProducerService) *Repository {
	return &Repository{
		DBUserRepos:          NewAuthPostgresRepo(db, kafka),
		DBTransactionManager: NewTxManagerRepo(db),
	}
}
