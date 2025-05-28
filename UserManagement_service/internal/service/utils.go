package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/erro"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/kafka"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/metrics"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/repository"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func validateData[T any](val *validator.Validate, data T, traceid string, place string, kafkaProd kafka.KafkaProducerService) map[string]string {
	err := val.Struct(data)
	if err != nil {
		var count = 0
		validationErrors, ok := err.(validator.ValidationErrors)
		if ok {
			erors := make(map[string]string)
			for _, err := range validationErrors {
				count++
				switch err.Tag() {
				case "email":
					kafkaProd.NewUserLog(kafka.LogLevelWarn, place, traceid, "Invalid email format")
					erors[erro.ErrorType] = erro.ClientErrorType + "_" + strconv.Itoa(count)
					erors[erro.ErrorMessage] = erro.ErrorNotEmailConst
				case "min":
					fmterr := fmt.Sprintf("%s is too short", err.Field())
					kafkaProd.NewUserLog(kafka.LogLevelWarn, place, traceid, fmterr)
					erors[erro.ErrorType] = erro.ClientErrorType + "_" + strconv.Itoa(count)
					erors[erro.ErrorMessage] = fmterr
				case "required":
					fmterr := fmt.Sprintf("%s is Null", err.Field())
					kafkaProd.NewUserLog(kafka.LogLevelWarn, place, traceid, fmterr)
					erors[erro.ErrorType] = erro.ClientErrorType + "_" + strconv.Itoa(count)
					erors[erro.ErrorMessage] = fmterr
				}
				metrics.UserErrorsTotal.WithLabelValues(erro.ClientErrorType).Inc()
			}
			return erors
		}
	}
	return nil
}
func beginTransaction(ctx context.Context, txman repository.DBTransactionManager, mapa map[string]string, place, traceid string, kafkaprod kafka.KafkaProducerService) (*sql.Tx, error) {
	tx, err := txman.BeginTx(ctx)
	metrics.UserDBQueriesTotal.WithLabelValues("Begin Transaction").Inc()
	if err != nil {
		fmterr := fmt.Sprintf("Transaction Error: %v", err)
		kafkaprod.NewUserLog(kafka.LogLevelError, place, traceid, fmterr)
		metrics.UserDBErrorsTotal.WithLabelValues("Begin Transaction", "Transaction").Inc()
		metrics.UserErrorsTotal.WithLabelValues(erro.ServerErrorType).Inc()
		return nil, err
	}
	metrics.UserDBQueriesTotal.WithLabelValues("Begin Transaction").Inc()
	return tx, nil
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
func commitTransaction(txMgr repository.DBTransactionManager, tx *sql.Tx, mapa map[string]string, traceid string, place string, kafkaprod kafka.KafkaProducerService) error {
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
	mapa[erro.ErrorType] = erro.ServerErrorType
	mapa[erro.ErrorMessage] = erro.UserServiceUnavalaible
	return fmt.Errorf("Failed to commit transaction after all attempts")
}
func checkContext(ctx context.Context, mapa map[string]string, place string, traceID string, kafkaprod kafka.KafkaProducerService) *ServiceResponse {
	select {
	case <-ctx.Done():
		fmterr := fmt.Sprintf("Context cancelled before operation: %v", ctx.Err())
		kafkaprod.NewUserLog(kafka.LogLevelError, place, traceID, fmterr)
		mapa[erro.ErrorType] = erro.ServerErrorType
		mapa[erro.ErrorMessage] = erro.RequestTimedOut
		metrics.UserErrorsTotal.WithLabelValues(erro.ServerErrorType).Inc()
		return &ServiceResponse{Success: false, Errors: mapa, ErrorType: erro.ServerErrorType}
	default:
		return nil
	}
}
func retryOperationGrpc[T any](ctx context.Context, operation func(context.Context) (T, error), traceID string, errorMap map[string]string, place string, kafkaprod kafka.KafkaProducerService) (T, *ServiceResponse) {
	var response T
	var err error
	for i := 1; i <= 3; i++ {
		md := metadata.Pairs("traceID", traceID)
		ctx = metadata.NewOutgoingContext(ctx, md)
		if ctxresponse := checkContext(ctx, errorMap, place, traceID, kafkaprod); ctxresponse != nil {
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
				errorMap[erro.ErrorType] = erro.ClientErrorType
				errorMap[erro.ErrorMessage] = st.Message()
				kafkaprod.NewUserLog(kafka.LogLevelWarn, place, traceID, st.Message())
				metrics.UserErrorsTotal.WithLabelValues(erro.ClientErrorType).Inc()
				return response, &ServiceResponse{
					Success:   false,
					Errors:    errorMap,
					ErrorType: erro.ClientErrorType,
				}
			}
		} else {
			return response, nil
		}
	}
	kafkaprod.NewUserLog(kafka.LogLevelError, place, traceID, "All retry attempts failed")
	errorMap[erro.ErrorType] = erro.ServerErrorType
	errorMap[erro.ErrorMessage] = erro.SessionServiceUnavalaible
	return response, &ServiceResponse{
		Success:   false,
		Errors:    errorMap,
		ErrorType: erro.ServerErrorType,
	}
}

func requestToDB(response *repository.RepositoryResponse, mapa map[string]string) (*repository.RepositoryResponse, *ServiceResponse) {
	if !response.Success && response.Errors != nil {
		switch response.Errors.Type {
		case erro.ServerErrorType:
			mapa[erro.ErrorType] = erro.ServerErrorType
			mapa[erro.ErrorMessage] = response.Errors.Message
			metrics.UserErrorsTotal.WithLabelValues(erro.ServerErrorType).Inc()
			return response, &ServiceResponse{
				Success:   false,
				Errors:    mapa,
				ErrorType: erro.ServerErrorType,
			}

		case erro.ClientErrorType:
			mapa[erro.ErrorType] = erro.ClientErrorType
			mapa[erro.ErrorMessage] = response.Errors.Message
			metrics.UserErrorsTotal.WithLabelValues(erro.ClientErrorType).Inc()
			return response, &ServiceResponse{
				Success:   false,
				Errors:    mapa,
				ErrorType: erro.ClientErrorType,
			}
		}
	}
	return response, nil
}
func parsingUserId(useridstr string, mapa map[string]string, traceid string, place string, kafkaprod kafka.KafkaProducerService) (uuid.UUID, error) {
	userid, err := uuid.Parse(useridstr)
	if err != nil {
		fmterr := fmt.Sprintf("UUID-parse Error: %v", err)
		kafkaprod.NewUserLog(kafka.LogLevelError, place, traceid, fmterr)
		mapa[erro.ErrorType] = erro.ServerErrorType
		mapa[erro.ErrorMessage] = erro.UserServiceUnavalaible
		metrics.UserErrorsTotal.WithLabelValues(erro.ServerErrorType).Inc()
		return uuid.Nil, err
	}
	return userid, nil
}
func updateAndCommit(
	ctx context.Context,
	txManager repository.DBTransactionManager,
	dbFunc func(context.Context, *sql.Tx, uuid.UUID, string, ...interface{}) *repository.RepositoryResponse,
	redisFunc func(context.Context, string) *repository.RepositoryResponse,
	tx *sql.Tx,
	userid uuid.UUID,
	updateType string,
	mapa map[string]string,
	traceid string,
	place string,
	kafkaProducer kafka.KafkaProducerService,
	args ...interface{},
) *ServiceResponse {
	bdresponse, serviceresponse := requestToDB(dbFunc(ctx, tx, userid, updateType, args...), mapa)
	if serviceresponse != nil {
		rollbackTransaction(txManager, tx, traceid, place, kafkaProducer)
		return serviceresponse
	}
	_, serviceresponse = requestToDB(redisFunc(ctx, userid.String()), mapa)
	if serviceresponse != nil {
		rollbackTransaction(txManager, tx, traceid, place, kafkaProducer)
		return serviceresponse
	}
	err := commitTransaction(txManager, tx, mapa, traceid, place, kafkaProducer)
	if err != nil {
		rollbackTransaction(txManager, tx, traceid, place, kafkaProducer)
		return &ServiceResponse{Success: false, Errors: mapa, ErrorType: erro.ServerErrorType}
	}
	kafkaProducer.NewUserLog(kafka.LogLevelInfo, place, traceid, "Transaction was successfully committed and profile's data updates")
	return &ServiceResponse{Success: bdresponse.Success}
}
