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
	kafkaProducer kafka.KafkaProducer
	validator     *validator.Validate
	grpcClient    *client.GrpcClient
}

func NewAuthService(repo repository.DBAuthenticateRepos, kafkaProd kafka.KafkaProducer, grpc *client.GrpcClient) *AuthService {
	validator := validator.New()
	return &AuthService{dbrepo: repo, validator: validator, kafkaProducer: kafkaProd, grpcClient: grpc}
}

type UserRegistrateEvent struct {
	UserID     uuid.UUID `json:"user_id"`
	LastUpdate time.Time `json:"last_update"`
}

func (as *AuthService) RegistrateAndLogin(ctx context.Context, user *model.Person) *ServiceResponse {
	registrateMap := make(map[string]error)
	var tx *sql.Tx
	var err error
	defer func() {
		if tx != nil {
			if err != nil {
				if rErr := as.dbrepo.RollbackTx(ctx, tx); rErr != nil {
					log.Printf("RegistrateAndLogin: Error rolling back transaction: %v", rErr)
				}
			} else {
				log.Println("RegistrateAndLogin: Transaction was successfully committed, no rollback needed")
			}
		}
	}()
	errorvalidate := validatePerson(as, user, true)
	if errorvalidate != nil {
		log.Printf("RegistrateAndLogin: Validate error %v", errorvalidate)
		return &ServiceResponse{Success: false, Errors: errorvalidate}
	}

	tx, err = as.dbrepo.BeginTx(ctx)
	if err != nil {
		log.Printf("RegistrateAndLogin: TransactionError %v", err)
		registrateMap["TransactionError"] = erro.ErrorStartTransaction
		return &ServiceResponse{Success: false, Errors: registrateMap}
	}

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
		registrateMap["GrpcResponseError"] = erro.ErrorHashPass
		return &ServiceResponse{Success: grpcresponse.Success, Errors: registrateMap}
	}
	err = as.dbrepo.CommitTx(ctx, tx)
	if err != nil {
		log.Printf("RegistrateAndLogin: Transaction commit error: %v", err)
		registrateMap["CommitError"] = erro.ErrorCommitTransaction
		return &ServiceResponse{Success: false, Errors: registrateMap}
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
	errorvalidate := validatePerson(as, user, false)
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
		authenticateMap["GrpcResponseError"] = erro.ErrorHashPass
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
	defer func() {
		if tx != nil {
			if err != nil {
				if rErr := as.dbrepo.RollbackTx(ctx, tx); rErr != nil {
					log.Printf("DeleteAccount: Error rolling back transaction: %v", rErr)
				}
			} else {
				log.Println("DeleteAccount: Transaction was successfully committed, no rollback needed")
			}
		}
	}()
	tx, err = as.dbrepo.BeginTx(ctx)
	if err != nil {
		log.Printf("DeleteAccount: TransactionError %v", err)
		deletemap["TransactionError"] = erro.ErrorStartTransaction
		return &ServiceResponse{Success: false, Errors: deletemap}
	}
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
	grpcresponse, err := as.grpcClient.DeleteSession(ctx, sessionID)
	if err != nil || !grpcresponse.Success {
		log.Printf("AuthenticateAndLogin: GrpcResponseError %v", err)
		deletemap["GrpcResponseError"] = erro.ErrorHashPass
		return &ServiceResponse{Success: grpcresponse.Success, Errors: deletemap}
	}
	return &ServiceResponse{
		Success: grpcresponse.Success,
	}
}

func validatePerson(as *AuthService, user *model.Person, flag bool) map[string]error {
	personToValidate := *user
	if !flag {
		personToValidate.Name = "qwertyuiopasdfghjklzxcvbn"
	}
	err := as.validator.Struct(&personToValidate)
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
