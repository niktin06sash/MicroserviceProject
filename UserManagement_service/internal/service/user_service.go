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
	dbrepo        repository.DBAuthenticateRepos
	dbtxmanager   repository.DBTransactionManager
	kafkaProducer kafka.KafkaProducer
	validator     *validator.Validate
	grpcClient    *client.GrpcClient
}

func NewAuthService(dbrepo repository.DBAuthenticateRepos, dbtxmanager repository.DBTransactionManager, kafkaProd kafka.KafkaProducer, grpc *client.GrpcClient) *AuthService {
	validator := validator.New()
	return &AuthService{dbrepo: dbrepo, dbtxmanager: dbtxmanager, validator: validator, kafkaProducer: kafkaProd, grpcClient: grpc}
}

type UserRegistrateEvent struct {
	UserID     uuid.UUID `json:"user_id"`
	LastUpdate time.Time `json:"last_update"`
}

func (as *AuthService) RegistrateAndLogin(ctx context.Context, user *model.Person) *ServiceResponse {
	registrateMap := make(map[string]error)
	var tx *sql.Tx
	errorvalidate := validatePerson(as.validator, user, true)
	if errorvalidate != nil {
		log.Printf("RegistrateAndLogin: Validate error %v", errorvalidate)
		return &ServiceResponse{Success: false, Errors: errorvalidate}
	}

	tx, err := as.dbtxmanager.BeginTx(ctx)
	if err != nil {
		log.Printf("RegistrateAndLogin: TransactionError %v", err)
		registrateMap["TransactionError"] = erro.ErrorStartTransaction
		return &ServiceResponse{Success: false, Errors: registrateMap}
	}
	defer func() {
		if r := recover(); r != nil {
			log.Printf("RegistrateAndLogin: Panic occurred: %v", r)
			rollbackTransaction(as.dbtxmanager, tx, "panic")
			panic(r)
		}
		commitOrRollbackTransaction(as.dbtxmanager, tx, err)
	}()
	hashpass, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("RegistrateAndLogin: HashPassError %v", err)
		registrateMap["HashPassError"] = erro.ErrorHashPass
		return &ServiceResponse{Success: false, Errors: registrateMap}
	}
	user.Password = string(hashpass)

	if ctx.Err() != nil {
		log.Printf("RegistrateAndLogin: Context cancelled before CreateUser: %v", ctx.Err())
		err = ctx.Err()
		registrateMap["ContextError"] = erro.ErrorContextTimeout
		return &ServiceResponse{Success: false, Errors: registrateMap}
	}

	userID := uuid.New()
	user.Id = userID
	response := as.dbrepo.CreateUser(ctx, user)

	if !response.Success {
		err = response.Errors
		log.Printf("RegistrateAndLogin: Error when creating person in the database %v", response.Errors)
		registrateMap["RegistrateError"] = response.Errors
		return &ServiceResponse{Success: response.Success, Errors: registrateMap}
	}
	if ctx.Err() != nil {
		log.Printf("RegistrateAndLogin: Context cancelled before CreateSession: %v", ctx.Err())
		err = ctx.Err()
		registrateMap["ContextError"] = erro.ErrorContextTimeout
		return &ServiceResponse{Success: false, Errors: registrateMap}
	}
	grpcresponse, err := as.grpcClient.CreateSession(ctx, user.Id.String())
	if err != nil || !grpcresponse.Success {
		log.Printf("RegistrateAndLogin: GrpcResponseError %v", err)
		registrateMap["GrpcResponseError"] = erro.ErrorGrpcResponse
		return &ServiceResponse{Success: grpcresponse.Success, Errors: registrateMap}
	}
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
	errorvalidate := validatePerson(as.validator, user, false)
	if errorvalidate != nil {
		log.Printf("AuthenticateAndLogin: Validate error %v", errorvalidate)
		return &ServiceResponse{Success: false, Errors: errorvalidate}
	}
	if ctx.Err() != nil {
		log.Printf("AuthenticateAndLogin: Context cancelled before GetUser: %v", ctx.Err())
		authenticateMap["ContextError"] = erro.ErrorContextTimeout
		return &ServiceResponse{Success: false, Errors: authenticateMap}
	}

	response := as.dbrepo.GetUser(ctx, user.Email, user.Password)
	if !response.Success {
		log.Printf("AuthenticateAndLogin: Failed to authenticate user: %v", response.Errors)
		authenticateMap["AuthenticateError"] = response.Errors
		return &ServiceResponse{Success: response.Success, Errors: authenticateMap}
	}
	if ctx.Err() != nil {
		log.Printf("AuthenticateAndLogin: Context cancelled before CreateSession: %v", ctx.Err())
		authenticateMap["ContextError"] = erro.ErrorContextTimeout
		return &ServiceResponse{Success: false, Errors: authenticateMap}
	}
	grpcresponse, err := as.grpcClient.CreateSession(ctx, user.Id.String())
	if err != nil || !grpcresponse.Success {
		log.Printf("AuthenticateAndLogin: GrpcResponseError %v", err)
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
	tx, err = as.dbtxmanager.BeginTx(ctx)
	if err != nil {
		log.Printf("DeleteAccount: TransactionError %v", err)
		deletemap["TransactionError"] = erro.ErrorStartTransaction
		return &ServiceResponse{Success: false, Errors: deletemap}
	}

	defer func() {
		if r := recover(); r != nil {
			log.Printf("DeleteAccount: Panic occurred: %v", r)
			rollbackTransaction(as.dbtxmanager, tx, "panic")
			panic(r)
		}
		commitOrRollbackTransaction(as.dbtxmanager, tx, err)
	}()

	if ctx.Err() != nil {
		err = ctx.Err()
		log.Printf("DeleteAccount: Context cancelled before DeleteUser: %v", ctx.Err())
		deletemap["ContextError"] = erro.ErrorContextTimeout
		return &ServiceResponse{Success: false, Errors: deletemap}
	}

	response := as.dbrepo.DeleteUser(ctx, userid, password)
	if !response.Success {
		err = response.Errors
		log.Printf("DeleteAccount: Failed to delete user: %v", response.Errors)
		deletemap["DeleteError"] = response.Errors
		return &ServiceResponse{Success: response.Success, Errors: deletemap}
	}

	if ctx.Err() != nil {
		err = ctx.Err()
		log.Printf("DeleteAccount: Context cancelled before DeleteSession: %v", ctx.Err())
		deletemap["ContextError"] = erro.ErrorContextTimeout
		return &ServiceResponse{Success: false, Errors: deletemap}
	}

	grpcresponse, gerr := as.grpcClient.DeleteSession(ctx, sessionID)
	if gerr != nil || !grpcresponse.Success {
		err = gerr
		log.Printf("AuthenticateAndLogin: GrpcResponseError %v", err)
		deletemap["GrpcResponseError"] = erro.ErrorGrpcResponse
		return &ServiceResponse{Success: grpcresponse.Success, Errors: deletemap}
	}

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

func commitOrRollbackTransaction(txMgr repository.DBTransactionManager, tx *sql.Tx, err error) {
	if tx != nil {
		if err != nil {
			rollbackTransaction(txMgr, tx, "error")
		} else {
			if cErr := txMgr.CommitTx(tx); cErr != nil {
				log.Printf("Error committing transaction: %v", cErr)
				rollbackTransaction(txMgr, tx, "commit failure")
			} else {
				log.Println("Transaction was successfully committed")
			}
		}
	}
}
