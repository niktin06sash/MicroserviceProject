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

type AuthPostgresRepo struct {
	Db *sql.DB
}

func NewAuthPostgresRepo(db *sql.DB) *AuthPostgresRepo {
	return &AuthPostgresRepo{Db: db}
}

func (repoap *AuthPostgresRepo) CreateUser(ctx context.Context, tx *sql.Tx, user *model.Person) *DBRepositoryResponse {
	traceid := ctx.Value("traceID").(string)
	var createdUserID uuid.UUID
	err := tx.QueryRowContext(ctx,
		"INSERT INTO UserZ (userid, username, useremail, userpassword) values ($1, $2, $3, $4) ON CONFLICT (useremail) DO NOTHING RETURNING userid;",
		user.Id, user.Name, user.Email, user.Password).Scan(&createdUserID)
	if err != nil {
		log.Printf("[ERROR] [UserManagement] [TraceID: %s]: CreateUser Error: %v", traceid, err)
		if errors.Is(err, sql.ErrNoRows) {
			return &DBRepositoryResponse{Success: false, Errors: erro.ErrorUniqueEmail, Type: erro.ClientErrorType}
		}
		return &DBRepositoryResponse{Success: false, Errors: erro.ErrorDbRepositoryError, Type: erro.ServerErrorType}
	}
	log.Printf("[INFO] [UserManagement] [TraceID: %s]: Successful create person!", traceid)
	return &DBRepositoryResponse{Success: true, UserId: createdUserID, Errors: nil}
}

func (repoap *AuthPostgresRepo) GetUser(ctx context.Context, useremail, userpassword string) *DBRepositoryResponse {
	traceid := ctx.Value("traceID").(string)
	var hashpass string
	var userId uuid.UUID
	err := repoap.Db.QueryRowContext(ctx, "SELECT userid, userpassword FROM userZ WHERE useremail = $1", useremail).Scan(&userId, &hashpass)
	if err != nil {
		log.Printf("[ERROR] [UserManagement] [TraceID: %s]: GetUser Error: %v", traceid, err)
		if errors.Is(err, sql.ErrNoRows) {
			return &DBRepositoryResponse{Success: false, Errors: erro.ErrorEmailNotRegister, Type: erro.ClientErrorType}
		}
		return &DBRepositoryResponse{Success: false, Errors: erro.ErrorDbRepositoryError, Type: erro.ServerErrorType}
	}
	err = bcrypt.CompareHashAndPassword([]byte(hashpass), []byte(userpassword))
	if err != nil {
		log.Printf("[ERROR] [UserManagement] [TraceID: %s]: CompareHashAndPassword Error: %v", traceid, err)
		return &DBRepositoryResponse{Success: false, Errors: erro.ErrorInvalidPassword, Type: erro.ClientErrorType}
	}
	log.Printf("[INFO] [UserManagement] [TraceID: %s]: Successful get person!", traceid)
	return &DBRepositoryResponse{Success: true, UserId: userId, Errors: nil}
}
func (repoap *AuthPostgresRepo) DeleteUser(ctx context.Context, tx *sql.Tx, userId uuid.UUID, password string) *DBRepositoryResponse {
	traceid := ctx.Value("traceID").(string)
	var hashpass string
	err := tx.QueryRowContext(ctx, "SELECT userpassword FROM userZ WHERE userid = $1", userId).Scan(&hashpass)
	if err != nil {
		log.Printf("[ERROR] [UserManagement] [TraceID: %s]: DeleteUser Error: %v", traceid, err)
		if errors.Is(err, sql.ErrNoRows) {
			return &DBRepositoryResponse{Success: false, Errors: erro.ErrorFoundUser, Type: erro.ClientErrorType}
		}
		return &DBRepositoryResponse{Success: false, Errors: erro.ErrorDbRepositoryError, Type: erro.ServerErrorType}
	}
	err = bcrypt.CompareHashAndPassword([]byte(hashpass), []byte(password))
	if err != nil {
		log.Printf("[ERROR] [UserManagement] [TraceID: %s]: CompareHashAndPassword Error: %v", traceid, err)
		return &DBRepositoryResponse{Success: false, Errors: erro.ErrorInvalidPassword, Type: erro.ClientErrorType}
	}
	_, err = tx.ExecContext(ctx, "DELETE FROM userZ where userId = $1", userId)
	if err != nil {
		log.Printf("[ERROR] [UserManagement] [TraceID: %s]: DeleteUser Error: %v", traceid, err)
		return &DBRepositoryResponse{Success: false, Errors: erro.ErrorDbRepositoryError, Type: erro.ServerErrorType}
	}
	log.Printf("[INFO] [UserManagement] [TraceID: %s]: Successful delete person!", traceid)
	return &DBRepositoryResponse{Success: true}
}
