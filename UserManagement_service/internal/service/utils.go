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
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/metrics"
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
					fmterr := fmt.Sprintf("Email format error: %v", err.Error())
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
				metrics.UserErrorsTotal.WithLabelValues("ClientError").Inc()
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
		metrics.UserDBQueriesTotal.WithLabelValues("Rollback Transaction").Inc()
		if err == nil {
			msg := fmt.Sprintf("Successful rollback on attempt %d", attempt)
			kafkaprod.NewUserLog(kafka.LogLevelInfo, place, traceid, msg)
			return
		}
		metrics.UserErrorsTotal.WithLabelValues("InternalServerError").Inc()
		metrics.UserDBErrorsTotal.WithLabelValues("Rollback Transaction", "Transaction").Inc()
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
func commitTransaction(txMgr repository.DBTransactionManager, tx *sql.Tx, traceid string, place string, kafkaprod kafka.KafkaProducerService) error {
	if tx == nil {
		return fmt.Errorf("Transaction is not active")
	}
	maxAttempts := 3
	attempt := 0
	for attempt < maxAttempts {
		attempt++
		err := txMgr.CommitTx(tx)
		metrics.UserDBQueriesTotal.WithLabelValues("Commit Transaction").Inc()
		if err == nil {
			msg := fmt.Sprintf("Successful commit on attempt %d", attempt)
			kafkaprod.NewUserLog(kafka.LogLevelInfo, place, traceid, msg)
			return nil
		}
		metrics.UserErrorsTotal.WithLabelValues("InternalServerError").Inc()
		metrics.UserDBErrorsTotal.WithLabelValues("Commit Transaction", "Transaction").Inc()
		if errors.Is(err, sql.ErrTxDone) {
			kafkaprod.NewUserLog(kafka.LogLevelWarn, place, traceid, "Transaction is already completed, skipping commit")
			return nil
		}
		fmterr := fmt.Sprintf("Error commiting back transaction on attempt %d: %v", attempt, err)
		kafkaprod.NewUserLog(kafka.LogLevelWarn, place, traceid, fmterr)
		if attempt == maxAttempts {
			kafkaprod.NewUserLog(kafka.LogLevelError, place, traceid, "Failed to commit transaction after all attempts")
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	return fmt.Errorf("Failed to commit transaction after all attempts")
}
func checkContext(ctx context.Context, mapa map[string]error, place string, traceID string, kafkaprod kafka.KafkaProducerService) (*ServiceResponse, bool) {
	select {
	case <-ctx.Done():
		fmterr := fmt.Sprintf("Context cancelled before operation: %v", ctx.Err())
		kafkaprod.NewUserLog(kafka.LogLevelError, place, traceID, fmterr)
		mapa["InternalServerError"] = fmt.Errorf("Request timed out")
		metrics.UserErrorsTotal.WithLabelValues("InternalServerError").Inc()
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
				kafkaprod.NewUserLog(kafka.LogLevelWarn, place, traceID, "Session-Service is unavailable, retrying...")
				metrics.UserErrorsTotal.WithLabelValues("InternalServerError").Inc()
				time.Sleep(time.Duration(i) * time.Second)
				continue
			default:
				errorMap["ClientError"] = fmt.Errorf(st.Message())
				kafkaprod.NewUserLog(kafka.LogLevelWarn, place, traceID, st.Message())
				metrics.UserErrorsTotal.WithLabelValues("ClientError").Inc()
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
	errorMap["InternalServerError"] = fmt.Errorf(erro.SessionServiceUnavalaible)
	return response, &ServiceResponse{
		Success: false,
		Errors:  errorMap,
		Type:    erro.ServerErrorType,
	}
}
