package repository

import (
	"context"
	"database/sql"

	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/model"

	"github.com/google/uuid"
)

//go:generate mockgen -source=repository.go -destination=mocks/mock.go
type DBAuthenticateRepos interface {
	CreateUser(ctx context.Context, user *model.Person) *DBRepositoryResponse
	GetUser(ctx context.Context, useremail, password string) *DBRepositoryResponse
	DeleteUser(ctx context.Context, userId uuid.UUID, password string) *DBRepositoryResponse
}
type DBTransactionManager interface {
	BeginTx(ctx context.Context) (*sql.Tx, error)
	RollbackTx(tx *sql.Tx) error
	CommitTx(tx *sql.Tx) error
}
type Repository struct {
	DBAuthenticateRepos
	DBTransactionManager
}
type DBRepositoryResponse struct {
	Success bool
	UserId  uuid.UUID
	Errors  error
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		DBAuthenticateRepos:  NewAuthPostgresRepo(db),
		DBTransactionManager: NewTxManagerRepo(db),
	}
}
