package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
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

func (repoap *AuthPostgresRepo) CreateUser(ctx context.Context, tx *sql.Tx, user *model.User) *DBRepositoryResponse {
	const place = CreateUser
	start := time.Now()
	defer func() {
		metrics.UserDBQueriesTotal.WithLabelValues(place).Inc()
		duration := time.Since(start).Seconds()
		metrics.UserDBQueryDuration.WithLabelValues(place).Observe(duration)
	}()
	traceid := ctx.Value("traceID").(string)
	var createdUserID uuid.UUID
	err := tx.QueryRowContext(ctx,
		"INSERT INTO users (userid, username, useremail, userpassword) VALUES ($1, $2, $3, $4) ON CONFLICT (useremail) DO NOTHING RETURNING userid;",
		user.Id, user.Name, user.Email, user.Password).Scan(&createdUserID)
	metrics.UserDBQueriesTotal.WithLabelValues("INSERT").Inc()
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			repoap.KafkaProducer.NewUserLog(kafka.LogLevelWarn, place, traceid, "Already registered email has been entered")
			metrics.UserDBErrorsTotal.WithLabelValues(erro.ClientErrorType, "INSERT").Inc()
			return &DBRepositoryResponse{Success: false, Errors: &erro.ErrorResponse{Type: erro.ClientErrorType, Message: erro.ErrorUniqueEmailConst}}
		}
		repoap.KafkaProducer.NewUserLog(kafka.LogLevelError, place, traceid, err.Error())
		metrics.UserDBErrorsTotal.WithLabelValues(erro.ServerErrorType, "INSERT").Inc()
		return &DBRepositoryResponse{Success: false, Errors: &erro.ErrorResponse{Type: erro.ServerErrorType, Message: erro.UserServiceUnavalaible}}
	}
	repoap.KafkaProducer.NewUserLog(kafka.LogLevelInfo, place, traceid, "Successful create person")
	return &DBRepositoryResponse{Success: true, Data: map[string]any{"userID": createdUserID}, Errors: nil}
}
func (repoap *AuthPostgresRepo) AuthenticateUser(ctx context.Context, useremail, userpassword string) *DBRepositoryResponse {
	const place = AuthenticateUser
	start := time.Now()
	defer func() {
		metrics.UserDBQueriesTotal.WithLabelValues(place).Inc()
		duration := time.Since(start).Seconds()
		metrics.UserDBQueryDuration.WithLabelValues(place).Observe(duration)
	}()
	traceid := ctx.Value("traceID").(string)
	var hashpass string
	var userId uuid.UUID
	err := repoap.Db.DB.QueryRowContext(ctx, "SELECT userid, userpassword FROM users WHERE useremail = $1", useremail).Scan(&userId, &hashpass)
	metrics.UserDBQueriesTotal.WithLabelValues("SELECT").Inc()
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			repoap.KafkaProducer.NewUserLog(kafka.LogLevelWarn, place, traceid, "Unregistered email has been entered")
			metrics.UserDBErrorsTotal.WithLabelValues(erro.ClientErrorType, "SELECT").Inc()
			return &DBRepositoryResponse{Success: false, Errors: &erro.ErrorResponse{Type: erro.ClientErrorType, Message: erro.ErrorEmailNotRegisterConst}}
		}
		repoap.KafkaProducer.NewUserLog(kafka.LogLevelError, place, traceid, err.Error())
		metrics.UserDBErrorsTotal.WithLabelValues(erro.ServerErrorType, "SELECT").Inc()
		return &DBRepositoryResponse{Success: false, Errors: &erro.ErrorResponse{Type: erro.ServerErrorType, Message: erro.UserServiceUnavalaible}}
	}
	err = bcrypt.CompareHashAndPassword([]byte(hashpass), []byte(userpassword))
	metrics.UserDBQueriesTotal.WithLabelValues("CompareHashAndPassword").Inc()
	if err != nil {
		repoap.KafkaProducer.NewUserLog(kafka.LogLevelWarn, place, traceid, "Incorrect password has been entered")
		metrics.UserDBErrorsTotal.WithLabelValues(erro.ClientErrorType, "CompareHashAndPassword").Inc()
		return &DBRepositoryResponse{Success: false, Errors: &erro.ErrorResponse{Type: erro.ClientErrorType, Message: erro.ErrorInvalidPasswordConst}}
	}

	repoap.KafkaProducer.NewUserLog(kafka.LogLevelInfo, place, traceid, "Successful authenticate person")
	return &DBRepositoryResponse{Success: true, Data: map[string]any{"userID": userId}, Errors: nil}
}
func (repoap *AuthPostgresRepo) DeleteUser(ctx context.Context, tx *sql.Tx, userId uuid.UUID, password string) *DBRepositoryResponse {
	const place = DeleteUser
	start := time.Now()
	defer func() {
		metrics.UserDBQueriesTotal.WithLabelValues(place).Inc()
		duration := time.Since(start).Seconds()
		metrics.UserDBQueryDuration.WithLabelValues(place).Observe(duration)
	}()
	traceid := ctx.Value("traceID").(string)
	var hashpass string
	err := tx.QueryRowContext(ctx, "SELECT userpassword FROM users WHERE userid = $1", userId).Scan(&hashpass)
	metrics.UserDBQueriesTotal.WithLabelValues("SELECT").Inc()
	if err != nil {
		repoap.KafkaProducer.NewUserLog(kafka.LogLevelError, place, traceid, err.Error())
		metrics.UserDBErrorsTotal.WithLabelValues(erro.ServerErrorType, "SELECT").Inc()
		return &DBRepositoryResponse{Success: false, Errors: &erro.ErrorResponse{Type: erro.ServerErrorType, Message: erro.UserServiceUnavalaible}}
	}
	err = bcrypt.CompareHashAndPassword([]byte(hashpass), []byte(password))
	metrics.UserDBQueriesTotal.WithLabelValues("CompareHashAndPassword").Inc()
	if err != nil {
		repoap.KafkaProducer.NewUserLog(kafka.LogLevelWarn, place, traceid, "Incorrect password has been entered")
		metrics.UserDBErrorsTotal.WithLabelValues(erro.ClientErrorType, "CompareHashAndPassword").Inc()
		return &DBRepositoryResponse{Success: false, Errors: &erro.ErrorResponse{Type: erro.ClientErrorType, Message: erro.ErrorInvalidPasswordConst}}
	}
	_, err = tx.ExecContext(ctx, "DELETE FROM users where userid = $1", userId)
	metrics.UserDBQueriesTotal.WithLabelValues("DELETE").Inc()
	if err != nil {
		repoap.KafkaProducer.NewUserLog(kafka.LogLevelError, place, traceid, err.Error())
		metrics.UserDBErrorsTotal.WithLabelValues(erro.ServerErrorType, "DELETE").Inc()
		return &DBRepositoryResponse{Success: false, Errors: &erro.ErrorResponse{Type: erro.ServerErrorType, Message: erro.UserServiceUnavalaible}}
	}
	repoap.KafkaProducer.NewUserLog(kafka.LogLevelInfo, place, traceid, "Successful delete person")
	return &DBRepositoryResponse{Success: true}
}
func (repoap *AuthPostgresRepo) UpdateUserName(ctx context.Context, userId uuid.UUID, name string) *DBRepositoryResponse {
	const place = UpdateName
	start := time.Now()
	defer func() {
		metrics.UserDBQueriesTotal.WithLabelValues(place).Inc()
		duration := time.Since(start).Seconds()
		metrics.UserDBQueryDuration.WithLabelValues(place).Observe(duration)
	}()
	traceid := ctx.Value("traceID").(string)
	_, err := repoap.Db.DB.ExecContext(ctx, "UPDATE users SET username = $1 where userid = $2", name, userId)
	if err != nil {
		repoap.KafkaProducer.NewUserLog(kafka.LogLevelError, place, traceid, err.Error())
		metrics.UserDBErrorsTotal.WithLabelValues(erro.ServerErrorType, "UPDATE").Inc()
		return &DBRepositoryResponse{Success: false, Errors: &erro.ErrorResponse{Type: erro.ServerErrorType, Message: erro.UserServiceUnavalaible}}
	}
	repoap.KafkaProducer.NewUserLog(kafka.LogLevelInfo, place, traceid, "Successful update username")
	return &DBRepositoryResponse{Success: true}
}
func (repoap *AuthPostgresRepo) UpdateUserEmail(ctx context.Context, userId uuid.UUID, email string, password string) *DBRepositoryResponse {
	const place = UpdateEmail
	start := time.Now()
	defer func() {
		metrics.UserDBQueriesTotal.WithLabelValues(place).Inc()
		duration := time.Since(start).Seconds()
		metrics.UserDBQueryDuration.WithLabelValues(place).Observe(duration)
	}()
	traceid := ctx.Value("traceID").(string)
	var hashpass string
	err := repoap.Db.DB.QueryRowContext(ctx, "SELECT userpassword FROM users WHERE userid = $1", userId).Scan(&hashpass)
	metrics.UserDBQueriesTotal.WithLabelValues("SELECT").Inc()
	if err != nil {
		repoap.KafkaProducer.NewUserLog(kafka.LogLevelError, place, traceid, err.Error())
		metrics.UserDBErrorsTotal.WithLabelValues(erro.ServerErrorType, "SELECT").Inc()
		return &DBRepositoryResponse{Success: false, Errors: &erro.ErrorResponse{Type: erro.ServerErrorType, Message: erro.UserServiceUnavalaible}}
	}
	err = bcrypt.CompareHashAndPassword([]byte(hashpass), []byte(password))
	metrics.UserDBQueriesTotal.WithLabelValues("CompareHashAndPassword").Inc()
	if err != nil {
		repoap.KafkaProducer.NewUserLog(kafka.LogLevelWarn, place, traceid, "Incorrect password has been entered")
		metrics.UserDBErrorsTotal.WithLabelValues(erro.ClientErrorType, "CompareHashAndPassword").Inc()
		return &DBRepositoryResponse{Success: false, Errors: &erro.ErrorResponse{Type: erro.ClientErrorType, Message: erro.ErrorInvalidPasswordConst}}
	}
	var count int
	err = repoap.Db.DB.QueryRowContext(ctx, "SELECT COUNT(*) FROM users WHERE useremail = $1", email).Scan(&count)
	if err != nil {
		repoap.KafkaProducer.NewUserLog(kafka.LogLevelError, place, traceid, err.Error())
		metrics.UserDBErrorsTotal.WithLabelValues(erro.ServerErrorType, "SELECT").Inc()
		return &DBRepositoryResponse{Success: false, Errors: &erro.ErrorResponse{Type: erro.ServerErrorType, Message: erro.UserServiceUnavalaible}}
	}
	if count > 0 {
		repoap.KafkaProducer.NewUserLog(kafka.LogLevelWarn, place, traceid, "Already registered email has been entered")
		metrics.UserDBErrorsTotal.WithLabelValues(erro.ClientErrorType, "SELECT").Inc()
		return &DBRepositoryResponse{Success: false, Errors: &erro.ErrorResponse{Type: erro.ClientErrorType, Message: erro.ErrorUniqueEmailConst}}
	}
	_, err = repoap.Db.DB.ExecContext(ctx, "UPDATE users SET useremail = $1 where userid = $2", email, userId)
	metrics.UserDBQueriesTotal.WithLabelValues("UPDATE").Inc()
	if err != nil {
		repoap.KafkaProducer.NewUserLog(kafka.LogLevelError, place, traceid, err.Error())
		metrics.UserDBErrorsTotal.WithLabelValues(erro.ServerErrorType, "UPDATE").Inc()
		return &DBRepositoryResponse{Success: false, Errors: &erro.ErrorResponse{Type: erro.ServerErrorType, Message: erro.UserServiceUnavalaible}}
	}
	repoap.KafkaProducer.NewUserLog(kafka.LogLevelInfo, place, traceid, "Successful update useremail")
	return &DBRepositoryResponse{Success: true}
}
func (repoap *AuthPostgresRepo) UpdateUserPassword(ctx context.Context, userId uuid.UUID, lastpassword string, newpassword string) *DBRepositoryResponse {
	const place = UpdatePassword
	start := time.Now()
	defer func() {
		metrics.UserDBQueriesTotal.WithLabelValues(place).Inc()
		duration := time.Since(start).Seconds()
		metrics.UserDBQueryDuration.WithLabelValues(place).Observe(duration)
	}()
	traceid := ctx.Value("traceID").(string)
	var hashpass string
	err := repoap.Db.DB.QueryRowContext(ctx, "SELECT userpassword FROM users WHERE userid = $1", userId).Scan(&hashpass)
	metrics.UserDBQueriesTotal.WithLabelValues("SELECT").Inc()
	if err != nil {
		repoap.KafkaProducer.NewUserLog(kafka.LogLevelError, place, traceid, err.Error())
		metrics.UserDBErrorsTotal.WithLabelValues(erro.ServerErrorType, "SELECT").Inc()
		return &DBRepositoryResponse{Success: false, Errors: &erro.ErrorResponse{Type: erro.ServerErrorType, Message: erro.UserServiceUnavalaible}}
	}
	err = bcrypt.CompareHashAndPassword([]byte(hashpass), []byte(lastpassword))
	metrics.UserDBQueriesTotal.WithLabelValues("CompareHashAndPassword").Inc()
	if err != nil {
		repoap.KafkaProducer.NewUserLog(kafka.LogLevelWarn, place, traceid, "Incorrect password has been entered")
		metrics.UserDBErrorsTotal.WithLabelValues(erro.ClientErrorType, "CompareHashAndPassword").Inc()
		return &DBRepositoryResponse{Success: false, Errors: &erro.ErrorResponse{Type: erro.ClientErrorType, Message: erro.ErrorInvalidPasswordConst}}
	}
	hashnewpass, err := bcrypt.GenerateFromPassword([]byte(newpassword), bcrypt.DefaultCost)
	metrics.UserDBQueriesTotal.WithLabelValues("GenerateHashPassword").Inc()
	if err != nil {
		fmterr := fmt.Sprintf("Generate HashPassword Error: %v", err)
		repoap.KafkaProducer.NewUserLog(kafka.LogLevelError, place, traceid, fmterr)
		metrics.UserDBErrorsTotal.WithLabelValues(erro.ServerErrorType, "GenerateHashPassword").Inc()
		return &DBRepositoryResponse{Success: false, Errors: &erro.ErrorResponse{Type: erro.ServerErrorType, Message: erro.UserServiceUnavalaible}}
	}
	_, err = repoap.Db.DB.ExecContext(ctx, "UPDATE users SET userpassword = $1 where userid = $2", hashnewpass, userId)
	metrics.UserDBQueriesTotal.WithLabelValues("UPDATE").Inc()
	if err != nil {
		repoap.KafkaProducer.NewUserLog(kafka.LogLevelError, place, traceid, err.Error())
		metrics.UserDBErrorsTotal.WithLabelValues(erro.ServerErrorType, "UPDATE").Inc()
		return &DBRepositoryResponse{Success: false, Errors: &erro.ErrorResponse{Type: erro.ServerErrorType, Message: erro.UserServiceUnavalaible}}
	}
	repoap.KafkaProducer.NewUserLog(kafka.LogLevelInfo, place, traceid, "Successful update userpassword")
	return &DBRepositoryResponse{Success: true}
}
func (repoap *AuthPostgresRepo) GetMyProfile(ctx context.Context, userid uuid.UUID) *DBRepositoryResponse {
	const place = GetMyProfile
	start := time.Now()
	defer func() {
		metrics.UserDBQueriesTotal.WithLabelValues(place).Inc()
		duration := time.Since(start).Seconds()
		metrics.UserDBQueryDuration.WithLabelValues(place).Observe(duration)
	}()
	traceid := ctx.Value("traceID").(string)
	var email string
	var name string
	err := repoap.Db.DB.QueryRowContext(ctx, "SELECT useremail, username FROM users WHERE userid = $1", userid).Scan(&email, &name)
	metrics.UserDBQueriesTotal.WithLabelValues("SELECT").Inc()
	if err != nil {
		repoap.KafkaProducer.NewUserLog(kafka.LogLevelError, place, traceid, err.Error())
		metrics.UserDBErrorsTotal.WithLabelValues(erro.ServerErrorType, "SELECT").Inc()
		return &DBRepositoryResponse{Success: false, Errors: &erro.ErrorResponse{Type: erro.ServerErrorType, Message: erro.UserServiceUnavalaible}}
	}
	repoap.KafkaProducer.NewUserLog(kafka.LogLevelInfo, place, traceid, "Successful get my profile")
	return &DBRepositoryResponse{Success: true, Data: map[string]any{"userID": userid.String(), "userEmail": email, "userName": name}, Errors: nil}
}
