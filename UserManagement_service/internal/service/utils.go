package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/erro"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/kafka"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/model"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/repository"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func validatePerson(val *validator.Validate, user *model.Person, flag bool, traceid string, place string, kafkaProd kafka.KafkaProducerService) map[string]error {
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
					fmterr := fmt.Sprintf("Email format error: %v", err)
					kafkaProd.NewUserLog(kafka.LogLevelWarn, place, traceid, fmterr)
					erors[err.Field()] = erro.ErrorNotEmail
				case "min":
					fmterr := fmt.Errorf("%s is too short", err.Field())
					kafkaProd.NewUserLog(kafka.LogLevelWarn, place, traceid, fmterr.Error())
					erors[err.Field()] = fmterr
				default:
					fmterr := fmt.Errorf("%s is Null", err.Field())
					kafkaProd.NewUserLog(kafka.LogLevelWarn, place, traceid, fmterr.Error())
					erors[err.Field()] = fmterr
				}
			}
			return erors
		}
	}
	return nil
}
func rollbackTransaction(txMgr repository.DBTransactionManager, tx *sql.Tx, traceid string, place string, kafkaprod kafka.KafkaProducerService) {
	if tx == nil {
		return
	}
	maxAttempts := 3
	attempt := 0
	for attempt < maxAttempts {
		attempt++
		err := txMgr.RollbackTx(tx)
		if err == nil {
			msg := fmt.Sprintf("Successful rollback on attempt %d", attempt)
			kafkaprod.NewUserLog(kafka.LogLevelInfo, place, traceid, msg)
			return
		}
		if errors.Is(err, sql.ErrTxDone) {
			kafkaprod.NewUserLog(kafka.LogLevelWarn, place, traceid, "Transaction is already completed, skipping rollback")
			return
		}
		fmterr := fmt.Sprintf("Error rolling back transaction on attempt %d: %v", attempt, err)
		kafkaprod.NewUserLog(kafka.LogLevelWarn, place, traceid, fmterr)
		if attempt == maxAttempts {
			kafkaprod.NewUserLog(kafka.LogLevelError, place, traceid, "Failed to rollback transaction after all attempts")
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
}
func checkContext(ctx context.Context, mapa map[string]error, place string, traceID string, kafkaprod kafka.KafkaProducerService) (*ServiceResponse, bool) {
	select {
	case <-ctx.Done():
		fmterr := fmt.Sprintf("Context cancelled before operation: %v", ctx.Err())
		kafkaprod.NewUserLog(kafka.LogLevelError, place, traceID, fmterr)
		mapa["InternalServerError"] = erro.ErrorContextTimeout
		return &ServiceResponse{Success: false, Errors: mapa, Type: erro.ServerErrorType}, true
	default:
		return nil, false
	}
}
func retryOperationGrpc[T any](ctx context.Context, operation func(context.Context) (T, error), traceID string, errorMap map[string]error, place string, kafkaprod kafka.KafkaProducerService) (T, *ServiceResponse) {
	var response T
	var err error
	for i := 1; i <= 3; i++ {
		md := metadata.Pairs("traceID", traceID)
		ctx = metadata.NewOutgoingContext(ctx, md)
		if ctxresponse, shouldReturn := checkContext(ctx, errorMap, place, traceID, kafkaprod); shouldReturn {
			return response, ctxresponse
		}
		response, err = operation(ctx)
		if err != nil {
			st, _ := status.FromError(err)
			fmterr := fmt.Sprintf("Operation attempt %d failed", i)
			kafkaprod.NewUserLog(kafka.LogLevelWarn, place, traceID, fmterr)
			switch st.Code() {
			case codes.Internal, codes.Unavailable, codes.Canceled:
				fmterr := fmt.Sprintf("Server unavailable:(%s), retrying...", err)
				kafkaprod.NewUserLog(kafka.LogLevelWarn, place, traceID, fmterr)
				time.Sleep(time.Duration(i) * time.Second)
				continue
			default:
				errorMap["ClientError"] = err
				return response, &ServiceResponse{
					Success: false,
					Errors:  errorMap,
					Type:    erro.ClientErrorType,
				}
			}
		} else {
			return response, nil
		}
	}
	kafkaprod.NewUserLog(kafka.LogLevelError, place, traceID, "All retry attempts failed")
	errorMap["InternalServerError"] = erro.ErrorAllRetryFailed
	return response, &ServiceResponse{
		Success: false,
		Errors:  errorMap,
		Type:    erro.ServerErrorType,
	}
}
func retryOperationDB(ctx context.Context, operation func(context.Context) *repository.DBRepositoryResponse, traceID string, errorMap map[string]error, place string, kafkaprod kafka.KafkaProducerService) (*repository.DBRepositoryResponse, *ServiceResponse) {
	var response *repository.DBRepositoryResponse
	for i := 1; i <= 3; i++ {
		if ctxresponse, shouldReturn := checkContext(ctx, errorMap, place, traceID, kafkaprod); shouldReturn {
			return nil, ctxresponse
		}
		response = operation(ctx)
		if !response.Success && response.Errors != nil {
			fmterr := fmt.Sprintf("Operation attempt %d failed", i)
			kafkaprod.NewUserLog(kafka.LogLevelWarn, place, traceID, fmterr)
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
	kafkaprod.NewUserLog(kafka.LogLevelError, place, traceID, "All retry attempts failed")
	errorMap["InternalServerError"] = erro.ErrorAllRetryFailed
	return nil, &ServiceResponse{
		Success: false,
		Errors:  errorMap,
		Type:    erro.ServerErrorType,
	}
}
