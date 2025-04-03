package service

import (
	"UserManagement_service/internal/erro"
	"UserManagement_service/internal/kafka"
	"UserManagement_service/internal/model"
	"UserManagement_service/internal/repository"
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	dbrepo        repository.DBAuthenticateRepos
	kafkaProducer kafka.KafkaProducer
	validator     *validator.Validate
}

func NewAuthService(repo repository.DBAuthenticateRepos, kafkaProd kafka.KafkaProducer) *AuthService {
	validator := validator.New()
	return &AuthService{dbrepo: repo, validator: validator, kafkaProducer: kafkaProd}
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
					log.Printf("Error rolling back transaction: %v", rErr)
				}
			} else {

				fmt.Println("Transaction was successfully committed, no rollback needed")
			}
		}
	}()
	errorvalidate := validatePerson(as, user, true)
	if errorvalidate != nil {
		log.Printf("Validate error %v", errorvalidate)
		return &ServiceResponse{Success: false, Errors: errorvalidate}
	}

	tx, err = as.dbrepo.BeginTx(ctx)
	if err != nil {
		log.Printf("TransactionError %v", err)
		registrateMap["TransactionError"] = erro.ErrorStartTransaction
		return &ServiceResponse{Success: false, Errors: registrateMap}
	}

	hashpass, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("HashPassError %v", err)
		registrateMap["HashPassError"] = erro.ErrorHashPass
		return &ServiceResponse{Success: false, Errors: registrateMap}
	}
	user.Password = string(hashpass)

	if ctx.Err() != nil {
		log.Printf("Context cancelled before CreateUser: %v", ctx.Err())
		err = ctx.Err()
		registrateMap["ContextError"] = erro.ErrorContextTimeout
		return &ServiceResponse{Success: false, Errors: registrateMap}
	}

	userID := uuid.New()
	user.Id = userID
	response := as.dbrepo.CreateUser(ctx, user)

	if !response.Success {
		err = response.Errors
		log.Printf("Error when creating person in the database %v", response.Errors)
		registrateMap["RegistrateError"] = response.Errors
		return &ServiceResponse{Success: false, Errors: registrateMap}
	}
	//придумать логику для отправки запроса на микросервис с сессиями

	/*dbData, ok := response.Data.(repository.DBRepositoryResponseData)
	if !ok {
		err = erro.ErrorUnexpectedData
		log.Printf("Unexpected data type from repository: %T", response.Data)
		registrateMap["UnexpectedData"] = erro.ErrorUnexpectedData
		return &ServiceResponse{Success: false, Errors: registrateMap}
	}

	createdUserID := dbData.UserId

	if ctx.Err() != nil {
		log.Printf("Context cancelled before SetSession: %v", ctx.Err())
		err = ctx.Err()
		registrateMap["ContextError"] = erro.ErrorContextTimeout
		return &ServiceResponse{Success: false, Errors: registrateMap}
	}

	sessionID := uuid.New().String()
	expirationTime := time.Now().Add(time.Hour * 24)
	duration := time.Until(expirationTime)
	session := model.Session{
		SessionID:      sessionID,
		UserID:         createdUserID,
		ExpirationTime: expirationTime,
	}

	redisResponse := as.redisrepo.SetSession(ctx, session, duration)
	if !redisResponse.Success {
		err = redisResponse.Errors
		log.Printf("Error when creating a session in Redis: %v", redisResponse.Errors)
		registrateMap["SetSessionError"] = redisResponse.Errors
		return &ServiceResponse{Success: false, Errors: registrateMap}
	}

	redisData, ok := redisResponse.Data.(repository.RedisRepositoryResponseData)
	if !ok {
		err = erro.ErrorUnexpectedData
		log.Printf("Unexpected data type from repository: %T", response.Data)
		registrateMap["UnexpectedData"] = erro.ErrorUnexpectedData
		return &ServiceResponse{Success: false, Errors: registrateMap}
	}

	if ctx.Err() != nil {
		log.Printf("Context cancelled before CommitTx: %v", ctx.Err())
		err = ctx.Err()
		registrateMap["ContextError"] = erro.ErrorContextTimeout
		return &ServiceResponse{Success: false, Errors: registrateMap}
	}

	err = as.dbrepo.CommitTx(ctx, tx)
	if err != nil {
		log.Printf("Transaction commit error: %v", err)
		registrateMap["CommitError"] = erro.ErrorCommitTransaction
		return &ServiceResponse{Success: false, Errors: registrateMap}
	}

	log.Println("The session was created successfully and the user is registered!")
	event := UserRegistrateEvent{
		UserID:     dbData.UserId,
		LastUpdate: time.Now(),
	}
	eventBytes, errv := json.Marshal(event)
	if errv != nil {
		registrateMap["MarshalError"] = erro.ErrorMarshal
		log.Printf("Error marshaling event to JSON: %v", errv)
		return &ServiceResponse{Success: false, Errors: registrateMap}
	}
	errv = as.kafkaProducer.SendMessage("user-registered-topic", dbData.UserId.String(), eventBytes)
	if errv != nil {
		registrateMap["SendMessage"] = erro.ErrorSendKafkaMessage
		log.Printf("Error marshaling event to JSON: %v", errv)
		return &ServiceResponse{Success: false, Errors: registrateMap}
	}
	log.Println("Messages have been successfully delivered to the broker")
	return &ServiceResponse{
		Success:        true,
		UserId:         dbData.UserId,
		SessionId:      redisData.SessionId,
		ExpirationTime: redisData.ExpirationTime,
	}*/
}

type UserAuthenticateEvent struct {
	UserID     uuid.UUID `json:"user_id"`
	LastUpdate time.Time `json:"last_update"`
}

func (as *AuthService) AuthenticateAndLogin(ctx context.Context, user *model.Person) *ServiceResponse {
	authenticateMap := make(map[string]error)

	errorvalidate := validatePerson(as, user, false)
	if errorvalidate != nil {
		log.Printf("Validate error %v", errorvalidate)
		return &ServiceResponse{Success: false, Errors: errorvalidate}
	}
	if ctx.Err() != nil {
		log.Printf("AuthenticateAndLogin: Context cancelled before GetUser: %v", ctx.Err())
		authenticateMap["ContextError"] = erro.ErrorContextTimeout
		return &ServiceResponse{Success: false, Errors: authenticateMap}
	}

	response := as.dbrepo.GetUser(ctx, user.Email, user.Password)
	if !response.Success {
		log.Printf("Failed to authenticate user: %v", response.Errors)
		authenticateMap["AuthenticateError"] = response.Errors
		return &ServiceResponse{Success: false, Errors: authenticateMap}
	}
	//придумать логику для отправки запроса на микросервис с сессиями

	/*
		dbData, ok := response.Data.(repository.DBRepositoryResponseData)
		if !ok {
			log.Printf("Unexpected data type from repository: %T", response.Data)
			authenticateMap["UnexpectedData"] = erro.ErrorUnexpectedData
			return &ServiceResponse{Success: false, Errors: authenticateMap}
		}

		userID := dbData.UserId

		sessionID := uuid.New().String()
		expirationTime := time.Now().Add(24 * time.Hour)
		session := model.Session{
			SessionID:      sessionID,
			UserID:         userID,
			ExpirationTime: expirationTime,
		}

		duration := time.Until(expirationTime)

		if ctx.Err() != nil {
			log.Printf("AuthenticateAndLogin: Context cancelled before SetSession: %v", ctx.Err())
			authenticateMap["ContextError"] = erro.ErrorContextTimeout
			return &ServiceResponse{Success: false, Errors: authenticateMap}
		}

		redisResponse := as.redisrepo.SetSession(ctx, session, duration)
		if !redisResponse.Success {
			log.Printf("Error when creating a session in Redis: %v", redisResponse.Errors)
			authenticateMap["SetSessionError"] = redisResponse.Errors
			return &ServiceResponse{Success: false, Errors: authenticateMap}
		}

		redisData, ok := redisResponse.Data.(repository.RedisRepositoryResponseData)
		if !ok {
			log.Printf("Unexpected data type from repository: %T", redisResponse.Data)
			authenticateMap["UnexpectedData"] = erro.ErrorUnexpectedData
			return &ServiceResponse{Success: false, Errors: authenticateMap}
		}

		log.Println("The session was created successfully and the user is authenticated!")
		event := UserAuthenticateEvent{
			UserID:     dbData.UserId,
			LastUpdate: time.Now(),
		}
		eventBytes, errv := json.Marshal(event)
		if errv != nil {
			authenticateMap["MarshalError"] = erro.ErrorMarshal
			log.Printf("Error marshaling event to JSON: %v", errv)
			return &ServiceResponse{Success: false, Errors: authenticateMap}
		}
		errv = as.kafkaProducer.SendMessage("user-authenticate-topic", dbData.UserId.String(), eventBytes)
		if errv != nil {
			authenticateMap["SendMessage"] = erro.ErrorSendKafkaMessage
			log.Printf("Error marshaling event to JSON: %v", errv)
			return &ServiceResponse{Success: false, Errors: authenticateMap}
		}
		log.Println("Messages have been successfully delivered to the broker")
		return &ServiceResponse{
			Success:        true,
			UserId:         dbData.UserId,
			SessionId:      redisData.SessionId,
			ExpirationTime: redisData.ExpirationTime,
		}*/
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
					log.Printf("Error rolling back transaction: %v", rErr)
				}
			} else {

				fmt.Println("Transaction was successfully committed, no rollback needed")
			}
		}
	}()
	tx, err = as.dbrepo.BeginTx(ctx)
	if err != nil {
		log.Printf("TransactionError %v", err)
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
		log.Printf("Failed to delete user: %v", response.Errors)
		deletemap["DeleteError"] = response.Errors
		return &ServiceResponse{Success: false, Errors: deletemap}
	}
	if ctx.Err() != nil {
		err = ctx.Err()
		log.Printf("DeleteAccount: Context cancelled before DeleteUser: %v", ctx.Err())
		deletemap["ContextError"] = erro.ErrorContextTimeout
		return &ServiceResponse{Success: false, Errors: deletemap}
	}
	//придумать логику для отправки запроса на микросервис с сессиями

	/*repoResponse := as.redisrepo.DeleteSession(ctx, sessionID)
	if !repoResponse.Success {
		err = repoResponse.Errors
		log.Printf("Error during session deletion from Redis: %v", repoResponse.Errors)
		deletemap["DelSessionError"] = repoResponse.Errors
		return &ServiceResponse{Success: false, Errors: deletemap}
	}
	if ctx.Err() != nil {
		err = ctx.Err()
		log.Printf("DeleteAccount: Context cancelled before CommitTx: %v", ctx.Err())
		deletemap["ContextError"] = erro.ErrorContextTimeout
		return &ServiceResponse{Success: false, Errors: deletemap}
	}
	err = as.dbrepo.CommitTx(ctx, tx)
	if err != nil {
		log.Printf("Transaction commit error: %v", err)
		deletemap["CommitError"] = erro.ErrorCommitTransaction
		return &ServiceResponse{Success: false, Errors: deletemap}
	}
	log.Println("The account was successfully deleted with all data")
	event := UserAuthenticateEvent{
		UserID:     userid,
		LastUpdate: time.Now(),
	}
	eventBytes, errv := json.Marshal(event)
	if errv != nil {
		deletemap["MarshalError"] = erro.ErrorMarshal
		log.Printf("Error marshaling event to JSON: %v", errv)
		return &ServiceResponse{Success: false, Errors: deletemap}
	}
	errv = as.kafkaProducer.SendMessage("user-delete-topic", userid.String(), eventBytes)
	if errv != nil {
		deletemap["SendMessage"] = erro.ErrorSendKafkaMessage
		log.Printf("Error marshaling event to JSON: %v", errv)
		return &ServiceResponse{Success: false, Errors: deletemap}
	}
	log.Println("Messages have been successfully delivered to the broker")
	return &ServiceResponse{
		Success: true,
	}*/
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
