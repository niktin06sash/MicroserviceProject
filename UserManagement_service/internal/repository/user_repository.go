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

func (repoap *AuthPostgresRepo) CreateUser(ctx context.Context, tx *sql.Tx, user *model.User) *RepositoryResponse {
	const place = CreateUser
	start := time.Now()
	defer deferMetrics(place, start)
	traceid := ctx.Value("traceID").(string)
	var createdUserID uuid.UUID
	err := tx.QueryRowContext(ctx,
		fmt.Sprintf("INSERT INTO %s (%s, %s, %s, %s) VALUES ($1, $2, $3, $4) ON CONFLICT (%s) DO NOTHING RETURNING %s;",
			KeyUserTable, KeyUserID, KeyUserName, KeyUserEmail, KeyUserPassword, KeyUserEmail, KeyUserID),
		user.Id, user.Name, user.Email, user.Password).Scan(&createdUserID)
	metrics.UserDBQueriesTotal.WithLabelValues("INSERT").Inc()
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			repoap.KafkaProducer.NewUserLog(kafka.LogLevelWarn, place, traceid, "Already registered email has been entered")
			metrics.UserDBErrorsTotal.WithLabelValues(erro.ClientErrorType, "INSERT").Inc()
			return &RepositoryResponse{Success: false, Errors: &erro.ErrorResponse{Type: erro.ClientErrorType, Message: erro.ErrorUniqueEmailConst}}
		}
		repoap.KafkaProducer.NewUserLog(kafka.LogLevelError, place, traceid, err.Error())
		metrics.UserDBErrorsTotal.WithLabelValues(erro.ServerErrorType, "INSERT").Inc()
		return &RepositoryResponse{Success: false, Errors: &erro.ErrorResponse{Type: erro.ServerErrorType, Message: erro.UserServiceUnavalaible}}
	}
	repoap.KafkaProducer.NewUserLog(kafka.LogLevelInfo, place, traceid, "Successful create person")
	return &RepositoryResponse{Success: true, Data: map[string]any{KeyUserID: createdUserID.String()}, Errors: nil}
}
func (repoap *AuthPostgresRepo) AuthenticateUser(ctx context.Context, useremail, userpassword string) *RepositoryResponse {
	const place = AuthenticateUser
	start := time.Now()
	defer deferMetrics(place, start)
	traceid := ctx.Value("traceID").(string)
	var hashpass string
	var userId uuid.UUID
	err := repoap.Db.DB.QueryRowContext(ctx, fmt.Sprintf("SELECT %s, %s FROM %s WHERE %s = $1", KeyUserID, KeyUserPassword, KeyUserTable, KeyUserEmail), useremail).Scan(&userId, &hashpass)
	metrics.UserDBQueriesTotal.WithLabelValues("SELECT").Inc()
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			repoap.KafkaProducer.NewUserLog(kafka.LogLevelWarn, place, traceid, "Unregistered email has been entered")
			metrics.UserDBErrorsTotal.WithLabelValues(erro.ClientErrorType, "SELECT").Inc()
			return &RepositoryResponse{Success: false, Errors: &erro.ErrorResponse{Type: erro.ClientErrorType, Message: erro.ErrorEmailNotRegisterConst}}
		}
		repoap.KafkaProducer.NewUserLog(kafka.LogLevelError, place, traceid, err.Error())
		metrics.UserDBErrorsTotal.WithLabelValues(erro.ServerErrorType, "SELECT").Inc()
		return &RepositoryResponse{Success: false, Errors: &erro.ErrorResponse{Type: erro.ServerErrorType, Message: erro.UserServiceUnavalaible}}
	}
	err = bcrypt.CompareHashAndPassword([]byte(hashpass), []byte(userpassword))
	metrics.UserDBQueriesTotal.WithLabelValues("CompareHashAndPassword").Inc()
	if err != nil {
		repoap.KafkaProducer.NewUserLog(kafka.LogLevelWarn, place, traceid, "Incorrect password has been entered")
		metrics.UserDBErrorsTotal.WithLabelValues(erro.ClientErrorType, "CompareHashAndPassword").Inc()
		return &RepositoryResponse{Success: false, Errors: &erro.ErrorResponse{Type: erro.ClientErrorType, Message: erro.ErrorInvalidPasswordConst}}
	}

	repoap.KafkaProducer.NewUserLog(kafka.LogLevelInfo, place, traceid, "Successful authenticate person")
	return &RepositoryResponse{Success: true, Data: map[string]any{KeyUserID: userId.String()}, Errors: nil}
}
func (repoap *AuthPostgresRepo) DeleteUser(ctx context.Context, tx *sql.Tx, userId uuid.UUID, password string) *RepositoryResponse {
	const place = DeleteUser
	start := time.Now()
	defer deferMetrics(place, start)
	traceid := ctx.Value("traceID").(string)
	var hashpass string
	err := tx.QueryRowContext(ctx, fmt.Sprintf("SELECT %s FROM %s WHERE %s = $1", KeyUserPassword, KeyUserTable, KeyUserID), userId).Scan(&hashpass)
	metrics.UserDBQueriesTotal.WithLabelValues("SELECT").Inc()
	if err != nil {
		repoap.KafkaProducer.NewUserLog(kafka.LogLevelError, place, traceid, err.Error())
		metrics.UserDBErrorsTotal.WithLabelValues(erro.ServerErrorType, "SELECT").Inc()
		return &RepositoryResponse{Success: false, Errors: &erro.ErrorResponse{Type: erro.ServerErrorType, Message: erro.UserServiceUnavalaible}}
	}
	err = bcrypt.CompareHashAndPassword([]byte(hashpass), []byte(password))
	metrics.UserDBQueriesTotal.WithLabelValues("CompareHashAndPassword").Inc()
	if err != nil {
		repoap.KafkaProducer.NewUserLog(kafka.LogLevelWarn, place, traceid, "Incorrect password has been entered")
		metrics.UserDBErrorsTotal.WithLabelValues(erro.ClientErrorType, "CompareHashAndPassword").Inc()
		return &RepositoryResponse{Success: false, Errors: &erro.ErrorResponse{Type: erro.ClientErrorType, Message: erro.ErrorInvalidPasswordConst}}
	}
	_, err = tx.ExecContext(ctx, fmt.Sprintf("DELETE FROM %s where %s = $1", KeyUserTable, KeyUserID), userId)
	metrics.UserDBQueriesTotal.WithLabelValues("DELETE").Inc()
	if err != nil {
		repoap.KafkaProducer.NewUserLog(kafka.LogLevelError, place, traceid, err.Error())
		metrics.UserDBErrorsTotal.WithLabelValues(erro.ServerErrorType, "DELETE").Inc()
		return &RepositoryResponse{Success: false, Errors: &erro.ErrorResponse{Type: erro.ServerErrorType, Message: erro.UserServiceUnavalaible}}
	}
	repoap.KafkaProducer.NewUserLog(kafka.LogLevelInfo, place, traceid, "Successful delete person")
	return &RepositoryResponse{Success: true}
}
func (repoap *AuthPostgresRepo) UpdateUserData(ctx context.Context, tx *sql.Tx, userId uuid.UUID, updateType string, args ...interface{}) *RepositoryResponse {
	switch updateType {
	case "name":
		name := args[0].(string)
		return repoap.updateUserName(ctx, tx, userId, name)
	case "email":
		email := args[0].(string)
		password := args[1].(string)
		return repoap.updateUserEmail(ctx, tx, userId, email, password)
	case "password":
		lastpassword := args[0].(string)
		newpassword := args[1].(string)
		return repoap.updateUserPassword(ctx, tx, userId, lastpassword, newpassword)
	}
	return &RepositoryResponse{Success: false, Errors: &erro.ErrorResponse{Message: erro.ErrorInvalidDinamicParameter, Type: erro.ClientErrorType}}
}
func (repoap *AuthPostgresRepo) GetMyProfile(ctx context.Context, userid uuid.UUID) *RepositoryResponse {
	const place = GetMyProfile
	start := time.Now()
	defer deferMetrics(place, start)
	traceid := ctx.Value("traceID").(string)
	var email string
	var name string
	err := repoap.Db.DB.QueryRowContext(ctx, fmt.Sprintf("SELECT %s, %s FROM %s WHERE %s = $1", KeyUserEmail, KeyUserName, KeyUserTable, KeyUserID), userid).Scan(&email, &name)
	metrics.UserDBQueriesTotal.WithLabelValues("SELECT").Inc()
	if err != nil {
		repoap.KafkaProducer.NewUserLog(kafka.LogLevelError, place, traceid, err.Error())
		metrics.UserDBErrorsTotal.WithLabelValues(erro.ServerErrorType, "SELECT").Inc()
		return &RepositoryResponse{Success: false, Errors: &erro.ErrorResponse{Type: erro.ServerErrorType, Message: erro.UserServiceUnavalaible}}
	}
	repoap.KafkaProducer.NewUserLog(kafka.LogLevelInfo, place, traceid, "Successful get my profile")
	return &RepositoryResponse{Success: true, Data: map[string]any{KeyUserID: userid.String(), KeyUserEmail: email, KeyUserName: name}, Errors: nil}
}
func (repoap *AuthPostgresRepo) GetProfileById(ctx context.Context, getid uuid.UUID) *RepositoryResponse {
	const place = GetProfileById
	start := time.Now()
	defer deferMetrics(place, start)
	traceid := ctx.Value("traceID").(string)
	var email string
	var name string
	err := repoap.Db.DB.QueryRowContext(ctx, fmt.Sprintf("SELECT %s, %s FROM %s WHERE %s = $1", KeyUserEmail, KeyUserName, KeyUserTable, KeyUserID), getid).Scan(&email, &name)
	metrics.UserDBQueriesTotal.WithLabelValues("SELECT").Inc()
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			repoap.KafkaProducer.NewUserLog(kafka.LogLevelWarn, place, traceid, "Unregistered id has been entered")
			metrics.UserDBErrorsTotal.WithLabelValues(erro.ClientErrorType, "SELECT").Inc()
			return &RepositoryResponse{Success: false, Errors: &erro.ErrorResponse{Type: erro.ClientErrorType, Message: "Unregistered id has been entered"}}
		}
		repoap.KafkaProducer.NewUserLog(kafka.LogLevelError, place, traceid, err.Error())
		metrics.UserDBErrorsTotal.WithLabelValues(erro.ServerErrorType, "SELECT").Inc()
		return &RepositoryResponse{Success: false, Errors: &erro.ErrorResponse{Type: erro.ServerErrorType, Message: erro.UserServiceUnavalaible}}
	}
	repoap.KafkaProducer.NewUserLog(kafka.LogLevelInfo, place, traceid, "Successful get profile by id")
	return &RepositoryResponse{Success: true, Data: map[string]any{KeyUserID: getid.String(), KeyUserEmail: email, KeyUserName: name}, Errors: nil}
}
func (repoap *AuthPostgresRepo) updateUserName(ctx context.Context, tx *sql.Tx, userId uuid.UUID, name string) *RepositoryResponse {
	const place = UpdateName
	start := time.Now()
	defer deferMetrics(place, start)
	traceid := ctx.Value("traceID").(string)
	_, err := tx.ExecContext(ctx, fmt.Sprintf("UPDATE %s SET %s = $1 where %s = $2", KeyUserTable, KeyUserName, KeyUserID), name, userId)
	if err != nil {
		repoap.KafkaProducer.NewUserLog(kafka.LogLevelError, place, traceid, err.Error())
		metrics.UserDBErrorsTotal.WithLabelValues(erro.ServerErrorType, "UPDATE").Inc()
		return &RepositoryResponse{Success: false, Errors: &erro.ErrorResponse{Type: erro.ServerErrorType, Message: erro.UserServiceUnavalaible}}
	}
	repoap.KafkaProducer.NewUserLog(kafka.LogLevelInfo, place, traceid, "Successful update username")
	return &RepositoryResponse{Success: true}
}
func (repoap *AuthPostgresRepo) updateUserEmail(ctx context.Context, tx *sql.Tx, userId uuid.UUID, email string, password string) *RepositoryResponse {
	const place = UpdateEmail
	start := time.Now()
	defer deferMetrics(place, start)
	traceid := ctx.Value("traceID").(string)
	var hashpass string
	err := tx.QueryRowContext(ctx, fmt.Sprintf("SELECT %s FROM %s WHERE %s = $1", KeyUserPassword, KeyUserTable, KeyUserID), userId).Scan(&hashpass)
	metrics.UserDBQueriesTotal.WithLabelValues("SELECT").Inc()
	if err != nil {
		repoap.KafkaProducer.NewUserLog(kafka.LogLevelError, place, traceid, err.Error())
		metrics.UserDBErrorsTotal.WithLabelValues(erro.ServerErrorType, "SELECT").Inc()
		return &RepositoryResponse{Success: false, Errors: &erro.ErrorResponse{Type: erro.ServerErrorType, Message: erro.UserServiceUnavalaible}}
	}
	err = bcrypt.CompareHashAndPassword([]byte(hashpass), []byte(password))
	metrics.UserDBQueriesTotal.WithLabelValues("CompareHashAndPassword").Inc()
	if err != nil {
		repoap.KafkaProducer.NewUserLog(kafka.LogLevelWarn, place, traceid, "Incorrect password has been entered")
		metrics.UserDBErrorsTotal.WithLabelValues(erro.ClientErrorType, "CompareHashAndPassword").Inc()
		return &RepositoryResponse{Success: false, Errors: &erro.ErrorResponse{Type: erro.ClientErrorType, Message: erro.ErrorInvalidPasswordConst}}
	}
	var count int
	err = tx.QueryRowContext(ctx, fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE %s = $1", KeyUserTable, KeyUserEmail), email).Scan(&count)
	if err != nil {
		repoap.KafkaProducer.NewUserLog(kafka.LogLevelError, place, traceid, err.Error())
		metrics.UserDBErrorsTotal.WithLabelValues(erro.ServerErrorType, "SELECT").Inc()
		return &RepositoryResponse{Success: false, Errors: &erro.ErrorResponse{Type: erro.ServerErrorType, Message: erro.UserServiceUnavalaible}}
	}
	if count > 0 {
		repoap.KafkaProducer.NewUserLog(kafka.LogLevelWarn, place, traceid, "Already registered email has been entered")
		metrics.UserDBErrorsTotal.WithLabelValues(erro.ClientErrorType, "SELECT").Inc()
		return &RepositoryResponse{Success: false, Errors: &erro.ErrorResponse{Type: erro.ClientErrorType, Message: erro.ErrorUniqueEmailConst}}
	}
	_, err = tx.ExecContext(ctx, fmt.Sprintf("UPDATE %s SET %s = $1 where %s = $2", KeyUserTable, KeyUserEmail, KeyUserID), email, userId)
	metrics.UserDBQueriesTotal.WithLabelValues("UPDATE").Inc()
	if err != nil {
		repoap.KafkaProducer.NewUserLog(kafka.LogLevelError, place, traceid, err.Error())
		metrics.UserDBErrorsTotal.WithLabelValues(erro.ServerErrorType, "UPDATE").Inc()
		return &RepositoryResponse{Success: false, Errors: &erro.ErrorResponse{Type: erro.ServerErrorType, Message: erro.UserServiceUnavalaible}}
	}
	repoap.KafkaProducer.NewUserLog(kafka.LogLevelInfo, place, traceid, "Successful update useremail")
	return &RepositoryResponse{Success: true}
}
func (repoap *AuthPostgresRepo) updateUserPassword(ctx context.Context, tx *sql.Tx, userId uuid.UUID, lastpassword string, newpassword string) *RepositoryResponse {
	const place = UpdatePassword
	start := time.Now()
	defer deferMetrics(place, start)
	traceid := ctx.Value("traceID").(string)
	var hashpass string
	err := tx.QueryRowContext(ctx, fmt.Sprintf("SELECT %s FROM %s WHERE %s = $1", KeyUserPassword, KeyUserTable, KeyUserID), userId).Scan(&hashpass)
	metrics.UserDBQueriesTotal.WithLabelValues("SELECT").Inc()
	if err != nil {
		repoap.KafkaProducer.NewUserLog(kafka.LogLevelError, place, traceid, err.Error())
		metrics.UserDBErrorsTotal.WithLabelValues(erro.ServerErrorType, "SELECT").Inc()
		return &RepositoryResponse{Success: false, Errors: &erro.ErrorResponse{Type: erro.ServerErrorType, Message: erro.UserServiceUnavalaible}}
	}
	err = bcrypt.CompareHashAndPassword([]byte(hashpass), []byte(lastpassword))
	metrics.UserDBQueriesTotal.WithLabelValues("CompareHashAndPassword").Inc()
	if err != nil {
		repoap.KafkaProducer.NewUserLog(kafka.LogLevelWarn, place, traceid, "Incorrect password has been entered")
		metrics.UserDBErrorsTotal.WithLabelValues(erro.ClientErrorType, "CompareHashAndPassword").Inc()
		return &RepositoryResponse{Success: false, Errors: &erro.ErrorResponse{Type: erro.ClientErrorType, Message: erro.ErrorInvalidPasswordConst}}
	}
	hashnewpass, err := bcrypt.GenerateFromPassword([]byte(newpassword), bcrypt.DefaultCost)
	metrics.UserDBQueriesTotal.WithLabelValues("GenerateHashPassword").Inc()
	if err != nil {
		fmterr := fmt.Sprintf("Generate HashPassword Error: %v", err)
		repoap.KafkaProducer.NewUserLog(kafka.LogLevelError, place, traceid, fmterr)
		metrics.UserDBErrorsTotal.WithLabelValues(erro.ServerErrorType, "GenerateHashPassword").Inc()
		return &RepositoryResponse{Success: false, Errors: &erro.ErrorResponse{Type: erro.ServerErrorType, Message: erro.UserServiceUnavalaible}}
	}
	_, err = tx.ExecContext(ctx, fmt.Sprintf("UPDATE %s SET %s = $1 where %s = $2", KeyUserTable, KeyUserPassword, KeyUserID), hashnewpass, userId)
	metrics.UserDBQueriesTotal.WithLabelValues("UPDATE").Inc()
	if err != nil {
		repoap.KafkaProducer.NewUserLog(kafka.LogLevelError, place, traceid, err.Error())
		metrics.UserDBErrorsTotal.WithLabelValues(erro.ServerErrorType, "UPDATE").Inc()
		return &RepositoryResponse{Success: false, Errors: &erro.ErrorResponse{Type: erro.ServerErrorType, Message: erro.UserServiceUnavalaible}}
	}
	repoap.KafkaProducer.NewUserLog(kafka.LogLevelInfo, place, traceid, "Successful update userpassword")
	return &RepositoryResponse{Success: true}
}
