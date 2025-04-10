package repository

import (
	"context"
	"database/sql"
	"errors"
	"log"

	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/erro"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/model"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthPostgres struct {
	Db *sql.DB
}

func (repoap *AuthPostgres) CreateUser(ctx context.Context, user *model.Person) *DBRepositoryResponse {
	var createdUserID uuid.UUID

	err := repoap.Db.QueryRowContext(ctx,
		"INSERT INTO UserZ (userid, username, useremail, userpassword) values ($1, $2, $3, $4) ON CONFLICT (useremail) DO NOTHING RETURNING userid;",
		user.Id, user.Name, user.Email, user.Password).Scan(&createdUserID)

	if err != nil {
		log.Printf("CreateUser Error: %v", err)
		if errors.Is(err, sql.ErrNoRows) {
			return &DBRepositoryResponse{Success: false, Errors: erro.ErrorUniqueEmail}
		}
		return &DBRepositoryResponse{Success: false, Errors: erro.ErrorDbRepositoryError}
	}

	log.Println("Successful create person!")
	return &DBRepositoryResponse{Success: true, UserId: createdUserID, Errors: nil}
}

func (repoap *AuthPostgres) GetUser(ctx context.Context, useremail, userpassword string) *DBRepositoryResponse {
	var hashpass string
	var userId uuid.UUID
	err := repoap.Db.QueryRowContext(ctx, "SELECT userid, userpassword FROM userZ WHERE useremail = $1", useremail).Scan(&userId, &hashpass)

	if err != nil {
		log.Printf("GetUser Error: %v", err)
		if errors.Is(err, sql.ErrNoRows) {
			return &DBRepositoryResponse{Success: false, Errors: erro.ErrorEmailNotRegister}
		}
		return &DBRepositoryResponse{Success: false, Errors: erro.ErrorDbRepositoryError}
	}

	err = bcrypt.CompareHashAndPassword([]byte(hashpass), []byte(userpassword))
	if err != nil {
		log.Printf("CompareHashAndPassword Error: %v", err)
		return &DBRepositoryResponse{Success: false, Errors: erro.ErrorInvalidPassword}
	}

	log.Println("Successful get person!")
	return &DBRepositoryResponse{Success: true, UserId: userId, Errors: nil}
}
func (repoap *AuthPostgres) DeleteUser(ctx context.Context, userId uuid.UUID, password string) *DBRepositoryResponse {
	var hashpass string
	err := repoap.Db.QueryRowContext(ctx, "SELECT userpassword FROM userZ WHERE userid = $1", userId).Scan(&hashpass)

	if err != nil {
		log.Printf("DeleteUser Error: %v", err)
		if errors.Is(err, sql.ErrNoRows) {
			return &DBRepositoryResponse{Success: false, Errors: erro.ErrorFoundUser}
		}
		return &DBRepositoryResponse{Success: false, Errors: erro.ErrorDbRepositoryError}
	}
	err = bcrypt.CompareHashAndPassword([]byte(hashpass), []byte(password))
	if err != nil {
		log.Printf("CompareHashAndPassword Error: %v", err)
		return &DBRepositoryResponse{Success: false, Errors: erro.ErrorInvalidPassword}
	}
	_, err = repoap.Db.ExecContext(ctx, "DELETE FROM userZ where userId = $1", userId)
	if err != nil {
		log.Printf("Delete Error: %v", err)
		return &DBRepositoryResponse{Success: false, Errors: erro.ErrorDbRepositoryError}
	}
	return &DBRepositoryResponse{Success: true}
}
func (r *AuthPostgres) BeginTx(ctx context.Context) (*sql.Tx, error) {
	return r.Db.BeginTx(ctx, nil)
}

func (r *AuthPostgres) RollbackTx(ctx context.Context, tx *sql.Tx) error {
	return tx.Rollback()
}

func (r *AuthPostgres) CommitTx(ctx context.Context, tx *sql.Tx) error {
	return tx.Commit()
}
func NewAuthPostgres(db *sql.DB) *AuthPostgres {
	return &AuthPostgres{Db: db}
}
