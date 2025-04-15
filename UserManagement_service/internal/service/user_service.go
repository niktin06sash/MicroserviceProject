package service

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/client"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/erro"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/kafka"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/model"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/repository"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	Dbrepo        repository.DBAuthenticateRepos
	Dbtxmanager   repository.DBTransactionManager
	KafkaProducer kafka.KafkaProducer
	Validator     *validator.Validate
	GrpcClient    client.GrpcClientService
}

func NewAuthService(dbrepo repository.DBAuthenticateRepos, dbtxmanager repository.DBTransactionManager, kafkaProd kafka.KafkaProducer, grpc *client.GrpcClient) *AuthService {
	validator := validator.New()
	return &AuthService{Dbrepo: dbrepo, Dbtxmanager: dbtxmanager, Validator: validator, KafkaProducer: kafkaProd, GrpcClient: grpc}
}

type UserRegistrateEvent struct {
	UserID     uuid.UUID `json:"user_id"`
	LastUpdate time.Time `json:"last_update"`
}

func (as *AuthService) RegistrateAndLogin(ctx context.Context, user *model.Person) *ServiceResponse {
	registrateMap := make(map[string]error)
	var tx *sql.Tx
	tx, err := as.Dbtxmanager.BeginTx(ctx)
	if err != nil {
		log.Printf("RegistrateAndLogin: TransactionError %v", err)
		registrateMap["TransactionError"] = erro.ErrorStartTransaction
		return &ServiceResponse{Success: false, Errors: registrateMap}
	}
	isTransactionActive := true
	defer func() {
		if r := recover(); r != nil {
			log.Printf("RegistrateAndLogin: Panic occurred: %v", r)
			if isTransactionActive {
				rollbackTransaction(as.Dbtxmanager, tx, "panic")
			}
			panic(r)
		}

		if isTransactionActive {
			rollbackTransaction(as.Dbtxmanager, tx, "failure")
		}
	}()

	hashpass, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("RegistrateAndLogin: HashPassError %v", err)
		if isTransactionActive {
			rollbackTransaction(as.Dbtxmanager, tx, "hash password failure")
			isTransactionActive = false
		}
		registrateMap["HashPassError"] = erro.ErrorHashPass
		return &ServiceResponse{Success: false, Errors: registrateMap}
	}
	user.Password = string(hashpass)

	if ctx.Err() != nil {
		log.Printf("RegistrateAndLogin: Context cancelled before CreateUser: %v", ctx.Err())
		if isTransactionActive {
			rollbackTransaction(as.Dbtxmanager, tx, "context timeout")
			isTransactionActive = false
		}
		registrateMap["ContextError"] = erro.ErrorContextTimeout
		return &ServiceResponse{Success: false, Errors: registrateMap}
	}

	userID := uuid.New()
	user.Id = userID
	response := as.Dbrepo.CreateUser(ctx, tx, user)
	if !response.Success {
		log.Printf("RegistrateAndLogin: Error when creating person in the database %v", response.Errors)
		if isTransactionActive {
			rollbackTransaction(as.Dbtxmanager, tx, "create user failure")
			isTransactionActive = false
		}
		registrateMap["RegistrateError"] = response.Errors
		return &ServiceResponse{Success: response.Success, Errors: registrateMap}
	}
	userID = response.UserId
	if ctx.Err() != nil {
		log.Printf("RegistrateAndLogin: Context cancelled before CreateSession: %v", ctx.Err())
		if isTransactionActive {
			rollbackTransaction(as.Dbtxmanager, tx, "context timeout")
			isTransactionActive = false
		}
		registrateMap["ContextError"] = erro.ErrorContextTimeout
		return &ServiceResponse{Success: false, Errors: registrateMap}
	}
	grpcresponse, err := as.GrpcClient.CreateSession(ctx, userID.String())
	if err != nil {
		log.Printf("RegistrateAndLogin: GrpcResponseError %v", err)
		if isTransactionActive {
			rollbackTransaction(as.Dbtxmanager, tx, "create session failure")
			isTransactionActive = false
		}
		registrateMap["GrpcResponseError"] = erro.ErrorGrpcResponse
		return &ServiceResponse{Success: false, Errors: registrateMap}
	}

	if !grpcresponse.Success {
		log.Printf("RegistrateAndLogin: GrpcResponseError: session creation failed")
		if isTransactionActive {
			rollbackTransaction(as.Dbtxmanager, tx, "create session failure")
			isTransactionActive = false
		}
		registrateMap["GrpcResponseError"] = erro.ErrorGrpcResponse
		return &ServiceResponse{Success: false, Errors: registrateMap}
	}

	if err := as.Dbtxmanager.CommitTx(tx); err != nil {
		log.Printf("RegistrateAndLogin: Error committing transaction: %v", err)
		if isTransactionActive {
			rollbackTransaction(as.Dbtxmanager, tx, "commit failure")
			isTransactionActive = false
		}
		_, err := as.GrpcClient.DeleteSession(ctx, grpcresponse.SessionID)
		if err != nil {
			log.Printf("RegistrateAndLogin: Failed to delete session after commit failure: %v", err)
			registrateMap["GrpcRollbackError"] = erro.ErrorGrpcRollback
			return &ServiceResponse{Success: false, Errors: registrateMap}
		}
		registrateMap["TransactionError"] = erro.ErrorCommitTransaction
		return &ServiceResponse{Success: false, Errors: registrateMap}
	}
	isTransactionActive = false
	log.Println("RegistrateAndLogin: Transaction was successfully committed")
	timeExpire := time.Unix(grpcresponse.ExpiryTime, 0)
	log.Println("RegistrateAndLogin: The session was created successfully and the user is registered!")
	return &ServiceResponse{Success: true, UserId: response.UserId, SessionId: grpcresponse.SessionID, ExpireSession: timeExpire}
}

type UserAuthenticateEvent struct {
	UserID     uuid.UUID `json:"user_id"`
	LastUpdate time.Time `json:"last_update"`
}

func (as *AuthService) AuthenticateAndLogin(ctx context.Context, user *model.Person) *ServiceResponse {
	authenticateMap := make(map[string]error)
	errorvalidate := validatePerson(as.Validator, user, false)
	if errorvalidate != nil {
		log.Printf("AuthenticateAndLogin: Validate error %v", errorvalidate)
		return &ServiceResponse{Success: false, Errors: errorvalidate}
	}
	if ctx.Err() != nil {
		log.Printf("AuthenticateAndLogin: Context cancelled before GetUser: %v", ctx.Err())
		authenticateMap["ContextError"] = erro.ErrorContextTimeout
		return &ServiceResponse{Success: false, Errors: authenticateMap}
	}

	response := as.Dbrepo.GetUser(ctx, user.Email, user.Password)
	if !response.Success {
		log.Printf("AuthenticateAndLogin: Failed to authenticate user: %v", response.Errors)
		authenticateMap["AuthenticateError"] = response.Errors
		return &ServiceResponse{Success: response.Success, Errors: authenticateMap}
	}
	userID := response.UserId
	if ctx.Err() != nil {
		log.Printf("AuthenticateAndLogin: Context cancelled before CreateSession: %v", ctx.Err())
		authenticateMap["ContextError"] = erro.ErrorContextTimeout
		return &ServiceResponse{Success: false, Errors: authenticateMap}
	}
	grpcresponse, err := as.GrpcClient.CreateSession(ctx, userID.String())
	if err != nil {
		log.Printf("AuthenticateAndLogin: GrpcResponseError %v", err)
		authenticateMap["GrpcResponseError"] = erro.ErrorGrpcResponse
		return &ServiceResponse{Success: false, Errors: authenticateMap}
	}
	if !grpcresponse.Success {
		log.Printf("AuthenticateAndLogin: GrpcResponseError: session creation failed")
		authenticateMap["GrpcResponseError"] = erro.ErrorGrpcResponse
		return &ServiceResponse{Success: false, Errors: authenticateMap}
	}
	timeExpire := time.Unix(grpcresponse.ExpiryTime, 0)
	log.Println("AuthenticateAndLogin: The session was created successfully and the user is authenticated!")
	return &ServiceResponse{Success: true, UserId: response.UserId, SessionId: grpcresponse.SessionID, ExpireSession: timeExpire}
}

type UserLogoutEvent struct {
	UserID     uuid.UUID `json:"user_id"`
	LastUpdate time.Time `json:"last_update"`
}

type UserDeleteEvent struct {
	UserID     uuid.UUID `json:"user_id"`
	LastUpdate time.Time `json:"last_update"`
}

func (as *AuthService) DeleteAccount(ctx context.Context, sessionID string, userid uuid.UUID, password string) *ServiceResponse {
	deletemap := make(map[string]error)
	var err error
	var tx *sql.Tx
	tx, err = as.Dbtxmanager.BeginTx(ctx)
	if err != nil {
		log.Printf("DeleteAccount: TransactionError %v", err)
		deletemap["TransactionError"] = erro.ErrorStartTransaction
		return &ServiceResponse{Success: false, Errors: deletemap}
	}

	isTransactionActive := true
	defer func() {
		if r := recover(); r != nil {
			log.Printf("DeleteAccount: Panic occurred: %v", r)
			if isTransactionActive {
				rollbackTransaction(as.Dbtxmanager, tx, "panic")
			}
			panic(r)
		}

		if isTransactionActive {
			rollbackTransaction(as.Dbtxmanager, tx, "failure")
		}
	}()

	if ctx.Err() != nil {
		if isTransactionActive {
			rollbackTransaction(as.Dbtxmanager, tx, "context timeout")
			isTransactionActive = false
		}
		log.Printf("DeleteAccount: Context cancelled before DeleteUser: %v", ctx.Err())
		deletemap["ContextError"] = erro.ErrorContextTimeout
		return &ServiceResponse{Success: false, Errors: deletemap}
	}

	response := as.Dbrepo.DeleteUser(ctx, tx, userid, password)
	if !response.Success {
		if isTransactionActive {
			rollbackTransaction(as.Dbtxmanager, tx, "delete user failure")
			isTransactionActive = false
		}
		log.Printf("DeleteAccount: Failed to delete user: %v", response.Errors)
		deletemap["DeleteError"] = response.Errors
		return &ServiceResponse{Success: response.Success, Errors: deletemap}
	}

	if ctx.Err() != nil {
		if isTransactionActive {
			rollbackTransaction(as.Dbtxmanager, tx, "context timeout")
			isTransactionActive = false
		}
		log.Printf("DeleteAccount: Context cancelled before DeleteSession: %v", ctx.Err())
		deletemap["ContextError"] = erro.ErrorContextTimeout
		return &ServiceResponse{Success: false, Errors: deletemap}
	}

	grpcresponse, err := as.GrpcClient.DeleteSession(ctx, sessionID)
	if err != nil {
		log.Printf("DeleteAccount: GrpcResponseError %v", err)
		if isTransactionActive {
			rollbackTransaction(as.Dbtxmanager, tx, "delete session failure")
			isTransactionActive = false
		}
		deletemap["GrpcResponseError"] = erro.ErrorGrpcResponse
		return &ServiceResponse{Success: false, Errors: deletemap}
	}

	if !grpcresponse.Success {
		log.Printf("DeleteAccount: GrpcResponseError: session creation failed")
		if isTransactionActive {
			rollbackTransaction(as.Dbtxmanager, tx, "delete session failure")
			isTransactionActive = false
		}
		deletemap["GrpcResponseError"] = erro.ErrorGrpcResponse
		return &ServiceResponse{Success: false, Errors: deletemap}
	}

	if err := as.Dbtxmanager.CommitTx(tx); err != nil {
		if isTransactionActive {
			rollbackTransaction(as.Dbtxmanager, tx, "commit failure")
			isTransactionActive = false
		}
		log.Printf("DeleteAccount: Error committing transaction: %v", err)
		deletemap["TransactionError"] = erro.ErrorCommitTransaction
		return &ServiceResponse{Success: false, Errors: deletemap}
	}
	isTransactionActive = false
	log.Println("DeleteAccount:  Transaction was successfully committed")
	log.Println("DeleteAccount: The user has successfully deleted his account with all data!")
	return &ServiceResponse{
		Success: grpcresponse.Success,
	}
}
func validatePerson(val *validator.Validate, user *model.Person, flag bool) map[string]error {
	personToValidate := *user
	if !flag {
		personToValidate.Name = "qwertyuiopasdfghjklzxcvbn"
	}
	err := val.Struct(&personToValidate)
	if err != nil {
		validationErrors, ok := err.(validator.ValidationErrors)
		if ok {
			erors := make(map[string]error)
			for _, err := range validationErrors {

				switch err.Tag() {

				case "email":
					log.Println("Email format error")
					erors[err.Field()] = erro.ErrorNotEmail
				case "min":
					errv := fmt.Errorf("%s is too short", err.Field())
					log.Println(err.Field() + " format error")
					erors[err.Field()] = errv

				default:
					errv := fmt.Errorf("%s is Null", err.Field())
					log.Println(err.Field() + " format error")
					erors[err.Field()] = errv
				}
			}
			return erors
		}
	}
	return nil
}
func rollbackTransaction(txMgr repository.DBTransactionManager, tx *sql.Tx, reason string) {
	if tx != nil {
		if err := txMgr.RollbackTx(tx); err != nil {
			log.Printf("Error rolling back transaction (%s): %v", reason, err)
		} else {
			log.Printf("Successful rollback (%s)", reason)
		}
	}
}
