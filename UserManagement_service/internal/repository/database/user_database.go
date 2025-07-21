package database

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/erro"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/metrics"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/model"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

type UserDatabase struct {
	databaseclient *DBObject
}

func NewUserDatabase(db *DBObject) *UserDatabase {
	return &UserDatabase{databaseclient: db}
}
func DBMetrics(place string, start time.Time) {
	metrics.UserDBQueriesTotal.WithLabelValues(place).Inc()
	duration := time.Since(start).Seconds()
	metrics.UserDBQueryDuration.WithLabelValues(place).Observe(duration)
}

const (
	insertUserQuery         = `INSERT INTO users (userid, username, useremail, userpassword) VALUES ($1, $2, $3, $4) ON CONFLICT (useremail) DO NOTHING RETURNING userid`
	selectUserGetQuery      = `SELECT userid, userpassword FROM users WHERE useremail = $1`
	selectUserPasswordQuery = `SELECT userpassword FROM users WHERE userid = $1`
	deleteUserQuery         = `DELETE FROM users WHERE userid = $1`
	selectUserGetProfile    = `SELECT useremail, username FROM users WHERE userid = $1`
	updateUserName          = `UPDATE users SET username = $1 where userid = $2`
	updateUserEmail         = `UPDATE users SET useremail = $1 where userid = $2`
	updateUserPassword      = `UPDATE users SET userpassword = $1 where userid = $2`
	selectEmailCount        = `SELECT COUNT(*) FROM users WHERE useremail = $1`
)

func (repoap *UserDatabase) CreateUser(ctx context.Context, tx pgx.Tx, user *model.User) *repository.RepositoryResponse {
	const place = CreateUser
	start := time.Now()
	defer DBMetrics(place, start)
	var insertedID string
	err := tx.QueryRow(ctx, insertUserQuery, user.Id, user.Name, user.Email, user.Password).Scan(&insertedID)
	metrics.UserDBQueriesTotal.WithLabelValues("INSERT").Inc()
	if err != nil {
		if err == pgx.ErrNoRows {
			metrics.UserDBErrorsTotal.WithLabelValues(erro.ClientErrorType, "INSERT").Inc()
			return repository.BadResponse(erro.ClientError(erro.ErrorUniqueEmailConst), place)
		}
		metrics.UserDBErrorsTotal.WithLabelValues(erro.ServerErrorType, "INSERT").Inc()
		return repository.BadResponse(erro.ServerError(fmt.Sprintf(erro.ErrorAfterReqUsers, err)), place)
	}
	return repository.SuccessResponse(nil, place, "Successful create user in database")
}
func (repoap *UserDatabase) GetUser(ctx context.Context, useremail, userpassword string) *repository.RepositoryResponse {
	const place = GetUser
	start := time.Now()
	defer DBMetrics(place, start)
	var hashpass string
	var userId uuid.UUID
	err := repoap.databaseclient.pool.QueryRow(ctx, selectUserGetQuery, useremail).Scan(&userId, &hashpass)
	metrics.UserDBQueriesTotal.WithLabelValues("SELECT").Inc()
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			metrics.UserDBErrorsTotal.WithLabelValues(erro.ClientErrorType, "SELECT").Inc()
			return repository.BadResponse(erro.ClientError(erro.ErrorEmailNotRegisterConst), place)
		}
		metrics.UserDBErrorsTotal.WithLabelValues(erro.ServerErrorType, "SELECT").Inc()
		return repository.BadResponse(erro.ServerError(fmt.Sprintf(erro.ErrorAfterReqUsers, err)), place)
	}
	err = bcrypt.CompareHashAndPassword([]byte(hashpass), []byte(userpassword))
	metrics.UserDBQueriesTotal.WithLabelValues("CompareHashAndPassword").Inc()
	if err != nil {
		metrics.UserDBErrorsTotal.WithLabelValues(erro.ClientErrorType, "CompareHashAndPassword").Inc()
		return repository.BadResponse(erro.ClientError(erro.ErrorIncorrectPassword), place)
	}
	return repository.SuccessResponse(map[string]any{repository.KeyUserID: userId.String()}, place, "Successful get user from database")
}
func (repoap *UserDatabase) DeleteUser(ctx context.Context, tx pgx.Tx, userId uuid.UUID, password string) *repository.RepositoryResponse {
	const place = DeleteUser
	start := time.Now()
	defer DBMetrics(place, start)
	var hashpass string
	err := tx.QueryRow(ctx, selectUserPasswordQuery, userId).Scan(&hashpass)
	metrics.UserDBQueriesTotal.WithLabelValues("SELECT").Inc()
	if err != nil {
		metrics.UserDBErrorsTotal.WithLabelValues(erro.ServerErrorType, "SELECT").Inc()
		return repository.BadResponse(erro.ServerError(fmt.Sprintf(erro.ErrorAfterReqUsers, err)), place)
	}
	err = bcrypt.CompareHashAndPassword([]byte(hashpass), []byte(password))
	metrics.UserDBQueriesTotal.WithLabelValues("CompareHashAndPassword").Inc()
	if err != nil {
		metrics.UserDBErrorsTotal.WithLabelValues(erro.ClientErrorType, "CompareHashAndPassword").Inc()
		return repository.BadResponse(erro.ClientError(erro.ErrorIncorrectPassword), place)
	}
	_, err = tx.Exec(ctx, deleteUserQuery, userId)
	metrics.UserDBQueriesTotal.WithLabelValues("DELETE").Inc()
	if err != nil {
		metrics.UserDBErrorsTotal.WithLabelValues(erro.ServerErrorType, "DELETE").Inc()
		return repository.BadResponse(erro.ServerError(fmt.Sprintf(erro.ErrorAfterReqUsers, err)), place)
	}
	return repository.SuccessResponse(nil, place, "Successful delete user from database")
}
func (repoap *UserDatabase) GetProfileById(ctx context.Context, userid uuid.UUID) *repository.RepositoryResponse {
	const place = GetProfileById
	start := time.Now()
	defer DBMetrics(place, start)
	var email string
	var name string
	err := repoap.databaseclient.pool.QueryRow(ctx, selectUserGetProfile, userid).Scan(&email, &name)
	metrics.UserDBQueriesTotal.WithLabelValues("SELECT").Inc()
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			metrics.UserDBErrorsTotal.WithLabelValues(erro.ClientErrorType, "SELECT").Inc()
			return repository.BadResponse(erro.ClientError(erro.ErrorIDNotRegisterConst), place)
		}
		metrics.UserDBErrorsTotal.WithLabelValues(erro.ServerErrorType, "SELECT").Inc()
		return repository.BadResponse(erro.ServerError(fmt.Sprintf(erro.ErrorAfterReqUsers, err)), place)
	}
	return repository.SuccessResponse(map[string]any{repository.KeyUser: &model.User{Id: userid, Name: name, Email: email}}, place, "Successful get profile by id from database")
}
func (repoap *UserDatabase) UpdateUserData(ctx context.Context, tx pgx.Tx, userId uuid.UUID, updateType string, args ...interface{}) *repository.RepositoryResponse {
	const place = UpdateUserData
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
	return repository.BadResponse(erro.ClientError(erro.ErrorInvalidDinamicParameter), place)
}
func (repoap *UserDatabase) updateUserName(ctx context.Context, tx pgx.Tx, userId uuid.UUID, name string) *repository.RepositoryResponse {
	const place = UpdateName
	start := time.Now()
	defer DBMetrics(place, start)
	_, err := tx.Exec(ctx, updateUserName, name, userId)
	metrics.UserDBQueriesTotal.WithLabelValues("UPDATE").Inc()
	if err != nil {
		metrics.UserDBErrorsTotal.WithLabelValues(erro.ServerErrorType, "UPDATE").Inc()
		return repository.BadResponse(erro.ServerError(fmt.Sprintf(erro.ErrorAfterReqUsers, err)), place)
	}
	return repository.SuccessResponse(nil, place, "Successful update username in database")
}
func (repoap *UserDatabase) updateUserEmail(ctx context.Context, tx pgx.Tx, userId uuid.UUID, email string, password string) *repository.RepositoryResponse {
	const place = UpdateEmail
	start := time.Now()
	defer DBMetrics(place, start)
	var hashpass string
	err := tx.QueryRow(ctx, selectUserPasswordQuery, userId).Scan(&hashpass)
	metrics.UserDBQueriesTotal.WithLabelValues("SELECT").Inc()
	if err != nil {
		metrics.UserDBErrorsTotal.WithLabelValues(erro.ServerErrorType, "SELECT").Inc()
		return repository.BadResponse(erro.ServerError(fmt.Sprintf(erro.ErrorAfterReqUsers, err)), place)
	}
	err = bcrypt.CompareHashAndPassword([]byte(hashpass), []byte(password))
	metrics.UserDBQueriesTotal.WithLabelValues("CompareHashAndPassword").Inc()
	if err != nil {
		metrics.UserDBErrorsTotal.WithLabelValues(erro.ClientErrorType, "CompareHashAndPassword").Inc()
		return repository.BadResponse(erro.ClientError(erro.ErrorIncorrectPassword), place)
	}
	var count int
	err = tx.QueryRow(ctx, selectEmailCount, email).Scan(&count)
	metrics.UserDBQueriesTotal.WithLabelValues("SELECT").Inc()
	if err != nil {
		metrics.UserDBErrorsTotal.WithLabelValues(erro.ServerErrorType, "SELECT").Inc()
		return repository.BadResponse(erro.ServerError(fmt.Sprintf(erro.ErrorAfterReqUsers, err)), place)
	}
	if count > 0 {
		metrics.UserDBErrorsTotal.WithLabelValues(erro.ClientErrorType, "SELECT").Inc()
		return repository.BadResponse(erro.ClientError(erro.ErrorUniqueEmailConst), place)
	}
	_, err = tx.Exec(ctx, updateUserEmail, email, userId)
	metrics.UserDBQueriesTotal.WithLabelValues("UPDATE").Inc()
	if err != nil {
		metrics.UserDBErrorsTotal.WithLabelValues(erro.ServerErrorType, "UPDATE").Inc()
		return repository.BadResponse(erro.ServerError(fmt.Sprintf(erro.ErrorAfterReqUsers, err)), place)
	}
	return repository.SuccessResponse(nil, place, "Successful update useremail in database")
}
func (repoap *UserDatabase) updateUserPassword(ctx context.Context, tx pgx.Tx, userId uuid.UUID, lastpassword string, newpassword string) *repository.RepositoryResponse {
	const place = UpdatePassword
	start := time.Now()
	defer DBMetrics(place, start)
	var hashpass string
	err := tx.QueryRow(ctx, selectUserPasswordQuery, userId).Scan(&hashpass)
	metrics.UserDBQueriesTotal.WithLabelValues("SELECT").Inc()
	if err != nil {
		metrics.UserDBErrorsTotal.WithLabelValues(erro.ServerErrorType, "SELECT").Inc()
		return repository.BadResponse(erro.ServerError(fmt.Sprintf(erro.ErrorAfterReqUsers, err)), place)
	}
	err = bcrypt.CompareHashAndPassword([]byte(hashpass), []byte(lastpassword))
	metrics.UserDBQueriesTotal.WithLabelValues("CompareHashAndPassword").Inc()
	if err != nil {
		metrics.UserDBErrorsTotal.WithLabelValues(erro.ClientErrorType, "CompareHashAndPassword").Inc()
		return repository.BadResponse(erro.ClientError(erro.ErrorIncorrectPassword), place)
	}
	hashnewpass, err := bcrypt.GenerateFromPassword([]byte(newpassword), bcrypt.DefaultCost)
	metrics.UserDBQueriesTotal.WithLabelValues("GenerateHashPassword").Inc()
	if err != nil {
		metrics.UserDBErrorsTotal.WithLabelValues(erro.ServerErrorType, "GenerateHashPassword").Inc()
		return repository.BadResponse(erro.ServerError(fmt.Sprintf(erro.ErrorGenerateHashPassword, err)), place)
	}
	_, err = tx.Exec(ctx, updateUserPassword, hashnewpass, userId)
	metrics.UserDBQueriesTotal.WithLabelValues("UPDATE").Inc()
	if err != nil {
		metrics.UserDBErrorsTotal.WithLabelValues(erro.ServerErrorType, "UPDATE").Inc()
		return repository.BadResponse(erro.ServerError(fmt.Sprintf(erro.ErrorAfterReqUsers, err)), place)
	}
	return repository.SuccessResponse(nil, place, "Successful update userpassword in database")
}
