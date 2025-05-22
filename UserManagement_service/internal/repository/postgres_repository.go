package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/erro"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/kafka"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/metrics"
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
	start := time.Now()
	defer func() {
		duration := time.Since(start).Seconds()
		metrics.UserDBQueryDuration.WithLabelValues("INSERT").Observe(duration)
	}()
	var place = "Repository-CreateUser"
	traceid := ctx.Value("traceID").(string)
	var createdUserID uuid.UUID
	err := tx.QueryRowContext(ctx,
		"INSERT INTO users (userid, username, useremail, userpassword) values ($1, $2, $3, $4) ON CONFLICT (useremail) DO NOTHING RETURNING userid;",
		user.Id, user.Name, user.Email, user.Password).Scan(&createdUserID)
	metrics.UserDBQueriesTotal.WithLabelValues("INSERT").Inc()
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			repoap.KafkaProducer.NewUserLog(kafka.LogLevelWarn, place, traceid, "Already registered email has been entered")
			metrics.UserDBErrorsTotal.WithLabelValues("ClientError", "INSERT").Inc()
			return &DBRepositoryResponse{Success: false, Errors: erro.ErrorUniqueEmail, Type: erro.ClientErrorType}
		}
		repoap.KafkaProducer.NewUserLog(kafka.LogLevelError, place, traceid, err.Error())
		metrics.UserDBErrorsTotal.WithLabelValues("InternalServerError", "INSERT").Inc()
		return &DBRepositoryResponse{Success: false, Errors: erro.ErrorDbRepositoryError, Type: erro.ServerErrorType}
	}
	repoap.KafkaProducer.NewUserLog(kafka.LogLevelInfo, place, traceid, "Successful create person")
	return &DBRepositoryResponse{Success: true, UserId: createdUserID, Errors: nil}
}

func (repoap *AuthPostgresRepo) GetUser(ctx context.Context, useremail, userpassword string) *DBRepositoryResponse {
	start := time.Now()
	defer func() {
		duration := time.Since(start).Seconds()
		metrics.UserDBQueryDuration.WithLabelValues("SELECT").Observe(duration)
	}()
	var place = "Repository-GetUser"
	traceid := ctx.Value("traceID").(string)
	var hashpass string
	var userId uuid.UUID
	err := repoap.Db.DB.QueryRowContext(ctx, "SELECT userid, userpassword FROM users WHERE useremail = $1", useremail).Scan(&userId, &hashpass)
	metrics.UserDBQueriesTotal.WithLabelValues("SELECT").Inc()
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			repoap.KafkaProducer.NewUserLog(kafka.LogLevelWarn, place, traceid, "Unregistered email has been entered")
			metrics.UserDBErrorsTotal.WithLabelValues("ClientError", "SELECT").Inc()
			return &DBRepositoryResponse{Success: false, Errors: erro.ErrorEmailNotRegister, Type: erro.ClientErrorType}
		}
		repoap.KafkaProducer.NewUserLog(kafka.LogLevelError, place, traceid, err.Error())
		metrics.UserDBErrorsTotal.WithLabelValues("InternalServerError", "SELECT").Inc()
		return &DBRepositoryResponse{Success: false, Errors: erro.ErrorDbRepositoryError, Type: erro.ServerErrorType}
	}
	err = bcrypt.CompareHashAndPassword([]byte(hashpass), []byte(userpassword))
	metrics.UserDBQueriesTotal.WithLabelValues("CompareHashAndPassword").Inc()
	if err != nil {
		repoap.KafkaProducer.NewUserLog(kafka.LogLevelWarn, place, traceid, "Incorrect password has been entered")
		metrics.UserDBErrorsTotal.WithLabelValues("ClientError", "CompareHashAndPassword").Inc()
		return &DBRepositoryResponse{Success: false, Errors: erro.ErrorInvalidPassword, Type: erro.ClientErrorType}
	}
	repoap.KafkaProducer.NewUserLog(kafka.LogLevelInfo, place, traceid, "Successful get person")
	return &DBRepositoryResponse{Success: true, UserId: userId, Errors: nil}
}
func (repoap *AuthPostgresRepo) DeleteUser(ctx context.Context, tx *sql.Tx, userId uuid.UUID, password string) *DBRepositoryResponse {
	start := time.Now()
	defer func() {
		duration := time.Since(start).Seconds()
		metrics.UserDBQueryDuration.WithLabelValues("SELECT-DELETE").Observe(duration)
	}()
	var place = "Repository-DeleteUser"
	traceid := ctx.Value("traceID").(string)
	var hashpass string
	err := tx.QueryRowContext(ctx, "SELECT userpassword FROM users WHERE userid = $1", userId).Scan(&hashpass)
	metrics.UserDBQueriesTotal.WithLabelValues("SELECT-DELETE").Inc()
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			repoap.KafkaProducer.NewUserLog(kafka.LogLevelWarn, place, traceid, "Unregistered email has been entered")
			metrics.UserDBErrorsTotal.WithLabelValues("ClientError", "SELECT-DELETE").Inc()
			return &DBRepositoryResponse{Success: false, Errors: erro.ErrorEmailNotRegister, Type: erro.ClientErrorType}
		}
		repoap.KafkaProducer.NewUserLog(kafka.LogLevelError, place, traceid, err.Error())
		metrics.UserDBErrorsTotal.WithLabelValues("InternalServerError", "SELECT-DELETE").Inc()
		return &DBRepositoryResponse{Success: false, Errors: erro.ErrorDbRepositoryError, Type: erro.ServerErrorType}
	}
	err = bcrypt.CompareHashAndPassword([]byte(hashpass), []byte(password))
	metrics.UserDBQueriesTotal.WithLabelValues("CompareHashAndPassword").Inc()
	if err != nil {
		repoap.KafkaProducer.NewUserLog(kafka.LogLevelWarn, place, traceid, "Incorrect password has been entered")
		metrics.UserDBErrorsTotal.WithLabelValues("ClientError", "CompareHashAndPassword").Inc()
		return &DBRepositoryResponse{Success: false, Errors: erro.ErrorInvalidPassword, Type: erro.ClientErrorType}
	}
	_, err = tx.ExecContext(ctx, "DELETE FROM users where userid = $1", userId)
	metrics.UserDBQueriesTotal.WithLabelValues("SELECT-DELETE").Inc()
	if err != nil {
		repoap.KafkaProducer.NewUserLog(kafka.LogLevelError, place, traceid, err.Error())
		metrics.UserDBErrorsTotal.WithLabelValues("InternalServerError", "SELECT-DELETE").Inc()
		return &DBRepositoryResponse{Success: false, Errors: erro.ErrorDbRepositoryError, Type: erro.ServerErrorType}
	}
	repoap.KafkaProducer.NewUserLog(kafka.LogLevelInfo, place, traceid, "Successful delete person")
	return &DBRepositoryResponse{Success: true}
}
