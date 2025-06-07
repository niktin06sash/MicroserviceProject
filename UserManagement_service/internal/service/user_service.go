package service

import (
	"context"
	"database/sql"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/niktin06sash/MicroserviceProject/SessionManagement_service/proto"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/brokers/kafka"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/erro"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/metrics"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/model"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	Dbrepo        DBUserRepos
	Dbtxmanager   DBTransactionManager
	Redisrepo     CacheUserRepos
	RabbitEvents  EventProducer
	KafkaProducer LogProducer
	Validator     *validator.Validate
	GrpcClient    SessionClient
}

func NewUserService(dbrepo DBUserRepos, dbtxmanager DBTransactionManager, redisrepo CacheUserRepos, kafkaProd LogProducer, rabbitevents EventProducer, grpc SessionClient) *UserService {
	validator := validator.New()
	return &UserService{Dbrepo: dbrepo, Dbtxmanager: dbtxmanager, Redisrepo: redisrepo, Validator: validator, KafkaProducer: kafkaProd, GrpcClient: grpc}
}
func (as *UserService) RegistrateAndLogin(ctx context.Context, req *model.RegistrationRequest) *ServiceResponse {
	const place = RegistrateAndLogin
	registrateMap := make(map[string]string)
	traceid := ctx.Value("traceID").(string)
	errorvalidate := validateData(as.Validator, req, traceid, place, as.KafkaProducer)
	if errorvalidate != nil {
		return &ServiceResponse{Success: false, Errors: errorvalidate, ErrorType: erro.ClientErrorType}
	}
	hashpass, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		as.KafkaProducer.NewUserLog(kafka.LogLevelError, place, traceid, err.Error())
		registrateMap[erro.ErrorType] = erro.ServerErrorType
		registrateMap[erro.ErrorMessage] = erro.UserServiceUnavalaible
		metrics.UserErrorsTotal.WithLabelValues(erro.ServerErrorType).Inc()
		return &ServiceResponse{Success: false, Errors: registrateMap, ErrorType: erro.ServerErrorType}
	}
	user := &model.User{Id: uuid.New(), Password: string(hashpass), Email: req.Email, Name: req.Name}
	var tx *sql.Tx
	tx, resp := beginTransaction(ctx, as.Dbtxmanager, registrateMap, place, traceid, as.KafkaProducer)
	if resp != nil {
		return resp
	}
	bdresponse, serviceresponse := requestToDB(as.Dbrepo.CreateUser(ctx, tx, user), traceid, registrateMap, as.KafkaProducer)
	if serviceresponse != nil {
		rollbackTransaction(as.Dbtxmanager, tx, traceid, place, as.KafkaProducer)
		return serviceresponse
	}
	userID := bdresponse.Data[repository.KeyUserID].(string)
	grpcresponse, serviceresponse := retryOperationGrpc(ctx, func(ctx context.Context) (*proto.CreateSessionResponse, error) {
		return as.GrpcClient.CreateSession(ctx, userID)
	}, traceid, registrateMap, place, as.KafkaProducer)
	if serviceresponse != nil {
		rollbackTransaction(as.Dbtxmanager, tx, traceid, place, as.KafkaProducer)
		return serviceresponse
	}
	err = as.RabbitEvents.NewUserEvent(ctx, "user.registration", userID, place, traceid)
	if err != nil {
		rollbackTransaction(as.Dbtxmanager, tx, traceid, place, as.KafkaProducer)
		registrateMap[erro.ErrorType] = erro.ServerErrorType
		registrateMap[erro.ErrorMessage] = erro.UserServiceUnavalaible
		metrics.UserErrorsTotal.WithLabelValues(erro.ServerErrorType).Inc()
		return &ServiceResponse{Success: false, Errors: registrateMap, ErrorType: erro.ServerErrorType}
	}
	err = commitTransaction(as.Dbtxmanager, tx, registrateMap, traceid, place, as.KafkaProducer)
	if err != nil {
		rollbackTransaction(as.Dbtxmanager, tx, traceid, place, as.KafkaProducer)
		_, serviceresponse := retryOperationGrpc(ctx, func(ctx context.Context) (*proto.DeleteSessionResponse, error) {
			return as.GrpcClient.DeleteSession(ctx, grpcresponse.SessionID)
		}, traceid, registrateMap, place, as.KafkaProducer)
		if serviceresponse != nil {
			as.KafkaProducer.NewUserLog(kafka.LogLevelError, place, traceid, "Failed to delete session after transaction failure")
		}
		return &ServiceResponse{Success: false, Errors: registrateMap, ErrorType: erro.ServerErrorType}
	}
	as.KafkaProducer.NewUserLog(kafka.LogLevelInfo, place, traceid, "Transaction was successfully committed and session received")
	return &ServiceResponse{Success: true, Data: map[string]any{repository.KeyUserID: userID, KeyExpirySession: grpcresponse.SessionID, KeySessionID: time.Unix(grpcresponse.ExpiryTime, 0)}}
}
func (as *UserService) AuthenticateAndLogin(ctx context.Context, req *model.AuthenticationRequest) *ServiceResponse {
	const place = AuthenticateAndLogin
	authenticateMap := make(map[string]string)
	traceid := ctx.Value("traceID").(string)
	errorvalidate := validateData(as.Validator, req, traceid, place, as.KafkaProducer)
	if errorvalidate != nil {
		return &ServiceResponse{Success: false, Errors: errorvalidate, ErrorType: erro.ClientErrorType}
	}
	bdresponse, serviceresponse := requestToDB(as.Dbrepo.AuthenticateUser(ctx, req.Email, req.Password), traceid, authenticateMap, as.KafkaProducer)
	if serviceresponse != nil {
		return serviceresponse
	}
	userID := bdresponse.Data[repository.KeyUserID].(string)
	grpcresponse, serviceresponse := retryOperationGrpc(ctx, func(ctx context.Context) (*proto.CreateSessionResponse, error) {
		return as.GrpcClient.CreateSession(ctx, userID)
	}, traceid, authenticateMap, place, as.KafkaProducer)
	if serviceresponse != nil {
		return serviceresponse
	}
	as.KafkaProducer.NewUserLog(kafka.LogLevelInfo, place, traceid, "The session was created successfully and received")
	return &ServiceResponse{Success: true, Data: map[string]any{repository.KeyUserID: userID, KeySessionID: grpcresponse.SessionID, KeyExpirySession: time.Unix(grpcresponse.ExpiryTime, 0)}}
}
func (as *UserService) DeleteAccount(ctx context.Context, req *model.DeletionRequest, sessionID string, useridstr string) *ServiceResponse {
	const place = DeleteAccount
	deletemap := make(map[string]string)
	traceid := ctx.Value("traceID").(string)
	userid, err := parsingUserId(useridstr, deletemap, traceid, place, as.KafkaProducer)
	if err != nil {
		return &ServiceResponse{Success: false, Errors: deletemap, ErrorType: erro.ServerErrorType}
	}
	errorvalidate := validateData(as.Validator, req, traceid, place, as.KafkaProducer)
	if errorvalidate != nil {
		return &ServiceResponse{Success: false, Errors: errorvalidate, ErrorType: erro.ClientErrorType}
	}
	var tx *sql.Tx
	tx, resp := beginTransaction(ctx, as.Dbtxmanager, deletemap, place, traceid, as.KafkaProducer)
	if resp != nil {
		return resp
	}
	_, serviceresponse := requestToDB(as.Dbrepo.DeleteUser(ctx, tx, userid, req.Password), traceid, deletemap, as.KafkaProducer)
	if serviceresponse != nil {
		rollbackTransaction(as.Dbtxmanager, tx, traceid, place, as.KafkaProducer)
		return serviceresponse
	}
	_, serviceresponse = requestToDB(as.Redisrepo.DeleteProfileCache(ctx, useridstr), traceid, deletemap, as.KafkaProducer)
	if serviceresponse != nil {
		rollbackTransaction(as.Dbtxmanager, tx, traceid, place, as.KafkaProducer)
		return serviceresponse
	}
	grpcresponse, serviceresponse := retryOperationGrpc(ctx, func(ctx context.Context) (*proto.DeleteSessionResponse, error) {
		return as.GrpcClient.DeleteSession(ctx, sessionID)
	}, traceid, deletemap, place, as.KafkaProducer)
	if serviceresponse != nil {
		rollbackTransaction(as.Dbtxmanager, tx, traceid, place, as.KafkaProducer)
		return serviceresponse
	}
	err = as.RabbitEvents.NewUserEvent(ctx, "user.delete", useridstr, place, traceid)
	if err != nil {
		rollbackTransaction(as.Dbtxmanager, tx, traceid, place, as.KafkaProducer)
		deletemap[erro.ErrorType] = erro.ServerErrorType
		deletemap[erro.ErrorMessage] = erro.UserServiceUnavalaible
		metrics.UserErrorsTotal.WithLabelValues(erro.ServerErrorType).Inc()
		return &ServiceResponse{Success: false, Errors: deletemap, ErrorType: erro.ServerErrorType}
	}
	err = commitTransaction(as.Dbtxmanager, tx, deletemap, traceid, place, as.KafkaProducer)
	if err != nil {
		rollbackTransaction(as.Dbtxmanager, tx, traceid, place, as.KafkaProducer)
		return &ServiceResponse{Success: false, Errors: deletemap, ErrorType: erro.ServerErrorType}
	}
	as.KafkaProducer.NewUserLog(kafka.LogLevelInfo, place, traceid, "Transaction was successfully committed and user has successfully deleted his account with all data")
	return &ServiceResponse{Success: grpcresponse.Success}
}
func (as *UserService) Logout(ctx context.Context, sessionID string) *ServiceResponse {
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
func (as *UserService) UpdateAccount(ctx context.Context, req *model.UpdateRequest, useridstr string, updateType string) *ServiceResponse {
	const place = UpdateAccount
	traceid := ctx.Value("traceID").(string)
	updatemap := make(map[string]string)
	userid, err := parsingUserId(useridstr, updatemap, traceid, place, as.KafkaProducer)
	if err != nil {
		return &ServiceResponse{Success: false, Errors: updatemap, ErrorType: erro.ServerErrorType}
	}
	errorvalidate := validateData(as.Validator, req, traceid, place, as.KafkaProducer)
	if errorvalidate != nil {
		return &ServiceResponse{Success: false, Errors: errorvalidate, ErrorType: erro.ClientErrorType}
	}
	var tx *sql.Tx
	tx, resp := beginTransaction(ctx, as.Dbtxmanager, updatemap, place, traceid, as.KafkaProducer)
	if resp != nil {
		return resp
	}
	if req.Name != "" && req.Email == "" && req.LastPassword == "" && req.NewPassword == "" && updateType == "name" {
		return updateAndCommit(ctx, as.Dbtxmanager, as.Dbrepo.UpdateUserData, as.Redisrepo.DeleteProfileCache, tx, userid, updateType, updatemap, traceid, place, as.KafkaProducer, req.Name)
	}
	if req.Name == "" && req.Email == "" && req.LastPassword != "" && req.NewPassword != "" && updateType == "password" {
		return updateAndCommit(ctx, as.Dbtxmanager, as.Dbrepo.UpdateUserData, as.Redisrepo.DeleteProfileCache, tx, userid, updateType, updatemap, traceid, place, as.KafkaProducer, req.LastPassword, req.NewPassword)
	}
	if req.Name == "" && req.Email != "" && req.LastPassword != "" && req.NewPassword == "" && updateType == "email" {
		return updateAndCommit(ctx, as.Dbtxmanager, as.Dbrepo.UpdateUserData, as.Redisrepo.DeleteProfileCache, tx, userid, updateType, updatemap, traceid, place, as.KafkaProducer, req.Email, req.LastPassword)
	}
	updatemap[erro.ErrorType] = erro.ClientErrorType
	updatemap[erro.ErrorMessage] = erro.ErrorInvalidCountDinamicParameter
	as.KafkaProducer.NewUserLog(kafka.LogLevelWarn, UpdateAccount, traceid, erro.ErrorInvalidCountDinamicParameter)
	metrics.UserErrorsTotal.WithLabelValues(erro.ClientErrorType).Inc()
	rollbackTransaction(as.Dbtxmanager, tx, traceid, place, as.KafkaProducer)
	return &ServiceResponse{Success: false, Errors: updatemap, ErrorType: erro.ServerErrorType}
}
func (as *UserService) GetMyProfile(ctx context.Context, useridstr string) *ServiceResponse {
	const place = GetMyProfile
	getMyProfileMap := make(map[string]string)
	traceid := ctx.Value("traceID").(string)
	userid, err := parsingUserId(useridstr, getMyProfileMap, traceid, place, as.KafkaProducer)
	if err != nil {
		return &ServiceResponse{Success: false, Errors: getMyProfileMap, ErrorType: erro.ServerErrorType}
	}
	redisresponse, serviceresponse := requestToDB(as.Redisrepo.GetProfileCache(ctx, useridstr), traceid, getMyProfileMap, as.KafkaProducer)
	if serviceresponse != nil {
		return serviceresponse
	}
	if redisresponse.Success {
		return &ServiceResponse{Success: redisresponse.Success, Data: redisresponse.Data}
	}
	bdresponse, serviceresponse := requestToDB(as.Dbrepo.GetMyProfile(ctx, userid), traceid, getMyProfileMap, as.KafkaProducer)
	if serviceresponse != nil {
		return serviceresponse
	}
	_, serviceresponse = requestToDB(as.Redisrepo.AddProfileCache(ctx, useridstr, bdresponse.Data), traceid, getMyProfileMap, as.KafkaProducer)
	if serviceresponse != nil {
		return serviceresponse
	}
	return &ServiceResponse{Success: bdresponse.Success, Data: bdresponse.Data}
}
func (as *UserService) GetProfileById(ctx context.Context, useridstr string, getidstr string) *ServiceResponse {
	const place = GetProfileById
	getProfilebyIdMap := make(map[string]string)
	traceid := ctx.Value("traceID").(string)
	userid, err := parsingUserId(useridstr, getProfilebyIdMap, traceid, place, as.KafkaProducer)
	if err != nil {
		return &ServiceResponse{Success: false, Errors: getProfilebyIdMap, ErrorType: erro.ServerErrorType}
	}
	getid, err := parsingUserId(getidstr, getProfilebyIdMap, traceid, place, as.KafkaProducer)
	if err != nil {
		return &ServiceResponse{Success: false, Errors: getProfilebyIdMap, ErrorType: erro.ServerErrorType}
	}
	if userid == getid {
		as.KafkaProducer.NewUserLog(kafka.LogLevelWarn, place, traceid, "Person attempted to search for their own profile by ID")
		getProfilebyIdMap[erro.ErrorType] = erro.ClientErrorType
		getProfilebyIdMap[erro.ErrorMessage] = erro.ErrorSearchOwnProfile
		metrics.UserErrorsTotal.WithLabelValues(erro.ClientErrorType).Inc()
		return &ServiceResponse{Success: false, Errors: getProfilebyIdMap, ErrorType: erro.ClientErrorType}
	}
	redisresponse, serviceresponse := requestToDB(as.Redisrepo.GetProfileCache(ctx, getidstr), traceid, getProfilebyIdMap, as.KafkaProducer)
	if serviceresponse != nil {
		return serviceresponse
	}
	if redisresponse.Success {
		return &ServiceResponse{Success: redisresponse.Success, Data: redisresponse.Data}
	}
	bdresponse, serviceresponse := requestToDB(as.Dbrepo.GetProfileById(ctx, getid), traceid, getProfilebyIdMap, as.KafkaProducer)
	if serviceresponse != nil {
		return serviceresponse
	}
	_, serviceresponse = requestToDB(as.Redisrepo.AddProfileCache(ctx, getidstr, bdresponse.Data), traceid, getProfilebyIdMap, as.KafkaProducer)
	if serviceresponse != nil {
		return serviceresponse
	}
	return &ServiceResponse{Success: bdresponse.Success, Data: bdresponse.Data}
}
