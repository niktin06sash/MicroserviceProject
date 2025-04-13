package repository

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/erro"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/model"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

func TestCreateUser(t *testing.T) {
	var fundamentUuid = uuid.New()
	tests := []struct {
		name            string
		user            *model.Person
		mockSetup       func(mock sqlmock.Sqlmock)
		expectedSuccess bool
		expectedUserId  uuid.UUID
		expectedErrors  error
	}{
		{
			name: "Success",
			user: &model.Person{
				Id:       uuid.New(),
				Name:     "testuser",
				Email:    "test@example.com",
				Password: "password123",
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("INSERT INTO UserZ").
					WithArgs(sqlmock.AnyArg(), "testuser", "test@example.com", "password123").
					WillReturnRows(sqlmock.NewRows([]string{"userid"}).AddRow(fundamentUuid))
			},
			expectedSuccess: true,
			expectedUserId:  fundamentUuid,
			expectedErrors:  nil,
		},
		{
			name: "Email Conflict",
			user: &model.Person{
				Id:       uuid.New(),
				Name:     "testuser",
				Email:    "test@example.com",
				Password: "password123",
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("INSERT INTO UserZ").
					WithArgs(sqlmock.AnyArg(), "testuser", "test@example.com", "password123").
					WillReturnError(sql.ErrNoRows)
			},
			expectedSuccess: false,
			expectedUserId:  uuid.Nil,
			expectedErrors:  erro.ErrorUniqueEmail,
		},
		{
			name: "DB Error",
			user: &model.Person{
				Id:       uuid.New(),
				Name:     "testuser",
				Email:    "test@example.com",
				Password: "password123",
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("INSERT INTO UserZ").
					WithArgs(sqlmock.AnyArg(), "testuser", "test@example.com", "password123").
					WillReturnError(errors.New("database connection error"))
			},
			expectedSuccess: false,
			expectedUserId:  uuid.Nil,
			expectedErrors:  erro.ErrorDbRepositoryError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				log.Fatalf("failed to open sqlmock database: %v", err)
			}
			defer db.Close()

			mock.ExpectBegin()
			tx, err := db.Begin()
			if err != nil {
				t.Fatalf("failed to begin transaction: %v", err)
			}

			if tt.mockSetup != nil {
				tt.mockSetup(mock)
			}

			repo := NewAuthPostgresRepo(db)
			response := repo.CreateUser(context.Background(), tx, tt.user)

			assert.Equal(t, tt.expectedSuccess, response.Success)
			assert.Equal(t, tt.expectedUserId, response.UserId)
			assert.Equal(t, tt.expectedErrors, response.Errors)

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}
func TestGetUser(t *testing.T) {
	var fundamentUuid = uuid.New()

	tests := []struct {
		name            string
		useremail       string
		userpassword    string
		mockSetup       func(mock sqlmock.Sqlmock)
		expectedSuccess bool
		expectedUserId  uuid.UUID
		expectedErrors  error
	}{
		{
			name:         "Success",
			useremail:    "test@example.com",
			userpassword: "password123",
			mockSetup: func(mock sqlmock.Sqlmock) {

				hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
				if err != nil {
					t.Fatalf("failed to hash password: %v", err)
				}

				mock.ExpectQuery(
					`(?i)SELECT userid, userpassword FROM userZ WHERE useremail = \$1`,
				).
					WithArgs("test@example.com").
					WillReturnRows(
						sqlmock.NewRows([]string{"userid", "userpassword"}).
							AddRow(fundamentUuid, string(hashedPassword)),
					)
			},
			expectedSuccess: true,
			expectedUserId:  fundamentUuid,
			expectedErrors:  nil,
		},
		{
			name:         "Email Not Registered",
			useremail:    "nonexistent@example.com",
			userpassword: "password123",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(
					`(?i)SELECT userid, userpassword FROM userZ WHERE useremail = \$1`,
				).
					WithArgs("nonexistent@example.com").
					WillReturnError(sql.ErrNoRows)
			},
			expectedSuccess: false,
			expectedUserId:  uuid.Nil,
			expectedErrors:  erro.ErrorEmailNotRegister,
		},
		{
			name:         "Invalid Password",
			useremail:    "test@example.com",
			userpassword: "wrongpassword",
			mockSetup: func(mock sqlmock.Sqlmock) {

				hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
				if err != nil {
					t.Fatalf("failed to hash password: %v", err)
				}

				mock.ExpectQuery(
					`(?i)SELECT userid, userpassword FROM userZ WHERE useremail = \$1`,
				).
					WithArgs("test@example.com").
					WillReturnRows(
						sqlmock.NewRows([]string{"userid", "userpassword"}).
							AddRow(fundamentUuid, string(hashedPassword)),
					)
			},
			expectedSuccess: false,
			expectedUserId:  uuid.Nil,
			expectedErrors:  erro.ErrorInvalidPassword,
		},
		{
			name:         "Database Error",
			useremail:    "test@example.com",
			userpassword: "password123",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(
					`(?i)SELECT userid, userpassword FROM userZ WHERE useremail = \$1`,
				).
					WithArgs("test@example.com").
					WillReturnError(erro.ErrorDbRepositoryError)
			},
			expectedSuccess: false,
			expectedUserId:  uuid.Nil,
			expectedErrors:  erro.ErrorDbRepositoryError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("failed to open sqlmock database: %v", err)
			}
			defer db.Close()

			if tt.mockSetup != nil {
				tt.mockSetup(mock)
			}

			repo := NewAuthPostgresRepo(db)
			response := repo.GetUser(context.Background(), tt.useremail, tt.userpassword)

			assert.Equal(t, tt.expectedSuccess, response.Success)
			assert.Equal(t, tt.expectedUserId, response.UserId)
			assert.Equal(t, tt.expectedErrors, response.Errors)

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}
func TestDeleteUser(t *testing.T) {
	var fundamentUuid = uuid.New()

	tests := []struct {
		name            string
		userid          uuid.UUID
		userpassword    string
		mockSetup       func(mock sqlmock.Sqlmock)
		expectedSuccess bool
		expectedErrors  error
	}{
		{
			name:         "Success",
			userid:       fundamentUuid,
			userpassword: "password123",
			mockSetup: func(mock sqlmock.Sqlmock) {
				hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
				if err != nil {
					t.Fatalf("failed to hash password: %v", err)
				}

				mock.ExpectQuery(
					`(?i)SELECT userpassword FROM userZ WHERE userid = \$1`,
				).
					WithArgs(fundamentUuid).
					WillReturnRows(
						sqlmock.NewRows([]string{"userpassword"}).
							AddRow(string(hashedPassword)),
					)

				mock.ExpectExec(
					`(?i)DELETE FROM userZ WHERE userId = \$1`,
				).
					WithArgs(fundamentUuid).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			expectedSuccess: true,
			expectedErrors:  nil,
		},
		{
			name:         "User Not Found",
			userid:       fundamentUuid,
			userpassword: "password123",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(
					`(?i)SELECT userpassword FROM userZ WHERE userid = \$1`,
				).
					WithArgs(fundamentUuid).
					WillReturnError(sql.ErrNoRows)
			},
			expectedSuccess: false,
			expectedErrors:  erro.ErrorFoundUser,
		},
		{
			name:         "Invalid Password",
			userid:       fundamentUuid,
			userpassword: "wrongpassword",
			mockSetup: func(mock sqlmock.Sqlmock) {
				hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
				if err != nil {
					t.Fatalf("failed to hash password: %v", err)
				}

				mock.ExpectQuery(
					`(?i)SELECT userpassword FROM userZ WHERE userid = \$1`,
				).
					WithArgs(fundamentUuid).
					WillReturnRows(
						sqlmock.NewRows([]string{"userpassword"}).
							AddRow(string(hashedPassword)),
					)
			},
			expectedSuccess: false,
			expectedErrors:  erro.ErrorInvalidPassword,
		},
		{
			name:         "Database Error(DELETE)",
			userid:       fundamentUuid,
			userpassword: "password123",
			mockSetup: func(mock sqlmock.Sqlmock) {
				hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
				if err != nil {
					t.Fatalf("failed to hash password: %v", err)
				}

				mock.ExpectQuery(
					`(?i)SELECT userpassword FROM userZ WHERE userid = \$1`,
				).
					WithArgs(fundamentUuid).
					WillReturnRows(
						sqlmock.NewRows([]string{"userpassword"}).
							AddRow(string(hashedPassword)),
					)

				mock.ExpectExec(
					`(?i)DELETE FROM userZ WHERE userId = \$1`,
				).
					WithArgs(fundamentUuid).
					WillReturnError(erro.ErrorDbRepositoryError)
			},
			expectedSuccess: false,
			expectedErrors:  erro.ErrorDbRepositoryError,
		},
		{
			name:         "Database Error(SELECT)",
			userid:       fundamentUuid,
			userpassword: "password123",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(
					`(?i)SELECT userpassword FROM userZ WHERE userid = \$1`,
				).
					WithArgs(fundamentUuid).
					WillReturnError(erro.ErrorDbRepositoryError)
			},
			expectedSuccess: false,
			expectedErrors:  erro.ErrorDbRepositoryError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("failed to open sqlmock database: %v", err)
			}
			defer db.Close()

			mock.ExpectBegin()
			tx, err := db.Begin()
			if err != nil {
				t.Fatalf("failed to begin transaction: %v", err)
			}

			if tt.mockSetup != nil {
				tt.mockSetup(mock)
			}

			repo := NewAuthPostgresRepo(db)
			response := repo.DeleteUser(context.Background(), tx, tt.userid, tt.userpassword)

			assert.Equal(t, tt.expectedSuccess, response.Success)
			assert.Equal(t, tt.expectedErrors, response.Errors)

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}
