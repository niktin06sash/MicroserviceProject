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
	requestid, ok := ctx.Value("requestID").(string)
	if !ok {
		log.Println("[ERROR] [UserManagement] CreateUser Error: Request ID not found in context")
		return &DBRepositoryResponse{
			Success: false,
			Errors:  erro.ErrorMissingRequestID,
		}
	}
	var createdUserID uuid.UUID
	err := tx.QueryRowContext(ctx,
		"INSERT INTO UserZ (userid, username, useremail, userpassword) values ($1, $2, $3, $4) ON CONFLICT (useremail) DO NOTHING RETURNING userid;",
		user.Id, user.Name, user.Email, user.Password).Scan(&createdUserID)

	if err != nil {
		log.Printf("[ERROR] [UserManagement] [RequestID: %s]: CreateUser Error: %v", requestid, err)
		if errors.Is(err, sql.ErrNoRows) {
			return &DBRepositoryResponse{Success: false, Errors: erro.ErrorUniqueEmail}
		}
		return &DBRepositoryResponse{Success: false, Errors: erro.ErrorDbRepositoryError}
	}

	log.Printf("[INFO] [UserManagement] [RequestID: %s]: Successful create person!", requestid)
	return &DBRepositoryResponse{Success: true, UserId: createdUserID, Errors: nil}
}

func (repoap *AuthPostgresRepo) GetUser(ctx context.Context, useremail, userpassword string) *DBRepositoryResponse {
	requestid, ok := ctx.Value("requestID").(string)
	if !ok {
		log.Println("[ERROR] [UserManagement] GetUser Error: Request ID not found in context")
		return &DBRepositoryResponse{
			Success: false,
			Errors:  erro.ErrorMissingRequestID,
		}
	}
	var hashpass string
	var userId uuid.UUID
	err := repoap.Db.QueryRowContext(ctx, "SELECT userid, userpassword FROM userZ WHERE useremail = $1", useremail).Scan(&userId, &hashpass)

	if err != nil {
		log.Printf("[ERROR] [UserManagement] [RequestID: %s]: GetUser Error: %v", requestid, err)
		if errors.Is(err, sql.ErrNoRows) {
			return &DBRepositoryResponse{Success: false, Errors: erro.ErrorEmailNotRegister}
		}
		return &DBRepositoryResponse{Success: false, Errors: erro.ErrorDbRepositoryError}
	}

	err = bcrypt.CompareHashAndPassword([]byte(hashpass), []byte(userpassword))
	if err != nil {
		log.Printf("[ERROR] [UserManagement] [RequestID: %s]: CompareHashAndPassword Error: %v", requestid, err)
		return &DBRepositoryResponse{Success: false, Errors: erro.ErrorInvalidPassword}
	}

	log.Printf("[INFO] [UserManagement] [RequestID: %s]: Successful get person!", requestid)
	return &DBRepositoryResponse{Success: true, UserId: userId, Errors: nil}
}
func (repoap *AuthPostgresRepo) DeleteUser(ctx context.Context, tx *sql.Tx, userId uuid.UUID, password string) *DBRepositoryResponse {
	requestid, ok := ctx.Value("requestID").(string)
	if !ok {
		log.Println("[ERROR] [UserManagement] DeleteUser Error: Request ID not found in context")
		return &DBRepositoryResponse{
			Success: false,
			Errors:  erro.ErrorMissingRequestID,
		}
	}
	var hashpass string
	err := tx.QueryRowContext(ctx, "SELECT userpassword FROM userZ WHERE userid = $1", userId).Scan(&hashpass)

	if err != nil {
		log.Printf("[ERROR] [UserManagement] [RequestID: %s]: DeleteUser Error: %v", requestid, err)
		if errors.Is(err, sql.ErrNoRows) {
			return &DBRepositoryResponse{Success: false, Errors: erro.ErrorFoundUser}
		}
		return &DBRepositoryResponse{Success: false, Errors: erro.ErrorDbRepositoryError}
	}
	err = bcrypt.CompareHashAndPassword([]byte(hashpass), []byte(password))
	if err != nil {
		log.Printf("[ERROR] [UserManagement] [RequestID: %s]: CompareHashAndPassword Error: %v", requestid, err)
		return &DBRepositoryResponse{Success: false, Errors: erro.ErrorInvalidPassword}
	}
	_, err = tx.ExecContext(ctx, "DELETE FROM userZ where userId = $1", userId)
	if err != nil {
		log.Printf("[ERROR] [UserManagement] [RequestID: %s]: DeleteUser Error: %v", requestid, err)
		return &DBRepositoryResponse{Success: false, Errors: erro.ErrorDbRepositoryError}
	}
	log.Printf("[INFO] [UserManagement] [RequestID: %s]: Successful delete person!", requestid)
	return &DBRepositoryResponse{Success: true}
}
