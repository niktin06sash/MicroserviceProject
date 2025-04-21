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
	"google.golang.org/grpc/metadata"

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
func (as *AuthService) RegistrateAndLogin(ctx context.Context, user *model.Person) *ServiceResponse {
	registrateMap := make(map[string]error)
	traceid := ctx.Value("traceID").(string)
	errorvalidate := validatePerson(as.Validator, user, true)
	if errorvalidate != nil {
		log.Printf("[ERROR] [UserManagement] [TraceID: %s]: RegistrateAndLogin: Validate error %v", traceid, errorvalidate)
		return &ServiceResponse{Success: false, Errors: errorvalidate, Type: erro.ClientErrorType}
	}
	var tx *sql.Tx
	tx, err := as.Dbtxmanager.BeginTx(ctx)
	if err != nil {
		log.Printf("[ERROR] [UserManagement] [TraceID: %s]: RegistrateAndLogin: TransactionError %v", traceid, err)
		registrateMap["TransactionError"] = erro.ErrorStartTransaction
		return &ServiceResponse{Success: false, Errors: registrateMap, Type: erro.ServerErrorType}
	}
	isTransactionActive := true
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[ERROR] [UserManagement] [TraceID: %s]: RegistrateAndLogin: Panic occurred: %v", traceid, r)
			if isTransactionActive {
				rollbackTransaction(as.Dbtxmanager, tx, traceid)
			}
		}
		if isTransactionActive {
			rollbackTransaction(as.Dbtxmanager, tx, traceid)
		}
		log.Printf("[INFO] [UserManagement] [TraceID: %s]: RegistrateAndLogin: Transaction was successfully committed", traceid)
		log.Printf("[INFO] [UserManagement] [TraceID: %s]: RegistrateAndLogin: The session was created successfully and the user is registered!", traceid)
	}()

	hashpass, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("[ERROR] [UserManagement] [TraceID: %s]: RegistrateAndLogin: HashPassError %v", traceid, err)
		registrateMap["HashPassError"] = erro.ErrorHashPass
		return &ServiceResponse{Success: false, Errors: registrateMap, Type: erro.ServerErrorType}
	}
	user.Password = string(hashpass)
	if ctxresponse, shouldReturn := checkContext(ctx, registrateMap); shouldReturn {
		log.Printf("[ERROR] [UserManagement] [TraceID: %s]: RegistrateAndLogin: Context cancelled before CreateUser: %v", traceid, ctx.Err())
		return ctxresponse
	}
	userID := uuid.New()
	user.Id = userID
	response := as.Dbrepo.CreateUser(ctx, tx, user)
	if !response.Success && response.Errors != nil {
		registrateMap["RegistrateError"] = response.Errors
		return &ServiceResponse{Success: response.Success, Errors: registrateMap, Type: response.Type}
	}
	userID = response.UserId
	if ctxresponse, shouldReturn := checkContext(ctx, registrateMap); shouldReturn {
		log.Printf("[ERROR] [UserManagement] [TraceID: %s]: RegistrateAndLogin: Context cancelled before CreateSession: %v", traceid, ctx.Err())
		return ctxresponse
	}
	md := metadata.Pairs("traceID", traceid)
	ctxgrpc := metadata.NewOutgoingContext(ctx, md)
	grpcresponse, err := as.GrpcClient.CreateSession(ctxgrpc, userID.String())
	if err != nil || !grpcresponse.Success {
		registrateMap["GrpcResponseError"] = erro.ErrorGrpcResponse
		return &ServiceResponse{Success: false, Errors: registrateMap, Type: erro.ServerErrorType}
	}
	if err := as.Dbtxmanager.CommitTx(tx); err != nil {
		log.Printf("[ERROR] [UserManagement] [TraceID: %s]: RegistrateAndLogin: Error committing transaction: %v", traceid, err)
		_, err := as.GrpcClient.DeleteSession(ctx, grpcresponse.SessionID)
		if err != nil {
			log.Printf("[ERROR] [UserManagement] [TraceID: %s]: RegistrateAndLogin: Failed to delete session after commit failure: %v", traceid, err)
			registrateMap["GrpcRollbackError"] = erro.ErrorGrpcRollback
			return &ServiceResponse{Success: false, Errors: registrateMap, Type: erro.ServerErrorType}
		}
		registrateMap["TransactionError"] = erro.ErrorCommitTransaction
		return &ServiceResponse{Success: false, Errors: registrateMap, Type: erro.ServerErrorType}
	}
	isTransactionActive = false
	timeExpire := time.Unix(grpcresponse.ExpiryTime, 0)
	return &ServiceResponse{Success: true, UserId: response.UserId, SessionId: grpcresponse.SessionID, ExpireSession: timeExpire}
}
func (as *AuthService) AuthenticateAndLogin(ctx context.Context, user *model.Person) *ServiceResponse {
	authenticateMap := make(map[string]error)
	traceid := ctx.Value("traceID").(string)
	errorvalidate := validatePerson(as.Validator, user, false)
	if errorvalidate != nil {
		log.Printf("[ERROR] [UserManagement] [TraceID: %s]: AuthenticateAndLogin: Validate error %v", traceid, errorvalidate)
		return &ServiceResponse{Success: false, Errors: errorvalidate, Type: erro.ClientErrorType}
	}
	if ctxresponse, shouldReturn := checkContext(ctx, authenticateMap); shouldReturn {
		log.Printf("[ERROR] [UserManagement] [TraceID: %s]: AuthenticateAndLogin: Context cancelled before GetUser: %v", traceid, ctx.Err())
		return ctxresponse
	}
	response := as.Dbrepo.GetUser(ctx, user.Email, user.Password)
	if !response.Success {
		authenticateMap["AuthenticateError"] = response.Errors
		return &ServiceResponse{Success: response.Success, Errors: authenticateMap, Type: response.Type}
	}
	userID := response.UserId
	if ctxresponse, shouldReturn := checkContext(ctx, authenticateMap); shouldReturn {
		log.Printf("[ERROR] [UserManagement] [TraceID: %s]: AuthenticateAndLogin: Context cancelled before CreateSession: %v", traceid, ctx.Err())
		return ctxresponse
	}
	md := metadata.Pairs("traceID", traceid)
	ctxgrpc := metadata.NewOutgoingContext(ctx, md)
	grpcresponse, err := as.GrpcClient.CreateSession(ctxgrpc, userID.String())
	if err != nil || !grpcresponse.Success {
		authenticateMap["GrpcResponseError"] = erro.ErrorGrpcResponse
		return &ServiceResponse{Success: false, Errors: authenticateMap, Type: erro.ServerErrorType}
	}
	timeExpire := time.Unix(grpcresponse.ExpiryTime, 0)
	log.Printf("[INFO] [UserManagement] [TraceID: %s]: AuthenticateAndLogin: The session was created successfully and the user is authenticated!", traceid)
	return &ServiceResponse{Success: true, UserId: response.UserId, SessionId: grpcresponse.SessionID, ExpireSession: timeExpire}
}
func (as *AuthService) DeleteAccount(ctx context.Context, sessionID string, userid uuid.UUID, password string) *ServiceResponse {
	deletemap := make(map[string]error)
	traceid := ctx.Value("traceID").(string)
	var tx *sql.Tx
	tx, err := as.Dbtxmanager.BeginTx(ctx)
	if err != nil {
		log.Printf("[ERROR] [UserManagement] [TraceID: %s]: DeleteAccount: TransactionError %v", traceid, err)
		deletemap["TransactionError"] = erro.ErrorStartTransaction
		return &ServiceResponse{Success: false, Errors: deletemap, Type: erro.ServerErrorType}
	}
	isTransactionActive := true
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[ERROR] [UserManagement] [TraceID: %s]: DeleteAccount: Panic occurred: %v", traceid, r)
			if isTransactionActive {
				rollbackTransaction(as.Dbtxmanager, tx, traceid)
			}
		}
		if isTransactionActive {
			rollbackTransaction(as.Dbtxmanager, tx, traceid)
		}
		log.Printf("[INFO] [UserManagement] [TraceID: %s]: DeleteAccount: Transaction was successfully committed", traceid)
		log.Printf("[INFO] [UserManagement] [TraceID: %s]: DeleteAccount: The user has successfully deleted his account with all data!", traceid)
	}()
	if ctxresponse, shouldReturn := checkContext(ctx, deletemap); shouldReturn {
		log.Printf("[ERROR] [UserManagement] [TraceID: %s]: DeleteAccount: Context cancelled before DeleteUser: %v", traceid, ctx.Err())
		return ctxresponse
	}
	response := as.Dbrepo.DeleteUser(ctx, tx, userid, password)
	if !response.Success && response.Errors != nil {
		deletemap["DeleteError"] = response.Errors
		return &ServiceResponse{Success: response.Success, Errors: deletemap, Type: response.Type}
	}
	if ctxresponse, shouldReturn := checkContext(ctx, deletemap); shouldReturn {
		log.Printf("[ERROR] [UserManagement] [TraceID: %s]: DeleteAccount: Context cancelled before DeleteSession: %v", traceid, ctx.Err())
		return ctxresponse
	}
	md := metadata.Pairs("traceID", traceid)
	ctxgrpc := metadata.NewOutgoingContext(ctx, md)
	grpcresponse, err := as.GrpcClient.DeleteSession(ctxgrpc, sessionID)
	if err != nil || !grpcresponse.Success {
		deletemap["GrpcResponseError"] = erro.ErrorGrpcResponse
		return &ServiceResponse{Success: false, Errors: deletemap, Type: erro.ServerErrorType}
	}
	if err := as.Dbtxmanager.CommitTx(tx); err != nil {
		log.Printf("[ERROR] [UserManagement] [TraceID: %s]: DeleteAccount: Error committing transaction: %v", traceid, err)
		deletemap["TransactionError"] = erro.ErrorCommitTransaction
		return &ServiceResponse{Success: false, Errors: deletemap, Type: erro.ServerErrorType}
	}
	isTransactionActive = false
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
func rollbackTransaction(txMgr repository.DBTransactionManager, tx *sql.Tx, traceid string) {
	if tx == nil {
		return
	}
	maxAttempts := 3
	attempt := 0
	for attempt < maxAttempts {
		attempt++
		err := txMgr.RollbackTx(tx)
		if err == nil {
			log.Printf("[INFO] [UserManagement] [TraceID: %s]: Successful rollback on attempt %d", traceid, attempt)
			return
		}
		log.Printf("[ERROR] [UserManagement] [TraceID: %s]: Error rolling back transaction on attempt %d: %v", traceid, attempt, err)
		if attempt == maxAttempts {
			log.Printf("[ERROR] [UserManagement] [TraceID: %s]: Failed to rollback transaction after %d attempts", traceid, maxAttempts)
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
}
func checkContext(ctx context.Context, Map map[string]error) (*ServiceResponse, bool) {
	select {
	case <-ctx.Done():
		Map["ContextError"] = erro.ErrorContextTimeout
		return &ServiceResponse{Success: false, Errors: Map, Type: erro.ServerErrorType}, true
	default:
		return nil, false
	}
}
