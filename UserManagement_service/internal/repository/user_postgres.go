package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/erro"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/metrics"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/model"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UserPostgresRepo struct {
	Db *DBObject
}

func NewUserPostgresRepo(db *DBObject) *UserPostgresRepo {
	return &UserPostgresRepo{Db: db}
}
func DBMetrics(place string, start time.Time) {
	metrics.UserDBQueriesTotal.WithLabelValues(place).Inc()
	duration := time.Since(start).Seconds()
	metrics.UserDBQueryDuration.WithLabelValues(place).Observe(duration)
}
func (repoap *UserPostgresRepo) CreateUser(ctx context.Context, tx *sql.Tx, user *model.User) *RepositoryResponse {
	const place = CreateUser
	start := time.Now()
	defer DBMetrics(place, start)
	var createdUserID uuid.UUID
	err := tx.QueryRowContext(ctx,
		fmt.Sprintf("INSERT INTO %s (%s, %s, %s, %s) VALUES ($1, $2, $3, $4) ON CONFLICT (%s) DO NOTHING RETURNING %s;",
			KeyUserTable, KeyUserID, KeyUserName, KeyUserEmail, KeyUserPassword, KeyUserEmail, KeyUserID),
		user.Id, user.Name, user.Email, user.Password).Scan(&createdUserID)
	metrics.UserDBQueriesTotal.WithLabelValues("INSERT").Inc()
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			metrics.UserDBErrorsTotal.WithLabelValues(erro.ClientErrorType, "INSERT").Inc()
			return &RepositoryResponse{Success: false, Errors: &ErrorResponse{Type: erro.ClientErrorType, Message: "Already registered email has been entered"}, Place: place}
		}
		fmterr := fmt.Sprintf("Error after request into %s: %v", KeyUserTable, err)
		metrics.UserDBErrorsTotal.WithLabelValues(erro.ServerErrorType, "INSERT").Inc()
		return &RepositoryResponse{Success: false, Errors: &ErrorResponse{Type: erro.ServerErrorType, Message: fmterr}, Place: place}
	}
	return &RepositoryResponse{Success: true, Data: map[string]any{KeyUserID: createdUserID.String()}, SuccessMessage: "Successful create user", Place: place}
}
func (repoap *UserPostgresRepo) AuthenticateUser(ctx context.Context, useremail, userpassword string) *RepositoryResponse {
	const place = AuthenticateUser
	start := time.Now()
	defer DBMetrics(place, start)
	var hashpass string
	var userId uuid.UUID
	err := repoap.Db.DB.QueryRowContext(ctx, fmt.Sprintf("SELECT %s, %s FROM %s WHERE %s = $1", KeyUserID, KeyUserPassword, KeyUserTable, KeyUserEmail), useremail).Scan(&userId, &hashpass)
	metrics.UserDBQueriesTotal.WithLabelValues("SELECT").Inc()
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			metrics.UserDBErrorsTotal.WithLabelValues(erro.ClientErrorType, "SELECT").Inc()
			return &RepositoryResponse{Success: false, Errors: &ErrorResponse{Type: erro.ClientErrorType, Message: "Unregistered email has been entered"}, Place: place}
		}
		fmterr := fmt.Sprintf("Error after request into %s: %v", KeyUserTable, err)
		metrics.UserDBErrorsTotal.WithLabelValues(erro.ServerErrorType, "SELECT").Inc()
		return &RepositoryResponse{Success: false, Errors: &ErrorResponse{Type: erro.ServerErrorType, Message: fmterr}, Place: place}
	}
	err = bcrypt.CompareHashAndPassword([]byte(hashpass), []byte(userpassword))
	metrics.UserDBQueriesTotal.WithLabelValues("CompareHashAndPassword").Inc()
	if err != nil {
		metrics.UserDBErrorsTotal.WithLabelValues(erro.ClientErrorType, "CompareHashAndPassword").Inc()
		return &RepositoryResponse{Success: false, Errors: &ErrorResponse{Type: erro.ClientErrorType, Message: "Incorrect password has been entered"}, Place: place}
	}
	return &RepositoryResponse{Success: true, Data: map[string]any{KeyUserID: userId.String()}, SuccessMessage: "Successful authenticate user", Place: place}
}
func (repoap *UserPostgresRepo) DeleteUser(ctx context.Context, tx *sql.Tx, userId uuid.UUID, password string) *RepositoryResponse {
	const place = DeleteUser
	start := time.Now()
	defer DBMetrics(place, start)
	var hashpass string
	err := tx.QueryRowContext(ctx, fmt.Sprintf("SELECT %s FROM %s WHERE %s = $1", KeyUserPassword, KeyUserTable, KeyUserID), userId).Scan(&hashpass)
	metrics.UserDBQueriesTotal.WithLabelValues("SELECT").Inc()
	if err != nil {
		fmterr := fmt.Sprintf("Error after request into %s: %v", KeyUserTable, err)
		metrics.UserDBErrorsTotal.WithLabelValues(erro.ServerErrorType, "SELECT").Inc()
		return &RepositoryResponse{Success: false, Errors: &ErrorResponse{Type: erro.ServerErrorType, Message: fmterr}, Place: place}
	}
	err = bcrypt.CompareHashAndPassword([]byte(hashpass), []byte(password))
	metrics.UserDBQueriesTotal.WithLabelValues("CompareHashAndPassword").Inc()
	if err != nil {
		metrics.UserDBErrorsTotal.WithLabelValues(erro.ClientErrorType, "CompareHashAndPassword").Inc()
		return &RepositoryResponse{Success: false, Errors: &ErrorResponse{Type: erro.ClientErrorType, Message: "Incorrect password has been entered"}, Place: place}
	}
	_, err = tx.ExecContext(ctx, fmt.Sprintf("DELETE FROM %s where %s = $1", KeyUserTable, KeyUserID), userId)
	metrics.UserDBQueriesTotal.WithLabelValues("DELETE").Inc()
	if err != nil {
		fmterr := fmt.Sprintf("Error after request into %s: %v", KeyUserTable, err)
		metrics.UserDBErrorsTotal.WithLabelValues(erro.ServerErrorType, "DELETE").Inc()
		return &RepositoryResponse{Success: false, Errors: &ErrorResponse{Type: erro.ServerErrorType, Message: fmterr}, Place: place}
	}
	return &RepositoryResponse{Success: true, SuccessMessage: "Successful delete user", Place: place}
}
func (repoap *UserPostgresRepo) GetMyProfile(ctx context.Context, userid uuid.UUID) *RepositoryResponse {
	const place = GetMyProfile
	start := time.Now()
	defer DBMetrics(place, start)
	var email string
	var name string
	err := repoap.Db.DB.QueryRowContext(ctx, fmt.Sprintf("SELECT %s, %s FROM %s WHERE %s = $1", KeyUserEmail, KeyUserName, KeyUserTable, KeyUserID), userid).Scan(&email, &name)
	metrics.UserDBQueriesTotal.WithLabelValues("SELECT").Inc()
	if err != nil {
		fmterr := fmt.Sprintf("Error after request into %s: %v", KeyUserTable, err)
		metrics.UserDBErrorsTotal.WithLabelValues(erro.ServerErrorType, "SELECT").Inc()
		return &RepositoryResponse{Success: false, Errors: &ErrorResponse{Type: erro.ServerErrorType, Message: fmterr}, Place: place}
	}
	return &RepositoryResponse{Success: true, Data: map[string]any{KeyUserID: userid.String(), KeyUserEmail: email, KeyUserName: name}, SuccessMessage: "Successful get my profile from db", Place: place}
}
func (repoap *UserPostgresRepo) GetProfileById(ctx context.Context, getid uuid.UUID) *RepositoryResponse {
	const place = GetProfileById
	start := time.Now()
	defer DBMetrics(place, start)
	var email string
	var name string
	err := repoap.Db.DB.QueryRowContext(ctx, fmt.Sprintf("SELECT %s, %s FROM %s WHERE %s = $1", KeyUserEmail, KeyUserName, KeyUserTable, KeyUserID), getid).Scan(&email, &name)
	metrics.UserDBQueriesTotal.WithLabelValues("SELECT").Inc()
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			metrics.UserDBErrorsTotal.WithLabelValues(erro.ClientErrorType, "SELECT").Inc()
			return &RepositoryResponse{Success: false, Errors: &ErrorResponse{Type: erro.ClientErrorType, Message: "Unregistered id has been entered"}, Place: place}
		}
		fmterr := fmt.Sprintf("Error after request into %s: %v", KeyUserTable, err)
		metrics.UserDBErrorsTotal.WithLabelValues(erro.ServerErrorType, "SELECT").Inc()
		return &RepositoryResponse{Success: false, Errors: &ErrorResponse{Type: erro.ServerErrorType, Message: fmterr}, Place: place}
	}
	return &RepositoryResponse{Success: true, Data: map[string]any{KeyUserID: getid.String(), KeyUserEmail: email, KeyUserName: name}, SuccessMessage: "Successful get profile by id from db", Place: place}
}
func (repoap *UserPostgresRepo) UpdateUserData(ctx context.Context, tx *sql.Tx, userId uuid.UUID, updateType string, args ...interface{}) *RepositoryResponse {
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
	return &RepositoryResponse{Success: false, Errors: &ErrorResponse{Message: erro.ErrorInvalidDinamicParameter, Type: erro.ClientErrorType}, Place: place}
}
func (repoap *UserPostgresRepo) updateUserName(ctx context.Context, tx *sql.Tx, userId uuid.UUID, name string) *RepositoryResponse {
	const place = UpdateName
	start := time.Now()
	defer DBMetrics(place, start)
	_, err := tx.ExecContext(ctx, fmt.Sprintf("UPDATE %s SET %s = $1 where %s = $2", KeyUserTable, KeyUserName, KeyUserID), name, userId)
	if err != nil {
		fmterr := fmt.Sprintf("Error after request into %s: %v", KeyUserTable, err)
		metrics.UserDBErrorsTotal.WithLabelValues(erro.ServerErrorType, "UPDATE").Inc()
		return &RepositoryResponse{Success: false, Errors: &ErrorResponse{Type: erro.ServerErrorType, Message: fmterr}, Place: place}
	}
	return &RepositoryResponse{Success: true, SuccessMessage: "Successful update username", Place: place}
}
func (repoap *UserPostgresRepo) updateUserEmail(ctx context.Context, tx *sql.Tx, userId uuid.UUID, email string, password string) *RepositoryResponse {
	const place = UpdateEmail
	start := time.Now()
	defer DBMetrics(place, start)
	var hashpass string
	err := tx.QueryRowContext(ctx, fmt.Sprintf("SELECT %s FROM %s WHERE %s = $1", KeyUserPassword, KeyUserTable, KeyUserID), userId).Scan(&hashpass)
	metrics.UserDBQueriesTotal.WithLabelValues("SELECT").Inc()
	if err != nil {
		fmterr := fmt.Sprintf("Error after request into %s: %v", KeyUserTable, err)
		metrics.UserDBErrorsTotal.WithLabelValues(erro.ServerErrorType, "SELECT").Inc()
		return &RepositoryResponse{Success: false, Errors: &ErrorResponse{Type: erro.ServerErrorType, Message: fmterr}, Place: place}
	}
	err = bcrypt.CompareHashAndPassword([]byte(hashpass), []byte(password))
	metrics.UserDBQueriesTotal.WithLabelValues("CompareHashAndPassword").Inc()
	if err != nil {
		metrics.UserDBErrorsTotal.WithLabelValues(erro.ClientErrorType, "CompareHashAndPassword").Inc()
		return &RepositoryResponse{Success: false, Errors: &ErrorResponse{Type: erro.ClientErrorType, Message: "Incorrect password has been entered"}, Place: place}
	}
	var count int
	err = tx.QueryRowContext(ctx, fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE %s = $1", KeyUserTable, KeyUserEmail), email).Scan(&count)
	if err != nil {
		fmterr := fmt.Sprintf("Error after request into %s: %v", KeyUserTable, err)
		metrics.UserDBErrorsTotal.WithLabelValues(erro.ServerErrorType, "SELECT").Inc()
		return &RepositoryResponse{Success: false, Errors: &ErrorResponse{Type: erro.ServerErrorType, Message: fmterr}, Place: place}
	}
	if count > 0 {
		metrics.UserDBErrorsTotal.WithLabelValues(erro.ClientErrorType, "SELECT").Inc()
		return &RepositoryResponse{Success: false, Errors: &ErrorResponse{Type: erro.ClientErrorType, Message: "Already registered email has been entered"}, Place: place}
	}
	_, err = tx.ExecContext(ctx, fmt.Sprintf("UPDATE %s SET %s = $1 where %s = $2", KeyUserTable, KeyUserEmail, KeyUserID), email, userId)
	metrics.UserDBQueriesTotal.WithLabelValues("UPDATE").Inc()
	if err != nil {
		fmterr := fmt.Sprintf("Error after request into %s: %v", KeyUserTable, err)
		metrics.UserDBErrorsTotal.WithLabelValues(erro.ServerErrorType, "UPDATE").Inc()
		return &RepositoryResponse{Success: false, Errors: &ErrorResponse{Type: erro.ServerErrorType, Message: fmterr}, Place: place}
	}
	return &RepositoryResponse{Success: true, SuccessMessage: "Successful update useremail", Place: place}
}
func (repoap *UserPostgresRepo) updateUserPassword(ctx context.Context, tx *sql.Tx, userId uuid.UUID, lastpassword string, newpassword string) *RepositoryResponse {
	const place = UpdatePassword
	start := time.Now()
	defer DBMetrics(place, start)
	var hashpass string
	err := tx.QueryRowContext(ctx, fmt.Sprintf("SELECT %s FROM %s WHERE %s = $1", KeyUserPassword, KeyUserTable, KeyUserID), userId).Scan(&hashpass)
	metrics.UserDBQueriesTotal.WithLabelValues("SELECT").Inc()
	if err != nil {
		fmterr := fmt.Sprintf("Error after request into %s: %v", KeyUserTable, err)
		metrics.UserDBErrorsTotal.WithLabelValues(erro.ServerErrorType, "SELECT").Inc()
		return &RepositoryResponse{Success: false, Errors: &ErrorResponse{Type: erro.ServerErrorType, Message: fmterr}, Place: place}
	}
	err = bcrypt.CompareHashAndPassword([]byte(hashpass), []byte(lastpassword))
	metrics.UserDBQueriesTotal.WithLabelValues("CompareHashAndPassword").Inc()
	if err != nil {
		metrics.UserDBErrorsTotal.WithLabelValues(erro.ClientErrorType, "CompareHashAndPassword").Inc()
		return &RepositoryResponse{Success: false, Errors: &ErrorResponse{Type: erro.ClientErrorType, Message: "Incorrect password has been entered"}, Place: place}
	}
	hashnewpass, err := bcrypt.GenerateFromPassword([]byte(newpassword), bcrypt.DefaultCost)
	metrics.UserDBQueriesTotal.WithLabelValues("GenerateHashPassword").Inc()
	if err != nil {
		metrics.UserDBErrorsTotal.WithLabelValues(erro.ServerErrorType, "GenerateHashPassword").Inc()
		return &RepositoryResponse{Success: false, Errors: &ErrorResponse{Type: erro.ServerErrorType, Message: fmt.Sprintf("Generate HashPassword Error: %v", err)}, Place: place}
	}
	_, err = tx.ExecContext(ctx, fmt.Sprintf("UPDATE %s SET %s = $1 where %s = $2", KeyUserTable, KeyUserPassword, KeyUserID), hashnewpass, userId)
	metrics.UserDBQueriesTotal.WithLabelValues("UPDATE").Inc()
	if err != nil {
		fmterr := fmt.Sprintf("Error after request into %s: %v", KeyUserTable, err)
		metrics.UserDBErrorsTotal.WithLabelValues(erro.ServerErrorType, "UPDATE").Inc()
		return &RepositoryResponse{Success: false, Errors: &ErrorResponse{Type: erro.ServerErrorType, Message: fmterr}, Place: place}
	}
	return &RepositoryResponse{Success: true, SuccessMessage: "Successful update userpassword", Place: place}
}
