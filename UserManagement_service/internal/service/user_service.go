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
	const place = RegistrateAndLogin
	registrateMap := make(map[string]string)
	traceid := ctx.Value("traceID").(string)
	errorvalidate := validatePerson(as.Validator, user, true, traceid, place, as.KafkaProducer)
	if errorvalidate != nil {
		return &ServiceResponse{Success: false, Errors: errorvalidate, ErrorType: erro.ClientErrorType}
	}
	hashpass, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		fmterr := fmt.Sprintf("HashPass Error: %v", err)
		as.KafkaProducer.NewUserLog(kafka.LogLevelError, place, traceid, fmterr)
		registrateMap[erro.ServerErrorType] = erro.UserServiceUnavalaible
		metrics.UserErrorsTotal.WithLabelValues(erro.ServerErrorType).Inc()
		return &ServiceResponse{Success: false, Errors: registrateMap, ErrorType: erro.ServerErrorType}
	}
	user.Password = string(hashpass)
	userID := uuid.New()
	user.Id = userID
	var tx *sql.Tx
	tx, err = as.Dbtxmanager.BeginTx(ctx)
	metrics.UserDBQueriesTotal.WithLabelValues("Begin Transaction").Inc()
	if err != nil {
		fmterr := fmt.Sprintf("Transaction Error: %v", err)
		as.KafkaProducer.NewUserLog(kafka.LogLevelError, place, traceid, fmterr)
		registrateMap[erro.ServerErrorType] = erro.UserServiceUnavalaible
		metrics.UserDBErrorsTotal.WithLabelValues("Begin Transaction", "Transaction").Inc()
		metrics.UserErrorsTotal.WithLabelValues(erro.ServerErrorType).Inc()
		return &ServiceResponse{Success: false, Errors: registrateMap, ErrorType: erro.ServerErrorType}
	}
	isTransactionActive := true
	defer func() {
		if isTransactionActive {
			rollbackTransaction(as.Dbtxmanager, tx, traceid, place, as.KafkaProducer)
			isTransactionActive = false
		} else {
			as.KafkaProducer.NewUserLog(kafka.LogLevelInfo, place, traceid, "Transaction was successfully committed and session received")
		}
	}()
	bdresponse := as.Dbrepo.CreateUser(ctx, tx, user)
	if !bdresponse.Success && bdresponse.Errors != nil {
		if bdresponse.Errors.Type == erro.ServerErrorType {
			registrateMap[erro.ServerErrorType] = bdresponse.Errors.Message
			metrics.UserErrorsTotal.WithLabelValues(erro.ServerErrorType).Inc()
			return &ServiceResponse{Success: false, Errors: registrateMap, ErrorType: erro.ServerErrorType}
		}
		registrateMap[erro.ClientErrorType] = bdresponse.Errors.Message
		metrics.UserErrorsTotal.WithLabelValues(erro.ClientErrorType).Inc()
		return &ServiceResponse{Success: false, Errors: registrateMap, ErrorType: erro.ClientErrorType}
	}
	userID = bdresponse.Data["userID"].(uuid.UUID)
	grpcresponse, serviceresponse := retryOperationGrpc(ctx, func(ctx context.Context) (*proto.CreateSessionResponse, error) {
		return as.GrpcClient.CreateSession(ctx, userID.String())
	}, traceid, registrateMap, place, as.KafkaProducer)
	if serviceresponse != nil {
		return serviceresponse
	}
	err = commitTransaction(as.Dbtxmanager, tx, traceid, place, as.KafkaProducer)
	if err != nil {
		_, serviceresponse := retryOperationGrpc(ctx, func(ctx context.Context) (*proto.DeleteSessionResponse, error) {
			return as.GrpcClient.DeleteSession(ctx, grpcresponse.SessionID)
		}, traceid, registrateMap, place, as.KafkaProducer)
		if serviceresponse != nil {
			as.KafkaProducer.NewUserLog(kafka.LogLevelError, place, traceid, "Failed to delete session after transaction failure")
		}
		registrateMap[erro.ServerErrorType] = erro.UserServiceUnavalaible
		return &ServiceResponse{Success: false, Errors: registrateMap, ErrorType: erro.ServerErrorType}
	}
	isTransactionActive = false
	return &ServiceResponse{Success: true, Data: map[string]any{"userID": userID, "sessionID": grpcresponse.SessionID, "expiresession": time.Unix(grpcresponse.ExpiryTime, 0)}}
}
func (as *AuthService) AuthenticateAndLogin(ctx context.Context, user *model.Person) *ServiceResponse {
	const place = AuthenticateAndLogin
	authenticateMap := make(map[string]string)
	traceid := ctx.Value("traceID").(string)
	errorvalidate := validatePerson(as.Validator, user, false, traceid, place, as.KafkaProducer)
	if errorvalidate != nil {
		return &ServiceResponse{Success: false, Errors: errorvalidate, ErrorType: erro.ClientErrorType}
	}
	bdresponse := as.Dbrepo.AuthenticateUser(ctx, user.Email, user.Password)
	if !bdresponse.Success && bdresponse.Errors != nil {
		if bdresponse.Errors.Type == erro.ServerErrorType {
			authenticateMap[erro.ServerErrorType] = bdresponse.Errors.Message
			metrics.UserErrorsTotal.WithLabelValues(erro.ServerErrorType).Inc()
			return &ServiceResponse{Success: false, Errors: authenticateMap, ErrorType: erro.ServerErrorType}
		}
		authenticateMap[erro.ClientErrorType] = bdresponse.Errors.Message
		metrics.UserErrorsTotal.WithLabelValues(erro.ClientErrorType).Inc()
		return &ServiceResponse{Success: false, Errors: authenticateMap, ErrorType: erro.ClientErrorType}
	}
	userID := bdresponse.Data["userID"].(uuid.UUID)
	grpcresponse, serviceresponse := retryOperationGrpc(ctx, func(ctx context.Context) (*proto.CreateSessionResponse, error) {
		return as.GrpcClient.CreateSession(ctx, userID.String())
	}, traceid, authenticateMap, place, as.KafkaProducer)
	if serviceresponse != nil {
		return serviceresponse
	}
	as.KafkaProducer.NewUserLog(kafka.LogLevelInfo, place, traceid, "The session was created successfully and received")
	return &ServiceResponse{Success: true, Data: map[string]any{"userID": userID, "sessionID": grpcresponse.SessionID, "expiresession": time.Unix(grpcresponse.ExpiryTime, 0)}}
}
func (as *AuthService) DeleteAccount(ctx context.Context, sessionID string, useridstr string, password string) *ServiceResponse {
	const place = DeleteAccount
	deletemap := make(map[string]string)
	traceid := ctx.Value("traceID").(string)
	userid, err := uuid.Parse(useridstr)
	if err != nil {
		fmterr := fmt.Sprintf("UUID-parse Error: %v", err)
		as.KafkaProducer.NewUserLog(kafka.LogLevelError, place, traceid, fmterr)
		deletemap[erro.ServerErrorType] = erro.UserServiceUnavalaible
		metrics.UserErrorsTotal.WithLabelValues(erro.ServerErrorType).Inc()
		return &ServiceResponse{Success: false, Errors: deletemap, ErrorType: erro.ServerErrorType}
	}
	var tx *sql.Tx
	tx, err = as.Dbtxmanager.BeginTx(ctx)
	metrics.UserDBQueriesTotal.WithLabelValues("Begin Transaction").Inc()
	if err != nil {
		fmterr := fmt.Sprintf("Transaction Error: %v", err)
		as.KafkaProducer.NewUserLog(kafka.LogLevelError, place, traceid, fmterr)
		deletemap[erro.ServerErrorType] = erro.UserServiceUnavalaible
		metrics.UserDBErrorsTotal.WithLabelValues("Begin Transaction", "Transaction").Inc()
		metrics.UserErrorsTotal.WithLabelValues(erro.ServerErrorType).Inc()
		return &ServiceResponse{Success: false, Errors: deletemap, ErrorType: erro.ServerErrorType}
	}
	isTransactionActive := true
	defer func() {
		if isTransactionActive {
			rollbackTransaction(as.Dbtxmanager, tx, traceid, place, as.KafkaProducer)
			isTransactionActive = false
		} else {
			as.KafkaProducer.NewUserLog(kafka.LogLevelInfo, place, traceid, "Transaction was successfully committed and user has successfully deleted his account with all data")
		}
	}()
	bdresponse := as.Dbrepo.DeleteUser(ctx, tx, userid, password)
	if !bdresponse.Success && bdresponse.Errors != nil {
		if bdresponse.Errors.Type == erro.ServerErrorType {
			deletemap[erro.ServerErrorType] = bdresponse.Errors.Message
			metrics.UserErrorsTotal.WithLabelValues(erro.ServerErrorType).Inc()
			return &ServiceResponse{Success: false, Errors: deletemap, ErrorType: erro.ServerErrorType}
		}
		deletemap[erro.ClientErrorType] = bdresponse.Errors.Message
		metrics.UserErrorsTotal.WithLabelValues(erro.ClientErrorType).Inc()
		return &ServiceResponse{Success: false, Errors: deletemap, ErrorType: erro.ClientErrorType}
	}
	grpcresponse, serviceresponse := retryOperationGrpc(ctx, func(ctx context.Context) (*proto.DeleteSessionResponse, error) {
		return as.GrpcClient.DeleteSession(ctx, sessionID)
	}, traceid, deletemap, place, as.KafkaProducer)
	if serviceresponse != nil {
		return serviceresponse
	}
	err = commitTransaction(as.Dbtxmanager, tx, traceid, place, as.KafkaProducer)
	if err != nil {
		deletemap[erro.ServerErrorType] = erro.UserServiceUnavalaible
		return &ServiceResponse{Success: false, Errors: deletemap, ErrorType: erro.ServerErrorType}
	}
	isTransactionActive = false
	return &ServiceResponse{Success: grpcresponse.Success}
}
func (as *AuthService) Logout(ctx context.Context, sessionID string) *ServiceResponse {
	const place = Logout
	logMap := make(map[string]string)
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
func (as *AuthService) UpdateAccount(ctx context.Context, data map[string]string, useridstr string, updateType string) *ServiceResponse {
	const place = UpdateAccount
	traceid := ctx.Value("traceID").(string)
	updatemap := make(map[string]string)
	userid, err := uuid.Parse(useridstr)
	if err != nil {
		fmterr := fmt.Sprintf("UUID-parse Error: %v", err)
		as.KafkaProducer.NewUserLog(kafka.LogLevelError, place, traceid, fmterr)
		updatemap[erro.ServerErrorType] = erro.UserServiceUnavalaible
		metrics.UserErrorsTotal.WithLabelValues(erro.ServerErrorType).Inc()
		return &ServiceResponse{Success: false, Errors: updatemap, ErrorType: erro.ServerErrorType}
	}
	name, name_ok := data["name"]
	email, email_ok := data["email"]
	last_password, last_password_ok := data["last_password"]
	new_password, new_password_ok := data["new_password"]
	if name_ok && !last_password_ok && !new_password_ok && !email_ok && updateType == "name" {
		response := as.Dbrepo.UpdateUserName(ctx, userid, name)
		return &ServiceResponse{Success: response.Success}
	}
	if !name_ok && last_password_ok && new_password_ok && !email_ok && updateType == "password" {
		response := as.Dbrepo.UpdateUserPassword(ctx, userid, last_password, new_password)
		return &ServiceResponse{Success: response.Success}
	}
	if !name_ok && last_password_ok && !new_password_ok && email_ok && updateType == "email" {
		response := as.Dbrepo.UpdateUserEmail(ctx, userid, email, last_password)
		return &ServiceResponse{Success: response.Success}
	}
	fmterr := "Invalid data in request"
	updatemap[erro.ClientErrorType] = fmterr
	as.KafkaProducer.NewUserLog(kafka.LogLevelWarn, UpdateAccount, traceid, fmterr)
	metrics.UserErrorsTotal.WithLabelValues("ClientError").Inc()
	return &ServiceResponse{Success: false, Errors: updatemap, ErrorType: erro.ClientErrorType}
}
