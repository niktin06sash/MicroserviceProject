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
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/metrics"
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
	registrateMap := make(map[string]error)
	traceid := ctx.Value("traceID").(string)
	errorvalidate := validatePerson(as.Validator, user, true, traceid, RegistrateAndLogin, as.KafkaProducer)
	if errorvalidate != nil {
		return &ServiceResponse{Success: false, Errors: errorvalidate, Type: erro.ClientErrorType}
	}
	hashpass, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		fmterr := fmt.Sprintf("HashPass Error: %v", err)
		as.KafkaProducer.NewUserLog(kafka.LogLevelError, RegistrateAndLogin, traceid, fmterr)
		registrateMap["InternalServerError"] = fmt.Errorf(erro.UserServiceUnavalaible)
		metrics.UserErrorsTotal.WithLabelValues("InternalServerError").Inc()
		return &ServiceResponse{Success: false, Errors: registrateMap, Type: erro.ServerErrorType}
	}
	user.Password = string(hashpass)
	userID := uuid.New()
	user.Id = userID
	var tx *sql.Tx
	tx, err = as.Dbtxmanager.BeginTx(ctx)
	metrics.UserDBQueriesTotal.WithLabelValues("Begin Transaction").Inc()
	if err != nil {
		fmterr := fmt.Sprintf("Transaction Error: %v", err)
		as.KafkaProducer.NewUserLog(kafka.LogLevelError, RegistrateAndLogin, traceid, fmterr)
		registrateMap["InternalServerError"] = fmt.Errorf("User-Service is unavailable")
		metrics.UserDBErrorsTotal.WithLabelValues("Begin Transaction", "Transaction").Inc()
		metrics.UserErrorsTotal.WithLabelValues("InternalServerError").Inc()
		return &ServiceResponse{Success: false, Errors: registrateMap, Type: erro.ServerErrorType}
	}
	isTransactionActive := true
	defer func() {
		if isTransactionActive {
			rollbackTransaction(as.Dbtxmanager, tx, traceid, RegistrateAndLogin, as.KafkaProducer)
			isTransactionActive = false
		} else {
			as.KafkaProducer.NewUserLog(kafka.LogLevelInfo, RegistrateAndLogin, traceid, "Transaction was successfully committed and session received")
		}
	}()
	bdresponse := as.Dbrepo.CreateUser(ctx, tx, user)
	if !bdresponse.Success && bdresponse.Errors != nil {
		if bdresponse.Type == erro.ServerErrorType {
			registrateMap["InternalServerError"] = bdresponse.Errors
			metrics.UserErrorsTotal.WithLabelValues("InternalServerError").Inc()
			return &ServiceResponse{Success: bdresponse.Success, Errors: registrateMap, Type: bdresponse.Type}
		}
		registrateMap["ClientError"] = bdresponse.Errors
		metrics.UserErrorsTotal.WithLabelValues("ClientError").Inc()
		return &ServiceResponse{Success: bdresponse.Success, Errors: registrateMap, Type: bdresponse.Type}
	}
	userID = bdresponse.UserId
	grpcresponse, serviceresponse := retryOperationGrpc(ctx, func(ctx context.Context) (*proto.CreateSessionResponse, error) {
		return as.GrpcClient.CreateSession(ctx, userID.String())
	}, traceid, registrateMap, RegistrateAndLogin, as.KafkaProducer)
	if serviceresponse != nil {
		return serviceresponse
	}
	err = commitTransaction(as.Dbtxmanager, tx, traceid, RegistrateAndLogin, as.KafkaProducer)
	if err != nil {
		_, serviceresponse := retryOperationGrpc(ctx, func(ctx context.Context) (*proto.DeleteSessionResponse, error) {
			return as.GrpcClient.DeleteSession(ctx, grpcresponse.SessionID)
		}, traceid, registrateMap, RegistrateAndLogin, as.KafkaProducer)
		if serviceresponse != nil {
			as.KafkaProducer.NewUserLog(kafka.LogLevelError, RegistrateAndLogin, traceid, "Failed to delete session after transaction failure")
		}
		registrateMap["InternalServerError"] = fmt.Errorf(erro.UserServiceUnavalaible)
		return &ServiceResponse{Success: false, Errors: registrateMap, Type: erro.ServerErrorType}
	}
	isTransactionActive = false
	timeExpire := time.Unix(grpcresponse.ExpiryTime, 0)
	return &ServiceResponse{Success: true, UserId: bdresponse.UserId, SessionId: grpcresponse.SessionID, ExpireSession: timeExpire}
}
func (as *AuthService) AuthenticateAndLogin(ctx context.Context, user *model.Person) *ServiceResponse {

	authenticateMap := make(map[string]error)
	traceid := ctx.Value("traceID").(string)
	errorvalidate := validatePerson(as.Validator, user, false, traceid, AuthenticateAndLogin, as.KafkaProducer)
	if errorvalidate != nil {
		return &ServiceResponse{Success: false, Errors: errorvalidate, Type: erro.ClientErrorType}
	}
	bdresponse := as.Dbrepo.GetUser(ctx, user.Email, user.Password)
	if !bdresponse.Success && bdresponse.Errors != nil {
		if bdresponse.Type == erro.ServerErrorType {
			authenticateMap["InternalServerError"] = bdresponse.Errors
			metrics.UserErrorsTotal.WithLabelValues("InternalServerError").Inc()
			return &ServiceResponse{Success: bdresponse.Success, Errors: authenticateMap, Type: bdresponse.Type}
		}
		authenticateMap["ClientError"] = bdresponse.Errors
		metrics.UserErrorsTotal.WithLabelValues("ClientError").Inc()
		return &ServiceResponse{Success: bdresponse.Success, Errors: authenticateMap, Type: bdresponse.Type}
	}
	userID := bdresponse.UserId
	grpcresponse, serviceresponse := retryOperationGrpc(ctx, func(ctx context.Context) (*proto.CreateSessionResponse, error) {
		return as.GrpcClient.CreateSession(ctx, userID.String())
	}, traceid, authenticateMap, AuthenticateAndLogin, as.KafkaProducer)
	if serviceresponse != nil {
		return serviceresponse
	}
	timeExpire := time.Unix(grpcresponse.ExpiryTime, 0)
	as.KafkaProducer.NewUserLog(kafka.LogLevelInfo, AuthenticateAndLogin, traceid, "The session was created successfully and received")
	return &ServiceResponse{Success: true, UserId: bdresponse.UserId, SessionId: grpcresponse.SessionID, ExpireSession: timeExpire}
}
func (as *AuthService) DeleteAccount(ctx context.Context, sessionID string, useridstr string, password string) *ServiceResponse {
	deletemap := make(map[string]error)
	traceid := ctx.Value("traceID").(string)
	userid, err := uuid.Parse(useridstr)
	if err != nil {
		fmterr := fmt.Sprintf("UUID-parse Error: %v", err)
		as.KafkaProducer.NewUserLog(kafka.LogLevelError, DeleteAccount, traceid, fmterr)
		deletemap["InternalServerError"] = fmt.Errorf(erro.UserServiceUnavalaible)
		metrics.UserErrorsTotal.WithLabelValues("InternalServerError").Inc()
		return &ServiceResponse{Success: false, Errors: deletemap, Type: erro.ServerErrorType}
	}
	var tx *sql.Tx
	tx, err = as.Dbtxmanager.BeginTx(ctx)
	metrics.UserDBQueriesTotal.WithLabelValues("Begin Transaction").Inc()
	if err != nil {
		fmterr := fmt.Sprintf("Transaction Error: %v", err)
		as.KafkaProducer.NewUserLog(kafka.LogLevelError, DeleteAccount, traceid, fmterr)
		deletemap["InternalServerError"] = fmt.Errorf(erro.UserServiceUnavalaible)
		metrics.UserDBErrorsTotal.WithLabelValues("Begin Transaction", "Transaction").Inc()
		metrics.UserErrorsTotal.WithLabelValues("InternalServerError").Inc()
		return &ServiceResponse{Success: false, Errors: deletemap, Type: erro.ServerErrorType}
	}
	isTransactionActive := true
	defer func() {
		if isTransactionActive {
			rollbackTransaction(as.Dbtxmanager, tx, traceid, DeleteAccount, as.KafkaProducer)
			isTransactionActive = false
		} else {
			as.KafkaProducer.NewUserLog(kafka.LogLevelInfo, DeleteAccount, traceid, "Transaction was successfully committed and user has successfully deleted his account with all data")
		}
	}()
	bdresponse := as.Dbrepo.DeleteUser(ctx, tx, userid, password)
	if !bdresponse.Success && bdresponse.Errors != nil {
		if bdresponse.Type == erro.ServerErrorType {
			deletemap["InternalServerError"] = bdresponse.Errors
			metrics.UserErrorsTotal.WithLabelValues("InternalServerError").Inc()
			return &ServiceResponse{Success: bdresponse.Success, Errors: deletemap, Type: bdresponse.Type}
		}
		deletemap["ClientError"] = bdresponse.Errors
		metrics.UserErrorsTotal.WithLabelValues("ClientError").Inc()
		return &ServiceResponse{Success: bdresponse.Success, Errors: deletemap, Type: bdresponse.Type}
	}
	grpcresponse, serviceresponse := retryOperationGrpc(ctx, func(ctx context.Context) (*proto.DeleteSessionResponse, error) {
		return as.GrpcClient.DeleteSession(ctx, sessionID)
	}, traceid, deletemap, DeleteAccount, as.KafkaProducer)
	if serviceresponse != nil {
		return serviceresponse
	}
	err = commitTransaction(as.Dbtxmanager, tx, traceid, DeleteAccount, as.KafkaProducer)
	if err != nil {
		deletemap["InternalServerError"] = fmt.Errorf(erro.UserServiceUnavalaible)
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
	}, traceid, logMap, Logout, as.KafkaProducer)
	if serviceresponse != nil {
		return serviceresponse
	}
	as.KafkaProducer.NewUserLog(kafka.LogLevelInfo, Logout, traceid, "The session was deleted successfully")
	return &ServiceResponse{Success: grpcresponse.Success}
}
func (as *AuthService) UpdateAccount(ctx context.Context, data map[string]string, updateType string) *ServiceResponse {
	traceid := ctx.Value("traceID").(string)
	updatemap := make(map[string]error)
	name, name_ok := data["name"]
	email, email_ok := data["email"]
	last_password, last_password_ok := data["last_password"]
	new_password, new_password_ok := data["new_password"]
	if name_ok && !last_password_ok && !new_password_ok && !email_ok && updateType == "name" {

	}
	if !name_ok && last_password_ok && new_password_ok && !email_ok && updateType == "password" {

	}
	if !name_ok && last_password_ok && !new_password_ok && email_ok && updateType == "email" {

	}
	fmterr := fmt.Sprintf("Invalid data in request")
	updatemap["ClientError"] = fmt.Errorf(fmterr)
	as.KafkaProducer.NewUserLog(kafka.LogLevelWarn, UpdateAccount, traceid, fmterr)
	metrics.UserErrorsTotal.WithLabelValues("ClientError").Inc()
	return &ServiceResponse{Success: false, Errors: updatemap, Type: erro.ClientErrorType}
}
