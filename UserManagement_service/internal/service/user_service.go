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
	Dbrepo        repository.DBUserRepos
	Dbtxmanager   repository.DBTransactionManager
	KafkaProducer kafka.KafkaProducerService
	Validator     *validator.Validate
	GrpcClient    client.GrpcClientService
}

func NewAuthService(dbrepo repository.DBUserRepos, dbtxmanager repository.DBTransactionManager, kafkaProd kafka.KafkaProducerService, grpc *client.GrpcClient) *AuthService {
	validator := validator.New()
	return &AuthService{Dbrepo: dbrepo, Dbtxmanager: dbtxmanager, Validator: validator, KafkaProducer: kafkaProd, GrpcClient: grpc}
}
func (as *AuthService) RegistrateAndLogin(ctx context.Context, req *model.RegistrationRequest) *ServiceResponse {
	const place = RegistrateAndLogin
	registrateMap := make(map[string]string)
	traceid := ctx.Value("traceID").(string)
	errorvalidate := validateData(as.Validator, req, traceid, place, as.KafkaProducer)
	if errorvalidate != nil {
		return &ServiceResponse{Success: false, Errors: errorvalidate, ErrorType: erro.ClientErrorType}
	}
	hashpass, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		fmterr := fmt.Sprintf("HashPass Error: %v", err)
		as.KafkaProducer.NewUserLog(kafka.LogLevelError, place, traceid, fmterr)
		registrateMap[erro.ServerErrorType] = erro.UserServiceUnavalaible
		metrics.UserErrorsTotal.WithLabelValues(erro.ServerErrorType).Inc()
		return &ServiceResponse{Success: false, Errors: registrateMap, ErrorType: erro.ServerErrorType}
	}
	user := &model.User{Id: uuid.New(), Password: string(hashpass), Email: req.Email, Name: req.Name}
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
	bdresponse, serviceresponse := requestToDB(as.Dbrepo.CreateUser(ctx, tx, user), registrateMap)
	if serviceresponse != nil {
		return serviceresponse
	}
	userID := bdresponse.Data["userID"].(uuid.UUID)
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
func (as *AuthService) AuthenticateAndLogin(ctx context.Context, req *model.AuthenticationRequest) *ServiceResponse {
	const place = AuthenticateAndLogin
	authenticateMap := make(map[string]string)
	traceid := ctx.Value("traceID").(string)
	errorvalidate := validateData(as.Validator, req, traceid, place, as.KafkaProducer)
	if errorvalidate != nil {
		return &ServiceResponse{Success: false, Errors: errorvalidate, ErrorType: erro.ClientErrorType}
	}
	bdresponse, serviceresponse := requestToDB(as.Dbrepo.AuthenticateUser(ctx, req.Email, req.Password), authenticateMap)
	if serviceresponse != nil {
		return serviceresponse
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
func (as *AuthService) DeleteAccount(ctx context.Context, req *model.DeletionRequest, sessionID string, useridstr string) *ServiceResponse {
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
	errorvalidate := validateData(as.Validator, req, traceid, place, as.KafkaProducer)
	if errorvalidate != nil {
		return &ServiceResponse{Success: false, Errors: errorvalidate, ErrorType: erro.ClientErrorType}
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
	_, serviceresponse := requestToDB(as.Dbrepo.DeleteUser(ctx, tx, userid, req.Password), deletemap)
	if serviceresponse != nil {
		return serviceresponse
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
func (as *AuthService) UpdateAccount(ctx context.Context, req *model.UpdateRequest, useridstr string, updateType string) *ServiceResponse {
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
	errorvalidate := validateData(as.Validator, req, traceid, place, as.KafkaProducer)
	if errorvalidate != nil {
		return &ServiceResponse{Success: false, Errors: errorvalidate, ErrorType: erro.ClientErrorType}
	}
	if req.Name != "" && req.Email == "" && req.LastPassword == "" && req.NewPassword == "" && updateType == "name" {
		bdresponse, serviceresponse := requestToDB(as.Dbrepo.UpdateUserName(ctx, userid, req.Name), updatemap)
		if serviceresponse != nil {
			return serviceresponse
		}
		return &ServiceResponse{Success: bdresponse.Success}
	}
	if req.Name == "" && req.Email == "" && req.LastPassword != "" && req.NewPassword != "" && updateType == "password" {
		bdresponse, serviceresponse := requestToDB(as.Dbrepo.UpdateUserPassword(ctx, userid, req.LastPassword, req.NewPassword), updatemap)
		if serviceresponse != nil {
			return serviceresponse
		}
		return &ServiceResponse{Success: bdresponse.Success}
	}
	if req.Name == "" && req.Email != "" && req.LastPassword != "" && req.NewPassword == "" && updateType == "email" {
		bdresponse, serviceresponse := requestToDB(as.Dbrepo.UpdateUserEmail(ctx, userid, req.Email, req.LastPassword), updatemap)
		if serviceresponse != nil {
			return serviceresponse
		}
		return &ServiceResponse{Success: bdresponse.Success}
	}
	fmterr := "Invalid data in request"
	updatemap[erro.ClientErrorType] = fmterr
	as.KafkaProducer.NewUserLog(kafka.LogLevelWarn, UpdateAccount, traceid, fmterr)
	metrics.UserErrorsTotal.WithLabelValues("ClientError").Inc()
	return &ServiceResponse{Success: false, Errors: updatemap, ErrorType: erro.ClientErrorType}
}
func (as *AuthService) GetMyProfile(ctx context.Context, useridstr string) *ServiceResponse {
	const place = GetMyProfile
	getMyProfileMap := make(map[string]string)
	traceid := ctx.Value("traceID").(string)
	userid, err := uuid.Parse(useridstr)
	if err != nil {
		fmterr := fmt.Sprintf("UUID-parse Error: %v", err)
		as.KafkaProducer.NewUserLog(kafka.LogLevelError, place, traceid, fmterr)
		getMyProfileMap[erro.ServerErrorType] = erro.UserServiceUnavalaible
		metrics.UserErrorsTotal.WithLabelValues(erro.ServerErrorType).Inc()
		return &ServiceResponse{Success: false, Errors: getMyProfileMap, ErrorType: erro.ServerErrorType}
	}
	bdresponse, serviceresponse := requestToDB(as.Dbrepo.GetMyProfile(ctx, userid), getMyProfileMap)
	if serviceresponse != nil {
		return serviceresponse
	}
	return &ServiceResponse{Success: bdresponse.Success, Data: bdresponse.Data}
}
func (as *AuthService) GetProfileById(ctx context.Context, useridstr string, getidstr string) *ServiceResponse {
	const place = GetProfileById
	getProfilebyIdMap := make(map[string]string)
	traceid := ctx.Value("traceID").(string)
	userid, err := uuid.Parse(useridstr)
	if err != nil {
		fmterr := fmt.Sprintf("UUID-parse Error: %v", err)
		as.KafkaProducer.NewUserLog(kafka.LogLevelError, place, traceid, fmterr)
		getProfilebyIdMap[erro.ServerErrorType] = erro.UserServiceUnavalaible
		metrics.UserErrorsTotal.WithLabelValues(erro.ServerErrorType).Inc()
		return &ServiceResponse{Success: false, Errors: getProfilebyIdMap, ErrorType: erro.ServerErrorType}
	}
	getid, err := uuid.Parse(getidstr)
	if err != nil {
		fmterr := fmt.Sprintf("UUID-parse Error: %v", err)
		as.KafkaProducer.NewUserLog(kafka.LogLevelError, place, traceid, fmterr)
		getProfilebyIdMap[erro.ServerErrorType] = erro.UserServiceUnavalaible
		metrics.UserErrorsTotal.WithLabelValues(erro.ServerErrorType).Inc()
		return &ServiceResponse{Success: false, Errors: getProfilebyIdMap, ErrorType: erro.ServerErrorType}
	}
	if userid == getid {
		as.KafkaProducer.NewUserLog(kafka.LogLevelWarn, place, traceid, "Person attempted to search for their own profile by ID")
		getProfilebyIdMap[erro.ClientErrorType] = "You cannot search for your own profile by ID"
		metrics.UserErrorsTotal.WithLabelValues(erro.ClientErrorType).Inc()
		return &ServiceResponse{Success: false, Errors: getProfilebyIdMap, ErrorType: erro.ClientErrorType}
	}
	bdresponse, serviceresponse := requestToDB(as.Dbrepo.GetProfileById(ctx, getid), getProfilebyIdMap)
	if serviceresponse != nil {
		return serviceresponse
	}
	return &ServiceResponse{Success: bdresponse.Success, Data: bdresponse.Data}
}
