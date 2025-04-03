package repository

import (
	"UserManagement_service/internal/model"
	"context"
	"database/sql"

	"github.com/google/uuid"
)

//go:generate mockgen -source=repository.go -destination=mocks/mock.go
type DBAuthenticateRepos interface {
	CreateUser(ctx context.Context, user *model.Person) *DBRepositoryResponse
	GetUser(ctx context.Context, useremail, password string) *DBRepositoryResponse
	DeleteUser(ctx context.Context, userId uuid.UUID, password string) *DBRepositoryResponse
	BeginTx(ctx context.Context) (*sql.Tx, error)
	RollbackTx(ctx context.Context, tx *sql.Tx) error
	CommitTx(ctx context.Context, tx *sql.Tx) error
}
type Repository struct {
	DBAuthenticateRepos
}
type DBRepositoryResponse struct {
	Success bool
	UserId  uuid.UUID
	Errors  error
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		DBAuthenticateRepos: NewAuthPostgres(db),
	}
}
