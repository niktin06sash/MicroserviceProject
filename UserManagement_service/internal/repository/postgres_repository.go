package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/erro"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/kafka"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/model"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthPostgresRepo struct {
	Db            *DBObject
	KafkaProducer kafka.KafkaProducerService
}

func NewAuthPostgresRepo(db *DBObject, kafkaprod kafka.KafkaProducerService) *AuthPostgresRepo {
	return &AuthPostgresRepo{Db: db, KafkaProducer: kafkaprod}
}

func (repoap *AuthPostgresRepo) CreateUser(ctx context.Context, tx *sql.Tx, user *model.Person) *DBRepositoryResponse {
	var place = "Repository-CreateUser"
	traceid := ctx.Value("traceID").(string)
	var createdUserID uuid.UUID
	err := tx.QueryRowContext(ctx,
		"INSERT INTO UserZ (userid, username, useremail, userpassword) values ($1, $2, $3, $4) ON CONFLICT (useremail) DO NOTHING RETURNING userid;",
		user.Id, user.Name, user.Email, user.Password).Scan(&createdUserID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			repoap.KafkaProducer.NewUserLog(kafka.LogLevelWarn, place, traceid, "Already registered email has been entered")
			return &DBRepositoryResponse{Success: false, Errors: erro.ErrorUniqueEmail, Type: erro.ClientErrorType}
		}
		repoap.KafkaProducer.NewUserLog(kafka.LogLevelError, place, traceid, err.Error())
		return &DBRepositoryResponse{Success: false, Errors: err, Type: erro.ServerErrorType}
	}
	repoap.KafkaProducer.NewUserLog(kafka.LogLevelInfo, place, traceid, "Successful create person")
	return &DBRepositoryResponse{Success: true, UserId: createdUserID, Errors: nil}
}

func (repoap *AuthPostgresRepo) GetUser(ctx context.Context, useremail, userpassword string) *DBRepositoryResponse {
	var place = "Repository-GetUser"
	traceid := ctx.Value("traceID").(string)
	var hashpass string
	var userId uuid.UUID
	err := repoap.Db.DB.QueryRowContext(ctx, "SELECT userid, userpassword FROM userZ WHERE useremail = $1", useremail).Scan(&userId, &hashpass)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			repoap.KafkaProducer.NewUserLog(kafka.LogLevelWarn, place, traceid, "Unregistered email has been entered")
			return &DBRepositoryResponse{Success: false, Errors: erro.ErrorEmailNotRegister, Type: erro.ClientErrorType}
		}
		repoap.KafkaProducer.NewUserLog(kafka.LogLevelError, place, traceid, err.Error())
		return &DBRepositoryResponse{Success: false, Errors: err, Type: erro.ServerErrorType}
	}
	err = bcrypt.CompareHashAndPassword([]byte(hashpass), []byte(userpassword))
	if err != nil {
		repoap.KafkaProducer.NewUserLog(kafka.LogLevelWarn, place, traceid, "Incorrect password has been entered")
		return &DBRepositoryResponse{Success: false, Errors: erro.ErrorInvalidPassword, Type: erro.ClientErrorType}
	}
	repoap.KafkaProducer.NewUserLog(kafka.LogLevelInfo, place, traceid, "Successful get person")
	return &DBRepositoryResponse{Success: true, UserId: userId, Errors: nil}
}
func (repoap *AuthPostgresRepo) DeleteUser(ctx context.Context, tx *sql.Tx, userId uuid.UUID, password string) *DBRepositoryResponse {
	var place = "Repository-DeleteUser"
	traceid := ctx.Value("traceID").(string)
	var hashpass string
	err := tx.QueryRowContext(ctx, "SELECT userpassword FROM userZ WHERE userid = $1", userId).Scan(&hashpass)
	if err != nil {
		repoap.KafkaProducer.NewUserLog(kafka.LogLevelError, place, traceid, err.Error())
		return &DBRepositoryResponse{Success: false, Errors: err, Type: erro.ServerErrorType}
	}
	err = bcrypt.CompareHashAndPassword([]byte(hashpass), []byte(password))
	if err != nil {
		repoap.KafkaProducer.NewUserLog(kafka.LogLevelWarn, place, traceid, "Incorrect password has been entered")
		return &DBRepositoryResponse{Success: false, Errors: erro.ErrorInvalidPassword, Type: erro.ClientErrorType}
	}
	_, err = tx.ExecContext(ctx, "DELETE FROM userZ where userid = $1", userId)
	if err != nil {
		repoap.KafkaProducer.NewUserLog(kafka.LogLevelError, place, traceid, err.Error())
		return &DBRepositoryResponse{Success: false, Errors: err, Type: erro.ServerErrorType}
	}
	repoap.KafkaProducer.NewUserLog(kafka.LogLevelInfo, place, traceid, "Successful delete person")
	return &DBRepositoryResponse{Success: true}
}
