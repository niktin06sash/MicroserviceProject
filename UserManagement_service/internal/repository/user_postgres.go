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
	db *DBObject
}

func NewUserPostgresRepo(db *DBObject) *UserPostgresRepo {
	return &UserPostgresRepo{db: db}
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

func (repoap *UserPostgresRepo) CreateUser(ctx context.Context, tx *sql.Tx, user *model.User) *RepositoryResponse {
	const place = CreateUser
	start := time.Now()
	defer DBMetrics(place, start)
	var createdUserID uuid.UUID
	stmt := tx.StmtContext(ctx, repoap.db.mapstmt[insertUserQuery])
	err := stmt.QueryRowContext(ctx, user.Id, user.Name, user.Email, user.Password).Scan(&createdUserID)
	metrics.UserDBQueriesTotal.WithLabelValues("INSERT").Inc()
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			metrics.UserDBErrorsTotal.WithLabelValues(erro.ClientErrorType, "INSERT").Inc()
			return &RepositoryResponse{Success: false, Errors: map[string]string{erro.ErrorType: erro.ClientErrorType, erro.ErrorMessage: erro.ErrorUniqueEmailConst}, Place: place}
		}
		fmterr := fmt.Sprintf("Error after request into %s: %v", KeyUserTable, err)
		metrics.UserDBErrorsTotal.WithLabelValues(erro.ServerErrorType, "INSERT").Inc()
		return &RepositoryResponse{Success: false, Errors: map[string]string{erro.ErrorType: erro.ServerErrorType, erro.ErrorMessage: fmterr}, Place: place}
	}
	return &RepositoryResponse{Success: true, Data: map[string]any{KeyUserID: createdUserID.String()}, SuccessMessage: "Successful create user in database", Place: place}
}
func (repoap *UserPostgresRepo) GetUser(ctx context.Context, useremail, userpassword string) *RepositoryResponse {
	const place = AuthenticateUser
	start := time.Now()
	defer DBMetrics(place, start)
	var hashpass string
	var userId uuid.UUID
	err := repoap.db.mapstmt[selectUserGetQuery].QueryRowContext(ctx, useremail).Scan(&userId, &hashpass)
	metrics.UserDBQueriesTotal.WithLabelValues("SELECT").Inc()
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			metrics.UserDBErrorsTotal.WithLabelValues(erro.ClientErrorType, "SELECT").Inc()
			return &RepositoryResponse{Success: false, Errors: map[string]string{erro.ErrorType: erro.ClientErrorType, erro.ErrorMessage: erro.ErrorEmailNotRegisterConst}, Place: place}
		}
		fmterr := fmt.Sprintf("Error after request into %s: %v", KeyUserTable, err)
		metrics.UserDBErrorsTotal.WithLabelValues(erro.ServerErrorType, "SELECT").Inc()
		return &RepositoryResponse{Success: false, Errors: map[string]string{erro.ErrorType: erro.ServerErrorType, erro.ErrorMessage: fmterr}, Place: place}
	}
	err = bcrypt.CompareHashAndPassword([]byte(hashpass), []byte(userpassword))
	metrics.UserDBQueriesTotal.WithLabelValues("CompareHashAndPassword").Inc()
	if err != nil {
		metrics.UserDBErrorsTotal.WithLabelValues(erro.ClientErrorType, "CompareHashAndPassword").Inc()
		return &RepositoryResponse{Success: false, Errors: map[string]string{erro.ErrorType: erro.ClientErrorType, erro.ErrorMessage: erro.ErrorIncorrectPassword}, Place: place}
	}
	return &RepositoryResponse{Success: true, Data: map[string]any{KeyUserID: userId.String()}, SuccessMessage: "Successful get user from database", Place: place}
}
func (repoap *UserPostgresRepo) DeleteUser(ctx context.Context, tx *sql.Tx, userId uuid.UUID, password string) *RepositoryResponse {
	const place = DeleteUser
	start := time.Now()
	defer DBMetrics(place, start)
	var hashpass string
	stmt := tx.StmtContext(ctx, repoap.db.mapstmt[selectUserPasswordQuery])
	err := stmt.QueryRowContext(ctx, userId).Scan(&hashpass)
	metrics.UserDBQueriesTotal.WithLabelValues("SELECT").Inc()
	if err != nil {
		fmterr := fmt.Sprintf("Error after request into %s: %v", KeyUserTable, err)
		metrics.UserDBErrorsTotal.WithLabelValues(erro.ServerErrorType, "SELECT").Inc()
		return &RepositoryResponse{Success: false, Errors: map[string]string{erro.ErrorType: erro.ServerErrorType, erro.ErrorMessage: fmterr}, Place: place}
	}
	err = bcrypt.CompareHashAndPassword([]byte(hashpass), []byte(password))
	metrics.UserDBQueriesTotal.WithLabelValues("CompareHashAndPassword").Inc()
	if err != nil {
		metrics.UserDBErrorsTotal.WithLabelValues(erro.ClientErrorType, "CompareHashAndPassword").Inc()
		return &RepositoryResponse{Success: false, Errors: map[string]string{erro.ErrorType: erro.ClientErrorType, erro.ErrorMessage: erro.ErrorIncorrectPassword}, Place: place}
	}
	stmt = tx.StmtContext(ctx, repoap.db.mapstmt[deleteUserQuery])
	_, err = stmt.ExecContext(ctx, userId)
	metrics.UserDBQueriesTotal.WithLabelValues("DELETE").Inc()
	if err != nil {
		fmterr := fmt.Sprintf("Error after request into %s: %v", KeyUserTable, err)
		metrics.UserDBErrorsTotal.WithLabelValues(erro.ServerErrorType, "DELETE").Inc()
		return &RepositoryResponse{Success: false, Errors: map[string]string{erro.ErrorType: erro.ServerErrorType, erro.ErrorMessage: fmterr}, Place: place}
	}
	return &RepositoryResponse{Success: true, SuccessMessage: "Successful delete user from database", Place: place}
}
func (repoap *UserPostgresRepo) GetProfileById(ctx context.Context, userid uuid.UUID) *RepositoryResponse {
	const place = GetProfileById
	start := time.Now()
	defer DBMetrics(place, start)
	var email string
	var name string
	err := repoap.db.mapstmt[selectUserGetProfile].QueryRowContext(ctx, userid).Scan(&email, &name)
	metrics.UserDBQueriesTotal.WithLabelValues("SELECT").Inc()
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			metrics.UserDBErrorsTotal.WithLabelValues(erro.ClientErrorType, "SELECT").Inc()
			return &RepositoryResponse{Success: false, Errors: map[string]string{erro.ErrorType: erro.ClientErrorType, erro.ErrorMessage: erro.ErrorIDNotRegisterConst}, Place: place}
		}
		fmterr := fmt.Sprintf("Error after request into %s: %v", KeyUserTable, err)
		metrics.UserDBErrorsTotal.WithLabelValues(erro.ServerErrorType, "SELECT").Inc()
		return &RepositoryResponse{Success: false, Errors: map[string]string{erro.ErrorType: erro.ServerErrorType, erro.ErrorMessage: fmterr}, Place: place}
	}
	return &RepositoryResponse{Success: true, Data: map[string]any{KeyUserID: userid.String(), KeyUserEmail: email, KeyUserName: name}, SuccessMessage: "Successful get profile by id from database", Place: place}
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
	return &RepositoryResponse{Success: false, Errors: map[string]string{erro.ErrorType: erro.ClientErrorType, erro.ErrorMessage: erro.ErrorInvalidDinamicParameter}, Place: place}
}
func (repoap *UserPostgresRepo) updateUserName(ctx context.Context, tx *sql.Tx, userId uuid.UUID, name string) *RepositoryResponse {
	const place = UpdateName
	start := time.Now()
	defer DBMetrics(place, start)
	stmt := tx.StmtContext(ctx, repoap.db.mapstmt[updateUserName])
	_, err := stmt.ExecContext(ctx, name, userId)
	if err != nil {
		fmterr := fmt.Sprintf("Error after request into %s: %v", KeyUserTable, err)
		metrics.UserDBErrorsTotal.WithLabelValues(erro.ServerErrorType, "UPDATE").Inc()
		return &RepositoryResponse{Success: false, Errors: map[string]string{erro.ErrorType: erro.ServerErrorType, erro.ErrorMessage: fmterr}, Place: place}
	}
	return &RepositoryResponse{Success: true, SuccessMessage: "Successful update username in database", Place: place}
}
func (repoap *UserPostgresRepo) updateUserEmail(ctx context.Context, tx *sql.Tx, userId uuid.UUID, email string, password string) *RepositoryResponse {
	const place = UpdateEmail
	start := time.Now()
	defer DBMetrics(place, start)
	var hashpass string
	stmt := tx.StmtContext(ctx, repoap.db.mapstmt[selectUserPasswordQuery])
	err := stmt.QueryRowContext(ctx, userId).Scan(&hashpass)
	metrics.UserDBQueriesTotal.WithLabelValues("SELECT").Inc()
	if err != nil {
		fmterr := fmt.Sprintf("Error after request into %s: %v", KeyUserTable, err)
		metrics.UserDBErrorsTotal.WithLabelValues(erro.ServerErrorType, "SELECT").Inc()
		return &RepositoryResponse{Success: false, Errors: map[string]string{erro.ErrorType: erro.ServerErrorType, erro.ErrorMessage: fmterr}, Place: place}
	}
	err = bcrypt.CompareHashAndPassword([]byte(hashpass), []byte(password))
	metrics.UserDBQueriesTotal.WithLabelValues("CompareHashAndPassword").Inc()
	if err != nil {
		metrics.UserDBErrorsTotal.WithLabelValues(erro.ClientErrorType, "CompareHashAndPassword").Inc()
		return &RepositoryResponse{Success: false, Errors: map[string]string{erro.ErrorType: erro.ClientErrorType, erro.ErrorMessage: erro.ErrorIncorrectPassword}, Place: place}
	}
	var count int
	stmt = tx.StmtContext(ctx, repoap.db.mapstmt[selectEmailCount])
	err = stmt.QueryRowContext(ctx, email).Scan(&count)
	if err != nil {
		fmterr := fmt.Sprintf("Error after request into %s: %v", KeyUserTable, err)
		metrics.UserDBErrorsTotal.WithLabelValues(erro.ServerErrorType, "SELECT").Inc()
		return &RepositoryResponse{Success: false, Errors: map[string]string{erro.ErrorType: erro.ServerErrorType, erro.ErrorMessage: fmterr}, Place: place}
	}
	if count > 0 {
		metrics.UserDBErrorsTotal.WithLabelValues(erro.ClientErrorType, "SELECT").Inc()
		return &RepositoryResponse{Success: false, Errors: map[string]string{erro.ErrorType: erro.ClientErrorType, erro.ErrorMessage: erro.ErrorUniqueEmailConst}, Place: place}
	}
	stmt = tx.StmtContext(ctx, repoap.db.mapstmt[updateUserEmail])
	_, err = stmt.ExecContext(ctx, email, userId)
	metrics.UserDBQueriesTotal.WithLabelValues("UPDATE").Inc()
	if err != nil {
		fmterr := fmt.Sprintf("Error after request into %s: %v", KeyUserTable, err)
		metrics.UserDBErrorsTotal.WithLabelValues(erro.ServerErrorType, "UPDATE").Inc()
		return &RepositoryResponse{Success: false, Errors: map[string]string{erro.ErrorType: erro.ServerErrorType, erro.ErrorMessage: fmterr}, Place: place}
	}
	return &RepositoryResponse{Success: true, SuccessMessage: "Successful update useremail in database", Place: place}
}
func (repoap *UserPostgresRepo) updateUserPassword(ctx context.Context, tx *sql.Tx, userId uuid.UUID, lastpassword string, newpassword string) *RepositoryResponse {
	const place = UpdatePassword
	start := time.Now()
	defer DBMetrics(place, start)
	var hashpass string
	stmt := tx.StmtContext(ctx, repoap.db.mapstmt[selectUserPasswordQuery])
	err := stmt.QueryRowContext(ctx, userId).Scan(&hashpass)
	metrics.UserDBQueriesTotal.WithLabelValues("SELECT").Inc()
	if err != nil {
		fmterr := fmt.Sprintf("Error after request into %s: %v", KeyUserTable, err)
		metrics.UserDBErrorsTotal.WithLabelValues(erro.ServerErrorType, "SELECT").Inc()
		return &RepositoryResponse{Success: false, Errors: map[string]string{erro.ErrorType: erro.ServerErrorType, erro.ErrorMessage: fmterr}, Place: place}
	}
	err = bcrypt.CompareHashAndPassword([]byte(hashpass), []byte(lastpassword))
	metrics.UserDBQueriesTotal.WithLabelValues("CompareHashAndPassword").Inc()
	if err != nil {
		metrics.UserDBErrorsTotal.WithLabelValues(erro.ClientErrorType, "CompareHashAndPassword").Inc()
		return &RepositoryResponse{Success: false, Errors: map[string]string{erro.ErrorType: erro.ClientErrorType, erro.ErrorMessage: erro.ErrorIncorrectPassword}, Place: place}
	}
	hashnewpass, err := bcrypt.GenerateFromPassword([]byte(newpassword), bcrypt.DefaultCost)
	metrics.UserDBQueriesTotal.WithLabelValues("GenerateHashPassword").Inc()
	if err != nil {
		metrics.UserDBErrorsTotal.WithLabelValues(erro.ServerErrorType, "GenerateHashPassword").Inc()
		return &RepositoryResponse{Success: false, Errors: map[string]string{erro.ErrorType: erro.ServerErrorType, erro.ErrorMessage: fmt.Sprintf("Generate HashPassword Error: %v", err)}, Place: place}
	}
	stmt = tx.StmtContext(ctx, repoap.db.mapstmt[UpdatePassword])
	_, err = stmt.ExecContext(ctx, hashnewpass, userId)
	metrics.UserDBQueriesTotal.WithLabelValues("UPDATE").Inc()
	if err != nil {
		fmterr := fmt.Sprintf("Error after request into %s: %v", KeyUserTable, err)
		metrics.UserDBErrorsTotal.WithLabelValues(erro.ServerErrorType, "UPDATE").Inc()
		return &RepositoryResponse{Success: false, Errors: map[string]string{erro.ErrorType: erro.ServerErrorType, erro.ErrorMessage: fmterr}, Place: place}
	}
	return &RepositoryResponse{Success: true, SuccessMessage: "Successful update userpassword in database", Place: place}
}
