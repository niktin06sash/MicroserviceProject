package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/erro"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/model"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/repository"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func validatePerson(val *validator.Validate, user *model.Person, flag bool, traceid string, place string) map[string]error {
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
					log.Printf("[INFO] [UserManagement] [TraceID: %s] %s: Email format error", traceid, place)
					erors["ClientError"] = erro.ErrorNotEmail
				case "min":
					errv := fmt.Errorf("%s is too short", err.Field())
					log.Printf("[INFO] [UserManagement] [TraceID: %s] %s: %s format error", traceid, place, errv)
					erors["ClientError"] = errv
				default:
					errv := fmt.Errorf("%s is Null", err.Field())
					log.Printf("[INFO] [UserManagement] [TraceID: %s] %s: %s format error", traceid, place, errv)
					erors["ClientError"] = errv
				}
			}
			return erors
		}
	}
	return nil
}
func rollbackTransaction(txMgr repository.DBTransactionManager, tx *sql.Tx, traceid string, place string) {
	if tx == nil {
		return
	}
	maxAttempts := 3
	attempt := 0
	for attempt < maxAttempts {
		attempt++
		err := txMgr.RollbackTx(tx)
		if err == nil {
			log.Printf("[INFO] [UserManagement] [TraceID: %s] %s: Successful rollback on attempt %d", traceid, place, attempt)
			return
		}
		if errors.Is(err, sql.ErrTxDone) {
			log.Printf("[INFO] [UserManagement] [TraceID: %s] %s: Transaction is already completed, skipping rollback", traceid, place)
			return
		}
		log.Printf("[ERROR] [UserManagement] [TraceID: %s] %s: Error rolling back transaction on attempt %d: %v", traceid, place, attempt, err)
		if attempt == maxAttempts {
			log.Printf("[ERROR] [UserManagement] [TraceID: %s] %s: Failed to rollback transaction after %d attempts", traceid, place, maxAttempts)
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
func retryOperationGrpc(ctx context.Context, operation func(context.Context) (interface{}, error), traceID string, errorMap map[string]error, place string) (interface{}, *ServiceResponse) {
	var response interface{}
	var err error
	for i := 1; i <= 3; i++ {
		if ctxresponse, shouldReturn := checkContext(ctx, errorMap); shouldReturn {
			log.Printf("[ERROR] [UserManagement] [TraceID: %s] %s: Context cancelled before operation: %v", traceID, place, ctx.Err())
			return nil, ctxresponse
		}
		response, err = operation(ctx)
		if err != nil {
			st, _ := status.FromError(err)
			log.Printf("[ERROR] [UserManagement] [TraceID: %s] %s: Operation attempt %d failed: %v", traceID, place, i, st.Message())
			switch st.Code() {
			case codes.Internal:
				time.Sleep(time.Duration(i) * time.Second)
				continue
			case codes.Unavailable:
				log.Printf("[WARN] [UserManagement] [TraceID: %s] %s: Server unavailable, retrying (%d)...", traceID, place, i)
				time.Sleep(time.Duration(i) * time.Second)
				continue
			case codes.Canceled:
				log.Printf("[WARN] [UserManagement] [TraceID: %s] %s: Server unavailable, retrying (%d)...", traceID, place, i)
				continue
			default:
				errorMap["ClientError"] = err
				return nil, &ServiceResponse{
					Success: false,
					Errors:  errorMap,
					Type:    erro.ClientErrorType,
				}
			}
		} else {
			return response, nil
		}
	}
	log.Printf("[ERROR] [UserManagement] [TraceID: %s] %s: Operation failed after all attempts", traceID, place)
	errorMap["InternalServerError"] = erro.ErrorAllRetryFailed
	return nil, &ServiceResponse{
		Success: false,
		Errors:  errorMap,
		Type:    erro.ServerErrorType,
	}
}
func retryOperationDB(ctx context.Context, operation func(context.Context) *repository.DBRepositoryResponse, traceID string, errorMap map[string]error, place string) (*repository.DBRepositoryResponse, *ServiceResponse) {
	var response *repository.DBRepositoryResponse
	for i := 1; i <= 3; i++ {
		if ctxresponse, shouldReturn := checkContext(ctx, errorMap); shouldReturn {
			log.Printf("[ERROR] [UserManagement] [TraceID: %s] %s: Context cancelled before operation: %v", traceID, place, ctx.Err())
			return nil, ctxresponse
		}
		response = operation(ctx)
		if !response.Success && response.Errors != nil {
			log.Printf("[ERROR] [UserManagement] [TraceID: %s] %s: Operation attempt %d failed: %v", traceID, place, i, response.Errors)
			if response.Type == erro.ServerErrorType {
				time.Sleep(time.Duration(i) * time.Second)
				continue
			}
			errorMap["ClientError"] = response.Errors
			return nil, &ServiceResponse{Success: response.Success, Errors: errorMap, Type: response.Type}
		} else {
			return response, nil
		}
	}
	log.Printf("[ERROR] [UserManagement] [TraceID: %s] %s: Operation failed after all attempts", traceID, place)
	errorMap["InternalServerError"] = erro.ErrorAllRetryFailed
	return nil, &ServiceResponse{
		Success: false,
		Errors:  errorMap,
		Type:    erro.ServerErrorType,
	}
}
