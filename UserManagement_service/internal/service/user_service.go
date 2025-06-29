package service

import (
	"context"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/niktin06sash/MicroserviceProject/SessionManagement_service/proto"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/brokers/kafka"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/brokers/rabbitmq"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/erro"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/metrics"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/model"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	Dbrepo            DBUserRepos
	Dbtxmanager       DBTransactionManager
	CacheUserRepos    CacheUserRepos
	EventProducer     EventProducer
	LogProducer       LogProducer
	Validator         *validator.Validate
	GrpcSessionClient SessionClient
}

func NewUserService(dbrepo DBUserRepos, dbtxmanager DBTransactionManager, redisrepo CacheUserRepos, LogProducer LogProducer, events EventProducer, grpc SessionClient) *UserService {
	validator := validator.New()
	return &UserService{Dbrepo: dbrepo, Dbtxmanager: dbtxmanager, CacheUserRepos: redisrepo, Validator: validator, LogProducer: LogProducer, GrpcSessionClient: grpc, EventProducer: events}
}
func (as *UserService) RegistrateAndLogin(ctx context.Context, req *model.RegistrationRequest) *ServiceResponse {
	const place = RegistrateAndLogin
	traceid := ctx.Value("traceID").(string)
	errorvalidate := validateData(as.Validator, req, traceid, place, as.LogProducer)
	if errorvalidate != nil {
		return &ServiceResponse{Success: false, Errors: errorvalidate}
	}
	hashpass, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		as.LogProducer.NewUserLog(kafka.LogLevelError, place, traceid, err.Error())
		metrics.UserErrorsTotal.WithLabelValues(erro.ServerErrorType).Inc()
		return &ServiceResponse{Success: false, Errors: erro.ServerError(erro.UserServiceUnavalaible)}
	}
	user := &model.User{Id: uuid.New(), Password: string(hashpass), Email: req.Email, Name: req.Name}
	var tx pgx.Tx
	tx, resp := as.beginTransaction(ctx, place, traceid)
	if resp != nil {
		return resp
	}
	_, serviceresponse := as.requestToDB(as.Dbrepo.CreateUser(ctx, tx, user), traceid)
	if serviceresponse != nil {
		as.rollbackTransaction(ctx, tx, traceid, place)
		return serviceresponse
	}
	grpcresponse, serviceresponse := retryOperationGrpc(ctx, func(ctx context.Context) (*proto.CreateSessionResponse, error) {
		return as.GrpcSessionClient.CreateSession(ctx, user.Id.String())
	}, traceid, place, as.LogProducer)
	if serviceresponse != nil {
		as.rollbackTransaction(ctx, tx, traceid, place)
		return serviceresponse
	}
	err = as.EventProducer.NewUserEvent(ctx, rabbitmq.UserRegistrationKey, user.Id.String(), place, traceid)
	if err != nil {
		as.rollbackTransaction(ctx, tx, traceid, place)
		return &ServiceResponse{Success: false, Errors: erro.ServerError(erro.UserServiceUnavalaible)}
	}
	err = as.commitTransaction(ctx, tx, traceid, place)
	if err != nil {
		_, _ = retryOperationGrpc(ctx, func(ctx context.Context) (*proto.DeleteSessionResponse, error) {
			return as.GrpcSessionClient.DeleteSession(ctx, grpcresponse.SessionID)
		}, traceid, place, as.LogProducer)
		_ = as.EventProducer.NewUserEvent(ctx, rabbitmq.UserDeleteKey, user.Id.String(), place, traceid)
		as.rollbackTransaction(ctx, tx, traceid, place)
		return &ServiceResponse{Success: false, Errors: erro.ServerError(erro.UserServiceUnavalaible)}
	}
	as.LogProducer.NewUserLog(kafka.LogLevelInfo, place, traceid, "Transaction was successfully committed and session received")
	return &ServiceResponse{Success: true, Data: map[string]any{repository.KeyUserID: user.Id.String(), KeyExpirySession: time.Unix(grpcresponse.ExpiryTime, 0), KeySessionID: grpcresponse.SessionID}}
}
func (as *UserService) AuthenticateAndLogin(ctx context.Context, req *model.AuthenticationRequest) *ServiceResponse {
	const place = AuthenticateAndLogin
	traceid := ctx.Value("traceID").(string)
	errorvalidate := validateData(as.Validator, req, traceid, place, as.LogProducer)
	if errorvalidate != nil {
		return &ServiceResponse{Success: false, Errors: errorvalidate}
	}
	bdresponse, serviceresponse := as.requestToDB(as.Dbrepo.GetUser(ctx, req.Email, req.Password), traceid)
	if serviceresponse != nil {
		return serviceresponse
	}
	userID := bdresponse.Data[repository.KeyUserID].(string)
	grpcresponse, serviceresponse := retryOperationGrpc(ctx, func(ctx context.Context) (*proto.CreateSessionResponse, error) {
		return as.GrpcSessionClient.CreateSession(ctx, userID)
	}, traceid, place, as.LogProducer)
	if serviceresponse != nil {
		return serviceresponse
	}
	as.LogProducer.NewUserLog(kafka.LogLevelInfo, place, traceid, "The session was created successfully and received")
	return &ServiceResponse{Success: true, Data: map[string]any{repository.KeyUserID: bdresponse.Data[repository.KeyUserID], KeyExpirySession: time.Unix(grpcresponse.ExpiryTime, 0), KeySessionID: grpcresponse.SessionID}}
}
func (as *UserService) DeleteAccount(ctx context.Context, req *model.DeletionRequest, sessionID string, useridstr string) *ServiceResponse {
	const place = DeleteAccount
	traceid := ctx.Value("traceID").(string)
	userid, err := as.parsingUserId(useridstr, traceid, place)
	if err != nil {
		return &ServiceResponse{Success: false, Errors: erro.ClientError(erro.ErrorInvalidUserIDFormat)}
	}
	errorvalidate := validateData(as.Validator, req, traceid, place, as.LogProducer)
	if errorvalidate != nil {
		return &ServiceResponse{Success: false, Errors: errorvalidate}
	}
	var tx pgx.Tx
	tx, resp := as.beginTransaction(ctx, place, traceid)
	if resp != nil {
		return resp
	}
	_, serviceresponse := as.requestToDB(as.Dbrepo.DeleteUser(ctx, tx, userid, req.Password), traceid)
	if serviceresponse != nil {
		as.rollbackTransaction(ctx, tx, traceid, place)
		return serviceresponse
	}
	_, serviceresponse = as.requestToDB(as.CacheUserRepos.DeleteProfileCache(ctx, useridstr), traceid)
	if serviceresponse != nil {
		as.rollbackTransaction(ctx, tx, traceid, place)
		return serviceresponse
	}
	grpcresponse, serviceresponse := retryOperationGrpc(ctx, func(ctx context.Context) (*proto.DeleteSessionResponse, error) {
		return as.GrpcSessionClient.DeleteSession(ctx, sessionID)
	}, traceid, place, as.LogProducer)
	if serviceresponse != nil {
		as.rollbackTransaction(ctx, tx, traceid, place)
		return serviceresponse
	}
	err = as.EventProducer.NewUserEvent(ctx, rabbitmq.UserDeleteKey, useridstr, place, traceid)
	if err != nil {
		as.rollbackTransaction(ctx, tx, traceid, place)
		return &ServiceResponse{Success: false, Errors: erro.ServerError(erro.UserServiceUnavalaible)}
	}
	err = as.commitTransaction(ctx, tx, traceid, place)
	if err != nil {
		as.rollbackTransaction(ctx, tx, traceid, place)
		return &ServiceResponse{Success: false, Errors: erro.ServerError(erro.UserServiceUnavalaible)}
	}
	as.LogProducer.NewUserLog(kafka.LogLevelInfo, place, traceid, "Transaction was successfully committed and user has successfully deleted his account with all data")
	return &ServiceResponse{Success: grpcresponse.Success}
}
func (as *UserService) Logout(ctx context.Context, sessionID string) *ServiceResponse {
	const place = Logout
	traceid := ctx.Value("traceID").(string)
	grpcresponse, serviceresponse := retryOperationGrpc(ctx, func(ctx context.Context) (*proto.DeleteSessionResponse, error) {
		return as.GrpcSessionClient.DeleteSession(ctx, sessionID)
	}, traceid, place, as.LogProducer)
	if serviceresponse != nil {
		return serviceresponse
	}
	as.LogProducer.NewUserLog(kafka.LogLevelInfo, place, traceid, "The session was deleted successfully")
	return &ServiceResponse{Success: grpcresponse.Success}
}
func (as *UserService) UpdateAccount(ctx context.Context, req *model.UpdateRequest, useridstr string, updateType string) *ServiceResponse {
	const place = UpdateAccount
	traceid := ctx.Value("traceID").(string)
	userid, err := as.parsingUserId(useridstr, traceid, place)
	if err != nil {
		return &ServiceResponse{Success: false, Errors: erro.ClientError(erro.ErrorInvalidUserIDFormat)}
	}
	errorvalidate := validateData(as.Validator, req, traceid, place, as.LogProducer)
	if errorvalidate != nil {
		return &ServiceResponse{Success: false, Errors: errorvalidate}
	}
	var tx pgx.Tx
	tx, resp := as.beginTransaction(ctx, place, traceid)
	if resp != nil {
		return resp
	}
	if req.Name != "" && req.Email == "" && req.LastPassword == "" && req.NewPassword == "" && updateType == "name" {
		return as.updateAndCommit(ctx, tx, userid, updateType, traceid, place, req.Name)
	}
	if req.Name == "" && req.Email == "" && req.LastPassword != "" && req.NewPassword != "" && updateType == "password" {
		return as.updateAndCommit(ctx, tx, userid, updateType, traceid, place, req.LastPassword, req.NewPassword)
	}
	if req.Name == "" && req.Email != "" && req.LastPassword != "" && req.NewPassword == "" && updateType == "email" {
		return as.updateAndCommit(ctx, tx, userid, updateType, traceid, place, req.Email, req.LastPassword)
	}
	as.LogProducer.NewUserLog(kafka.LogLevelWarn, UpdateAccount, traceid, erro.ErrorInvalidCountDinamicParameter)
	metrics.UserErrorsTotal.WithLabelValues(erro.ClientErrorType).Inc()
	as.rollbackTransaction(ctx, tx, traceid, place)
	return &ServiceResponse{Success: false, Errors: erro.ClientError(erro.ErrorInvalidCountDinamicParameter)}
}
func (as *UserService) GetProfileById(ctx context.Context, getidstr string) *ServiceResponse {
	const place = GetProfileById
	traceid := ctx.Value("traceID").(string)
	getid, err := as.parsingUserId(getidstr, traceid, place)
	if err != nil {
		return &ServiceResponse{Success: false, Errors: erro.ClientError(erro.ErrorInvalidUserIDFormat)}
	}
	redisresponse, serviceresponse := as.requestToDB(as.CacheUserRepos.GetProfileCache(ctx, getidstr), traceid)
	if serviceresponse != nil {
		return serviceresponse
	}
	if redisresponse.Success {
		return &ServiceResponse{Success: redisresponse.Success, Data: redisresponse.Data}
	}
	bdresponse, serviceresponse := as.requestToDB(as.Dbrepo.GetProfileById(ctx, getid), traceid)
	if serviceresponse != nil {
		return serviceresponse
	}
	_, serviceresponse = as.requestToDB(as.CacheUserRepos.AddProfileCache(ctx, bdresponse.Data[repository.KeyUser].(*model.User)), traceid)
	if serviceresponse != nil {
		return serviceresponse
	}
	return &ServiceResponse{Success: bdresponse.Success, Data: bdresponse.Data}
}
