package service

import (
	"context"
	"database/sql"
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

type UserServiceImplement struct {
	Dbrepo        repository.DBUserRepos
	Dbtxmanager   repository.DBTransactionManager
	Redisrepo     repository.CacheUserRepos
	KafkaProducer kafka.KafkaProducerService
	Validator     *validator.Validate
	GrpcClient    client.GrpcClientService
}

func NewUserServiceImplement(dbrepo repository.DBUserRepos, dbtxmanager repository.DBTransactionManager, redisrepo repository.CacheUserRepos, kafkaProd kafka.KafkaProducerService, grpc *client.GrpcClient) *UserServiceImplement {
	validator := validator.New()
	return &UserServiceImplement{Dbrepo: dbrepo, Dbtxmanager: dbtxmanager, Redisrepo: redisrepo, Validator: validator, KafkaProducer: kafkaProd, GrpcClient: grpc}
}
func (as *UserServiceImplement) RegistrateAndLogin(ctx context.Context, req *model.RegistrationRequest) *ServiceResponse {
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
	bdresponse, serviceresponse := requestToDB(as.Dbrepo.CreateUser(ctx, tx, user), registrateMap)
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
	return &ServiceResponse{Success: true, Data: map[string]any{"userID": userID, "sessionID": grpcresponse.SessionID, "expiresession": time.Unix(grpcresponse.ExpiryTime, 0)}}
}
func (as *UserServiceImplement) AuthenticateAndLogin(ctx context.Context, req *model.AuthenticationRequest) *ServiceResponse {
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
	userID := bdresponse.Data[repository.KeyUserID].(string)
	grpcresponse, serviceresponse := retryOperationGrpc(ctx, func(ctx context.Context) (*proto.CreateSessionResponse, error) {
		return as.GrpcClient.CreateSession(ctx, userID)
	}, traceid, authenticateMap, place, as.KafkaProducer)
	if serviceresponse != nil {
		return serviceresponse
	}
	as.KafkaProducer.NewUserLog(kafka.LogLevelInfo, place, traceid, "The session was created successfully and received")
	return &ServiceResponse{Success: true, Data: map[string]any{"userID": userID, "sessionID": grpcresponse.SessionID, "expiresession": time.Unix(grpcresponse.ExpiryTime, 0)}}
}
func (as *UserServiceImplement) DeleteAccount(ctx context.Context, req *model.DeletionRequest, sessionID string, useridstr string) *ServiceResponse {
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
	_, serviceresponse := requestToDB(as.Dbrepo.DeleteUser(ctx, tx, userid, req.Password), deletemap)
	if serviceresponse != nil {
		rollbackTransaction(as.Dbtxmanager, tx, traceid, place, as.KafkaProducer)
		return serviceresponse
	}
	_, serviceresponse = requestToDB(as.Redisrepo.DeleteProfileCache(ctx, useridstr), deletemap)
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
	err = commitTransaction(as.Dbtxmanager, tx, deletemap, traceid, place, as.KafkaProducer)
	if err != nil {
		rollbackTransaction(as.Dbtxmanager, tx, traceid, place, as.KafkaProducer)
		return &ServiceResponse{Success: false, Errors: deletemap, ErrorType: erro.ServerErrorType}
	}
	as.KafkaProducer.NewUserLog(kafka.LogLevelInfo, place, traceid, "Transaction was successfully committed and user has successfully deleted his account with all data")
	return &ServiceResponse{Success: grpcresponse.Success}
}
func (as *UserServiceImplement) Logout(ctx context.Context, sessionID string) *ServiceResponse {
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
func (as *UserServiceImplement) UpdateAccount(ctx context.Context, req *model.UpdateRequest, useridstr string, updateType string) *ServiceResponse {
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
	err = commitTransaction(as.Dbtxmanager, tx, updatemap, traceid, place, as.KafkaProducer)
	if err != nil {
		rollbackTransaction(as.Dbtxmanager, tx, traceid, place, as.KafkaProducer)
		return &ServiceResponse{Success: false, Errors: updatemap, ErrorType: erro.ServerErrorType}
	}
	return &ServiceResponse{Success: false, Errors: updatemap, ErrorType: erro.ClientErrorType}
}
func (as *UserServiceImplement) GetMyProfile(ctx context.Context, useridstr string) *ServiceResponse {
	const place = GetMyProfile
	getMyProfileMap := make(map[string]string)
	traceid := ctx.Value("traceID").(string)
	userid, err := parsingUserId(useridstr, getMyProfileMap, traceid, place, as.KafkaProducer)
	if err != nil {
		return &ServiceResponse{Success: false, Errors: getMyProfileMap, ErrorType: erro.ServerErrorType}
	}
	redisresponse, serviceresponse := requestToDB(as.Redisrepo.GetProfileCache(ctx, useridstr), getMyProfileMap)
	if serviceresponse != nil {
		return serviceresponse
	}
	if redisresponse.Success {
		return &ServiceResponse{Success: redisresponse.Success, Data: redisresponse.Data}
	}
	bdresponse, serviceresponse := requestToDB(as.Dbrepo.GetMyProfile(ctx, userid), getMyProfileMap)
	if serviceresponse != nil {
		return serviceresponse
	}
	_, serviceresponse = requestToDB(as.Redisrepo.AddProfileCache(ctx, useridstr, bdresponse.Data), getMyProfileMap)
	if serviceresponse != nil {
		return serviceresponse
	}
	return &ServiceResponse{Success: bdresponse.Success, Data: bdresponse.Data}
}
func (as *UserServiceImplement) GetProfileById(ctx context.Context, useridstr string, getidstr string) *ServiceResponse {
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
	redisresponse, serviceresponse := requestToDB(as.Redisrepo.GetProfileCache(ctx, getidstr), getProfilebyIdMap)
	if serviceresponse != nil {
		return serviceresponse
	}
	if redisresponse.Success {
		return &ServiceResponse{Success: redisresponse.Success, Data: redisresponse.Data}
	}
	bdresponse, serviceresponse := requestToDB(as.Dbrepo.GetProfileById(ctx, getid), getProfilebyIdMap)
	if serviceresponse != nil {
		return serviceresponse
	}
	_, serviceresponse = requestToDB(as.Redisrepo.AddProfileCache(ctx, getidstr, bdresponse.Data), getProfilebyIdMap)
	if serviceresponse != nil {
		return serviceresponse
	}
	return &ServiceResponse{Success: bdresponse.Success, Data: bdresponse.Data}
}
