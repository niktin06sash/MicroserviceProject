package service

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/niktin06sash/MicroserviceProject/SessionManagement_service/proto"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/client"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/erro"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/kafka"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/model"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/repository"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

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
	errorvalidate := validatePerson(as.Validator, user, true, traceid)
	if errorvalidate != nil {
		log.Printf("[ERROR] [UserManagement] [TraceID: %s]: RegistrateAndLogin: Validate error %v", traceid, errorvalidate)
		return &ServiceResponse{Success: false, Errors: errorvalidate, Type: erro.ClientErrorType}
	}
	var tx *sql.Tx
	tx, err := as.Dbtxmanager.BeginTx(ctx)
	if err != nil {
		log.Printf("[ERROR] [UserManagement] [TraceID: %s]: RegistrateAndLogin: TransactionError %v", traceid, err)
		registrateMap["InternalServerError"] = erro.ErrorStartTransaction
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
		registrateMap["InternalServerError"] = erro.ErrorHashPass
		return &ServiceResponse{Success: false, Errors: registrateMap, Type: erro.ServerErrorType}
	}
	user.Password = string(hashpass)

	userID := uuid.New()
	user.Id = userID
	bdresponse, serviceresponse := retryOperationBD(ctx, func(ctx context.Context) *repository.DBRepositoryResponse {
		return as.Dbrepo.CreateUser(ctx, tx, user)
	}, traceid, registrateMap)
	if serviceresponse != nil {
		return serviceresponse
	}
	userID = bdresponse.UserId
	md := metadata.Pairs("traceID", traceid)
	ctx = metadata.NewOutgoingContext(ctx, md)
	grpcresponse, serviceresponse := retryOperationGrpc(ctx, func(ctx context.Context) (interface{}, error) {
		return as.GrpcClient.CreateSession(ctx, userID.String())
	}, traceid, registrateMap)
	if serviceresponse != nil {
		return serviceresponse
	}
	sessionresponse := grpcresponse.(*proto.CreateSessionResponse)
	if err := as.Dbtxmanager.CommitTx(tx); err != nil {
		log.Printf("[ERROR] [UserManagement] [TraceID: %s]: RegistrateAndLogin: Error committing transaction: %v", traceid, err)
		_, serviceresponse := retryOperationGrpc(ctx, func(ctx context.Context) (interface{}, error) {
			return as.GrpcClient.DeleteSession(ctx, sessionresponse.SessionID)
		}, traceid, registrateMap)
		if serviceresponse != nil {
			log.Printf("[ERROR] [UserManagement] [TraceID: %s]: RegistrateAndLogin: Failed to delete session after transaction failure: %v", traceid, serviceresponse.Errors)
		}
		registrateMap["InternalServerError"] = erro.ErrorCommitTransaction
		return &ServiceResponse{Success: false, Errors: registrateMap, Type: erro.ServerErrorType}
	}
	isTransactionActive = false
	timeExpire := time.Unix(sessionresponse.ExpiryTime, 0)
	return &ServiceResponse{Success: true, UserId: bdresponse.UserId, SessionId: sessionresponse.SessionID, ExpireSession: timeExpire}
}
func (as *AuthService) AuthenticateAndLogin(ctx context.Context, user *model.Person) *ServiceResponse {
	authenticateMap := make(map[string]error)
	traceid := ctx.Value("traceID").(string)
	errorvalidate := validatePerson(as.Validator, user, false, traceid)
	if errorvalidate != nil {
		log.Printf("[ERROR] [UserManagement] [TraceID: %s]: AuthenticateAndLogin: Validate error %v", traceid, errorvalidate)
		return &ServiceResponse{Success: false, Errors: errorvalidate, Type: erro.ClientErrorType}
	}
	bdresponse, serviceresponse := retryOperationBD(ctx, func(ctx context.Context) *repository.DBRepositoryResponse {
		return as.Dbrepo.GetUser(ctx, user.Email, user.Password)
	}, traceid, authenticateMap)
	if serviceresponse != nil {
		return serviceresponse
	}
	userID := bdresponse.UserId
	md := metadata.Pairs("traceID", traceid)
	ctx = metadata.NewOutgoingContext(ctx, md)
	grpcresponse, serviceresponse := retryOperationGrpc(ctx, func(ctx context.Context) (interface{}, error) {
		return as.GrpcClient.CreateSession(ctx, userID.String())
	}, traceid, authenticateMap)
	if serviceresponse != nil {
		return serviceresponse
	}
	sessionresponse := grpcresponse.(*proto.CreateSessionResponse)
	timeExpire := time.Unix(sessionresponse.ExpiryTime, 0)
	log.Printf("[INFO] [UserManagement] [TraceID: %s]: AuthenticateAndLogin: The session was created successfully and the user is authenticated!", traceid)
	return &ServiceResponse{Success: true, UserId: bdresponse.UserId, SessionId: sessionresponse.SessionID, ExpireSession: timeExpire}
}
func (as *AuthService) DeleteAccount(ctx context.Context, sessionID string, userid uuid.UUID, password string) *ServiceResponse {
	deletemap := make(map[string]error)
	traceid := ctx.Value("traceID").(string)
	var tx *sql.Tx
	tx, err := as.Dbtxmanager.BeginTx(ctx)
	if err != nil {
		log.Printf("[ERROR] [UserManagement] [TraceID: %s]: DeleteAccount: TransactionError %v", traceid, err)
		deletemap["InternalServerError"] = erro.ErrorStartTransaction
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
	_, serviceresponse := retryOperationBD(ctx, func(ctx context.Context) *repository.DBRepositoryResponse {
		return as.Dbrepo.DeleteUser(ctx, tx, userid, password)
	}, traceid, deletemap)
	if serviceresponse != nil {
		return serviceresponse
	}
	md := metadata.Pairs("traceID", traceid)
	ctx = metadata.NewOutgoingContext(ctx, md)
	grpcresponse, serviceresponse := retryOperationGrpc(ctx, func(ctx context.Context) (interface{}, error) {
		return as.GrpcClient.DeleteSession(ctx, sessionID)
	}, traceid, deletemap)
	if serviceresponse != nil {
		return serviceresponse
	}
	sessionresponse := grpcresponse.(*proto.DeleteSessionResponse)
	isTransactionActive = false
	return &ServiceResponse{
		Success: sessionresponse.Success,
	}
}
func validatePerson(val *validator.Validate, user *model.Person, flag bool, traceid string) map[string]error {
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
					log.Printf("[INFO] [UserManagement] [TraceID: %s]: Email format error", traceid)
					erors["ClientError"] = erro.ErrorNotEmail
				case "min":
					errv := fmt.Errorf("%s is too short", err.Field())
					log.Printf("[INFO] [UserManagement] [TraceID: %s]: %s format error", traceid, errv)
					erors["ClientError"] = errv
				default:
					errv := fmt.Errorf("%s is Null", err.Field())
					log.Printf("[INFO] [UserManagement] [TraceID: %s]: %s format error", traceid, errv)
					erors["ClientError"] = errv
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
func checkContext(ctx context.Context, mapa map[string]error) (*ServiceResponse, bool) {
	select {
	case <-ctx.Done():
		mapa["InternalServerError"] = erro.ErrorContextTimeout
		return &ServiceResponse{Success: false, Errors: mapa, Type: erro.ServerErrorType}, true
	default:
		return nil, false
	}
}
func retryOperationGrpc(ctx context.Context, operation func(context.Context) (interface{}, error), traceID string, errorMap map[string]error) (interface{}, *ServiceResponse) {
	var response interface{}
	var err error
	for i := 1; i <= 5; i++ {
		if ctxresponse, shouldReturn := checkContext(ctx, errorMap); shouldReturn {
			log.Printf("[ERROR] [UserManagement] [TraceID: %s]: Context cancelled before operation: %v", traceID, ctx.Err())
			return nil, ctxresponse
		}
		response, err = operation(ctx)
		if err != nil {
			st, _ := status.FromError(err)
			log.Printf("[ERROR] [UserManagement] [TraceID: %s]: Operation attempt %d failed: %v", traceID, i, st.Message())
			switch st.Code() {
			case codes.Internal:
				time.Sleep(time.Duration(i) * time.Second)
				continue
			default:
				errorMap["ClientError"] = err
				return nil, &ServiceResponse{
					Success: false,
					Errors:  errorMap,
					Type:    erro.ClientErrorType,
				}
			}
		}
	}
	if response == nil {
		log.Printf("[ERROR] [UserManagement] [TraceID: %s]: Operation failed after all attempts", traceID)
		errorMap["InternalServerError"] = erro.ErrorAllRetryFailed
		return nil, &ServiceResponse{Success: false, Errors: errorMap, Type: erro.ServerErrorType}
	}
	return response, nil
}
func retryOperationBD(ctx context.Context, operation func(context.Context) *repository.DBRepositoryResponse, traceID string, errorMap map[string]error) (*repository.DBRepositoryResponse, *ServiceResponse) {
	var response *repository.DBRepositoryResponse
	for i := 1; i <= 5; i++ {
		if ctxresponse, shouldReturn := checkContext(ctx, errorMap); shouldReturn {
			log.Printf("[ERROR] [UserManagement] [TraceID: %s]: Context cancelled before operation: %v", traceID, ctx.Err())
			return nil, ctxresponse
		}
		response = operation(ctx)
		if !response.Success && response.Errors != nil {
			log.Printf("[ERROR] [UserManagement] [TraceID: %s]: Operation attempt %d failed: %v", traceID, i, response.Errors)
			if response.Type == erro.ServerErrorType {
				time.Sleep(time.Duration(i) * time.Second)
				continue
			}
			errorMap["ClientError"] = response.Errors
			return nil, &ServiceResponse{Success: response.Success, Errors: errorMap, Type: response.Type}
		}
		break
	}
	if response == nil || response.UserId == uuid.Nil {
		log.Printf("[ERROR] [UserManagement] [TraceID: %s]: Operation failed after all attempts", traceID)
		errorMap["InternalServerError"] = erro.ErrorAllRetryFailed
		return nil, &ServiceResponse{Success: false, Errors: errorMap, Type: erro.ServerErrorType}
	}
	return response, nil
}
