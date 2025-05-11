package service

import (
	"context"
	"database/sql"
	"log"
	"time"

	"github.com/niktin06sash/MicroserviceProject/SessionManagement_service/proto"
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
func (as *AuthService) RegistrateAndLogin(ctx context.Context, user *model.Person) *ServiceResponse {
	registrateMap := make(map[string]error)
	traceid := ctx.Value("traceID").(string)
	errorvalidate := validatePerson(as.Validator, user, true, traceid, "RegistrateAndLogin")
	if errorvalidate != nil {
		log.Printf("[ERROR] [UserManagement] [TraceID: %s] RegistrateAndLogin: Validate error %v", traceid, errorvalidate)
		return &ServiceResponse{Success: false, Errors: errorvalidate, Type: erro.ClientErrorType}
	}
	var tx *sql.Tx
	tx, err := as.Dbtxmanager.BeginTx(ctx)
	if err != nil {
		log.Printf("[ERROR] [UserManagement] [TraceID: %s] RegistrateAndLogin: TransactionError %v", traceid, err)
		registrateMap["InternalServerError"] = erro.ErrorStartTransaction
		return &ServiceResponse{Success: false, Errors: registrateMap, Type: erro.ServerErrorType}
	}
	isTransactionActive := true
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[ERROR] [UserManagement] [TraceID: %s] RegistrateAndLogin: Panic occurred: %v", traceid, r)
			if isTransactionActive {
				rollbackTransaction(as.Dbtxmanager, tx, traceid, "RegistrateAndLogin")
				isTransactionActive = false
			}
		}
		if isTransactionActive {
			rollbackTransaction(as.Dbtxmanager, tx, traceid, "RegistrateAndLogin")
			isTransactionActive = false
		} else {
			log.Printf("[INFO] [UserManagement] [TraceID: %s] RegistrateAndLogin: Transaction was successfully committed", traceid)
			log.Printf("[INFO] [UserManagement] [TraceID: %s] RegistrateAndLogin: The session was created successfully and the user is registered!", traceid)
		}
	}()

	hashpass, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("[ERROR] [UserManagement] [TraceID: %s] RegistrateAndLogin: HashPassError %v", traceid, err)
		registrateMap["InternalServerError"] = erro.ErrorHashPass
		return &ServiceResponse{Success: false, Errors: registrateMap, Type: erro.ServerErrorType}
	}
	user.Password = string(hashpass)

	userID := uuid.New()
	user.Id = userID
	bdresponse, serviceresponse := retryOperationDB(ctx, func(ctx context.Context) *repository.DBRepositoryResponse {
		return as.Dbrepo.CreateUser(ctx, tx, user)
	}, traceid, registrateMap, "RegistrateAndLogin")
	if serviceresponse != nil {
		return serviceresponse
	}
	userID = bdresponse.UserId
	grpcresponse, serviceresponse := retryOperationGrpc(ctx, func(ctx context.Context) (*proto.CreateSessionResponse, error) {
		return as.GrpcClient.CreateSession(ctx, userID.String())
	}, traceid, registrateMap, "RegistrateAndLogin")
	if serviceresponse != nil {
		return serviceresponse
	}
	if err := as.Dbtxmanager.CommitTx(tx); err != nil {
		log.Printf("[ERROR] [UserManagement] [TraceID: %s] RegistrateAndLogin: Error committing transaction: %v", traceid, err)
		_, serviceresponse := retryOperationGrpc(ctx, func(ctx context.Context) (*proto.DeleteSessionResponse, error) {
			return as.GrpcClient.DeleteSession(ctx, grpcresponse.SessionID)
		}, traceid, registrateMap, "RegistrateAndLogin")
		if serviceresponse != nil {
			log.Printf("[ERROR] [UserManagement] [TraceID: %s] RegistrateAndLogin: Failed to delete session after transaction failure: %v", traceid, serviceresponse.Errors)
		}
		registrateMap["InternalServerError"] = erro.ErrorCommitTransaction
		return &ServiceResponse{Success: false, Errors: registrateMap, Type: erro.ServerErrorType}
	}
	isTransactionActive = false
	timeExpire := time.Unix(grpcresponse.ExpiryTime, 0)
	return &ServiceResponse{Success: true, UserId: bdresponse.UserId, SessionId: grpcresponse.SessionID, ExpireSession: timeExpire}
}
func (as *AuthService) AuthenticateAndLogin(ctx context.Context, user *model.Person) *ServiceResponse {
	authenticateMap := make(map[string]error)
	traceid := ctx.Value("traceID").(string)
	errorvalidate := validatePerson(as.Validator, user, false, traceid, "AuthenticateAndLogin")
	if errorvalidate != nil {
		log.Printf("[ERROR] [UserManagement] [TraceID: %s] AuthenticateAndLogin: Validate error %v", traceid, errorvalidate)
		return &ServiceResponse{Success: false, Errors: errorvalidate, Type: erro.ClientErrorType}
	}
	bdresponse, serviceresponse := retryOperationDB(ctx, func(ctx context.Context) *repository.DBRepositoryResponse {
		return as.Dbrepo.GetUser(ctx, user.Email, user.Password)
	}, traceid, authenticateMap, "AuthenticateAndLogin")
	if serviceresponse != nil {
		return serviceresponse
	}
	userID := bdresponse.UserId
	grpcresponse, serviceresponse := retryOperationGrpc(ctx, func(ctx context.Context) (*proto.CreateSessionResponse, error) {
		return as.GrpcClient.CreateSession(ctx, userID.String())
	}, traceid, authenticateMap, "AuthenticateAndLogin")
	if serviceresponse != nil {
		return serviceresponse
	}
	timeExpire := time.Unix(grpcresponse.ExpiryTime, 0)
	log.Printf("[INFO] [UserManagement] [TraceID: %s] AuthenticateAndLogin: The session was created successfully and the user is authenticated!", traceid)
	return &ServiceResponse{Success: true, UserId: bdresponse.UserId, SessionId: grpcresponse.SessionID, ExpireSession: timeExpire}
}
func (as *AuthService) DeleteAccount(ctx context.Context, sessionID string, useridstr string, password string) *ServiceResponse {
	deletemap := make(map[string]error)
	traceid := ctx.Value("traceID").(string)
	userid, err := uuid.Parse(useridstr)
	if err != nil {
		deletemap["InternalServerError"] = erro.ErrorMissingUserID
		return &ServiceResponse{Success: false, Errors: deletemap, Type: erro.ServerErrorType}
	}
	var tx *sql.Tx
	tx, err = as.Dbtxmanager.BeginTx(ctx)
	if err != nil {
		log.Printf("[ERROR] [UserManagement] [TraceID: %s] DeleteAccount: TransactionError %v", traceid, err)
		deletemap["InternalServerError"] = erro.ErrorStartTransaction
		return &ServiceResponse{Success: false, Errors: deletemap, Type: erro.ServerErrorType}
	}
	isTransactionActive := true
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[ERROR] [UserManagement] [TraceID: %s] DeleteAccount: Panic occurred: %v", traceid, r)
			if isTransactionActive {
				rollbackTransaction(as.Dbtxmanager, tx, traceid, "DeleteAccount")
				isTransactionActive = false
			}
		}
		if isTransactionActive {
			rollbackTransaction(as.Dbtxmanager, tx, traceid, "DeleteAccount")
			isTransactionActive = false
		} else {
			log.Printf("[INFO] [UserManagement] [TraceID: %s] DeleteAccount: Transaction was successfully committed", traceid)
			log.Printf("[INFO] [UserManagement] [TraceID: %s] DeleteAccount: The user has successfully deleted his account with all data!", traceid)
		}
	}()
	_, serviceresponse := retryOperationDB(ctx, func(ctx context.Context) *repository.DBRepositoryResponse {
		return as.Dbrepo.DeleteUser(ctx, tx, userid, password)
	}, traceid, deletemap, "DeleteAccount")
	if serviceresponse != nil {
		return serviceresponse
	}
	grpcresponse, serviceresponse := retryOperationGrpc(ctx, func(ctx context.Context) (*proto.DeleteSessionResponse, error) {
		return as.GrpcClient.DeleteSession(ctx, sessionID)
	}, traceid, deletemap, "DeleteAccount")
	if serviceresponse != nil {
		return serviceresponse
	}
	if err := as.Dbtxmanager.CommitTx(tx); err != nil {
		log.Printf("[ERROR] [UserManagement] [TraceID: %s] DeleteAccount: Error committing transaction: %v", traceid, err)
		deletemap["InternalServerError"] = erro.ErrorCommitTransaction
		return &ServiceResponse{Success: false, Errors: deletemap, Type: erro.ServerErrorType}
	}
	isTransactionActive = false
	return &ServiceResponse{
		Success: grpcresponse.Success,
	}
}
func (as *AuthService) Logout(ctx context.Context, sessionID string) *ServiceResponse {
	logMap := make(map[string]error)
	traceid := ctx.Value("traceID").(string)
	grpcresponse, serviceresponse := retryOperationGrpc(ctx, func(ctx context.Context) (*proto.DeleteSessionResponse, error) {
		return as.GrpcClient.DeleteSession(ctx, sessionID)
	}, traceid, logMap, "Logout")
	if serviceresponse != nil {
		return serviceresponse
	}
	log.Printf("[INFO] [UserManagement] [TraceID: %s] Logout: The user has successfully logged out of the account!", traceid)
	return &ServiceResponse{Success: grpcresponse.Success}
}
