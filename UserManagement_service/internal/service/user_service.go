package service

import (
	"context"
	"database/sql"
	"fmt"
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
	KafkaProducer kafka.KafkaProducerService
	Validator     *validator.Validate
	GrpcClient    client.GrpcClientService
}

func NewAuthService(dbrepo repository.DBAuthenticateRepos, dbtxmanager repository.DBTransactionManager, kafkaProd kafka.KafkaProducerService, grpc *client.GrpcClient) *AuthService {
	validator := validator.New()
	return &AuthService{Dbrepo: dbrepo, Dbtxmanager: dbtxmanager, Validator: validator, KafkaProducer: kafkaProd, GrpcClient: grpc}
}
func (as *AuthService) RegistrateAndLogin(ctx context.Context, user *model.Person) *ServiceResponse {
	var place = "UseCase-RegistrateAndLogin"
	registrateMap := make(map[string]error)
	traceid := ctx.Value("traceID").(string)
	errorvalidate := validatePerson(as.Validator, user, true, traceid, place, as.KafkaProducer)
	if errorvalidate != nil {
		return &ServiceResponse{Success: false, Errors: errorvalidate, Type: erro.ClientErrorType}
	}
	var tx *sql.Tx
	tx, err := as.Dbtxmanager.BeginTx(ctx)
	if err != nil {
		fmterr := fmt.Sprintf("Transaction Error: %v", err)
		as.KafkaProducer.NewUserLog(kafka.LogLevelError, place, traceid, fmterr)
		registrateMap["InternalServerError"] = erro.ErrorStartTransaction
		return &ServiceResponse{Success: false, Errors: registrateMap, Type: erro.ServerErrorType}
	}
	isTransactionActive := true
	defer func() {
		if r := recover(); r != nil {
			fmterr := fmt.Sprintf("Panic occurred: %v", r)
			as.KafkaProducer.NewUserLog(kafka.LogLevelError, place, traceid, fmterr)
			if isTransactionActive {
				rollbackTransaction(as.Dbtxmanager, tx, traceid, place, as.KafkaProducer)
				isTransactionActive = false
			}
		}
		if isTransactionActive {
			rollbackTransaction(as.Dbtxmanager, tx, traceid, place, as.KafkaProducer)
			isTransactionActive = false
		} else {
			as.KafkaProducer.NewUserLog(kafka.LogLevelInfo, place, traceid, "Transaction was successfully committed and session received")
		}
	}()

	hashpass, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		fmterr := fmt.Sprintf("HashPass Error: %v", err)
		as.KafkaProducer.NewUserLog(kafka.LogLevelError, place, traceid, fmterr)
		registrateMap["InternalServerError"] = erro.ErrorHashPass
		return &ServiceResponse{Success: false, Errors: registrateMap, Type: erro.ServerErrorType}
	}
	user.Password = string(hashpass)
	userID := uuid.New()
	user.Id = userID
	bdresponse, serviceresponse := retryOperationDB(ctx, func(ctx context.Context) *repository.DBRepositoryResponse {
		return as.Dbrepo.CreateUser(ctx, tx, user)
	}, traceid, registrateMap, place, as.KafkaProducer)
	if serviceresponse != nil {
		return serviceresponse
	}
	userID = bdresponse.UserId
	grpcresponse, serviceresponse := retryOperationGrpc(ctx, func(ctx context.Context) (*proto.CreateSessionResponse, error) {
		return as.GrpcClient.CreateSession(ctx, userID.String())
	}, traceid, registrateMap, place, as.KafkaProducer)
	if serviceresponse != nil {
		return serviceresponse
	}
	if err := as.Dbtxmanager.CommitTx(tx); err != nil {
		fmterr := fmt.Sprintf("Error committing transaction: %v", err)
		as.KafkaProducer.NewUserLog(kafka.LogLevelError, place, traceid, fmterr)
		_, serviceresponse := retryOperationGrpc(ctx, func(ctx context.Context) (*proto.DeleteSessionResponse, error) {
			return as.GrpcClient.DeleteSession(ctx, grpcresponse.SessionID)
		}, traceid, registrateMap, place, as.KafkaProducer)
		if serviceresponse != nil {
			as.KafkaProducer.NewUserLog(kafka.LogLevelError, place, traceid, "Failed to delete session after transaction failure")
		}
		registrateMap["InternalServerError"] = erro.ErrorCommitTransaction
		return &ServiceResponse{Success: false, Errors: registrateMap, Type: erro.ServerErrorType}
	}
	isTransactionActive = false
	timeExpire := time.Unix(grpcresponse.ExpiryTime, 0)
	return &ServiceResponse{Success: true, UserId: bdresponse.UserId, SessionId: grpcresponse.SessionID, ExpireSession: timeExpire}
}
func (as *AuthService) AuthenticateAndLogin(ctx context.Context, user *model.Person) *ServiceResponse {
	var place = "UseCase-AuthenticateAndLogin"
	authenticateMap := make(map[string]error)
	traceid := ctx.Value("traceID").(string)
	errorvalidate := validatePerson(as.Validator, user, false, traceid, place, as.KafkaProducer)
	if errorvalidate != nil {
		return &ServiceResponse{Success: false, Errors: errorvalidate, Type: erro.ClientErrorType}
	}
	bdresponse, serviceresponse := retryOperationDB(ctx, func(ctx context.Context) *repository.DBRepositoryResponse {
		return as.Dbrepo.GetUser(ctx, user.Email, user.Password)
	}, traceid, authenticateMap, place, as.KafkaProducer)
	if serviceresponse != nil {
		return serviceresponse
	}
	userID := bdresponse.UserId
	grpcresponse, serviceresponse := retryOperationGrpc(ctx, func(ctx context.Context) (*proto.CreateSessionResponse, error) {
		return as.GrpcClient.CreateSession(ctx, userID.String())
	}, traceid, authenticateMap, place, as.KafkaProducer)
	if serviceresponse != nil {
		return serviceresponse
	}
	timeExpire := time.Unix(grpcresponse.ExpiryTime, 0)
	as.KafkaProducer.NewUserLog(kafka.LogLevelInfo, place, traceid, "The session was created successfully and received")
	return &ServiceResponse{Success: true, UserId: bdresponse.UserId, SessionId: grpcresponse.SessionID, ExpireSession: timeExpire}
}
func (as *AuthService) DeleteAccount(ctx context.Context, sessionID string, useridstr string, password string) *ServiceResponse {
	var place = "UseCase-DeleteAccount"
	deletemap := make(map[string]error)
	traceid := ctx.Value("traceID").(string)
	userid, err := uuid.Parse(useridstr)
	if err != nil {
		fmterr := fmt.Sprintf("UUID-parse Error: %v", err)
		as.KafkaProducer.NewUserLog(kafka.LogLevelError, place, traceid, fmterr)
		deletemap["InternalServerError"] = erro.ErrorMissingUserID
		return &ServiceResponse{Success: false, Errors: deletemap, Type: erro.ServerErrorType}
	}
	var tx *sql.Tx
	tx, err = as.Dbtxmanager.BeginTx(ctx)
	if err != nil {
		fmterr := fmt.Sprintf("Transaction Error: %v", err)
		as.KafkaProducer.NewUserLog(kafka.LogLevelError, place, traceid, fmterr)
		deletemap["InternalServerError"] = erro.ErrorStartTransaction
		return &ServiceResponse{Success: false, Errors: deletemap, Type: erro.ServerErrorType}
	}
	isTransactionActive := true
	defer func() {
		if r := recover(); r != nil {
			fmterr := fmt.Sprintf("Panic occurred: %v", r)
			as.KafkaProducer.NewUserLog(kafka.LogLevelError, place, traceid, fmterr)
			if isTransactionActive {
				rollbackTransaction(as.Dbtxmanager, tx, traceid, place, as.KafkaProducer)
				isTransactionActive = false
			}
		}
		if isTransactionActive {
			rollbackTransaction(as.Dbtxmanager, tx, traceid, place, as.KafkaProducer)
			isTransactionActive = false
		} else {
			as.KafkaProducer.NewUserLog(kafka.LogLevelInfo, place, traceid, "Transaction was successfully committed and user has successfully deleted his account with all data")
		}
	}()
	_, serviceresponse := retryOperationDB(ctx, func(ctx context.Context) *repository.DBRepositoryResponse {
		return as.Dbrepo.DeleteUser(ctx, tx, userid, password)
	}, traceid, deletemap, place, as.KafkaProducer)
	if serviceresponse != nil {
		return serviceresponse
	}
	grpcresponse, serviceresponse := retryOperationGrpc(ctx, func(ctx context.Context) (*proto.DeleteSessionResponse, error) {
		return as.GrpcClient.DeleteSession(ctx, sessionID)
	}, traceid, deletemap, place, as.KafkaProducer)
	if serviceresponse != nil {
		return serviceresponse
	}
	if err := as.Dbtxmanager.CommitTx(tx); err != nil {
		fmterr := fmt.Sprintf("Error committing transaction: %v", err)
		as.KafkaProducer.NewUserLog(kafka.LogLevelError, place, traceid, fmterr)
		deletemap["InternalServerError"] = erro.ErrorCommitTransaction
		return &ServiceResponse{Success: false, Errors: deletemap, Type: erro.ServerErrorType}
	}
	isTransactionActive = false
	return &ServiceResponse{
		Success: grpcresponse.Success,
	}
}
func (as *AuthService) Logout(ctx context.Context, sessionID string) *ServiceResponse {
	var place = "UseCase-Logout"
	logMap := make(map[string]error)
	traceid := ctx.Value("traceID").(string)
	grpcresponse, serviceresponse := retryOperationGrpc(ctx, func(ctx context.Context) (*proto.DeleteSessionResponse, error) {
		return as.GrpcClient.DeleteSession(ctx, sessionID)
	}, traceid, logMap, place, as.KafkaProducer)
	if serviceresponse != nil {
		return serviceresponse
	}
	as.KafkaProducer.NewUserLog(kafka.LogLevelInfo, place, traceid, "The session was deleted successfully")
	return &ServiceResponse{Success: grpcresponse.Success}
}
