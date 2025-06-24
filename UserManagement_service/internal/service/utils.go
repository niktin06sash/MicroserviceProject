package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/brokers/kafka"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/erro"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/metrics"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/repository"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func validateData[T any](val *validator.Validate, data T, traceid string, place string, logProducer LogProducer) *erro.CustomError {
	err := val.Struct(data)
	if err != nil {
		validationErrors, ok := err.(validator.ValidationErrors)
		if ok {
			for _, err := range validationErrors {
				metrics.UserErrorsTotal.WithLabelValues(erro.ClientErrorType).Inc()
				switch err.Tag() {
				case "email":
					logProducer.NewUserLog(kafka.LogLevelWarn, place, traceid, "Invalid email format")
					return erro.ClientError(erro.ErrorNotEmailConst)
				case "min":
					fmterr := fmt.Sprintf("%s is too short", err.Field())
					logProducer.NewUserLog(kafka.LogLevelWarn, place, traceid, fmterr)
					return erro.ClientError(fmterr)
				case "required":
					fmterr := fmt.Sprintf("%s is Null", err.Field())
					logProducer.NewUserLog(kafka.LogLevelWarn, place, traceid, fmterr)
					return erro.ClientError(fmterr)
				case "max":
					fmterr := fmt.Sprintf("%s is too long", err.Field())
					logProducer.NewUserLog(kafka.LogLevelWarn, place, traceid, fmterr)
					return erro.ClientError(fmterr)
				}
			}
		}
	}
	return nil
}
func (as *UserService) beginTransaction(ctx context.Context, place, traceid string) (*sql.Tx, *ServiceResponse) {
	tx, err := as.Dbtxmanager.BeginTx(ctx)
	metrics.UserDBQueriesTotal.WithLabelValues("Begin Transaction").Inc()
	if err != nil {
		fmterr := fmt.Sprintf("Transaction Error: %v", err)
		as.LogProducer.NewUserLog(kafka.LogLevelError, place, traceid, fmterr)
		metrics.UserDBErrorsTotal.WithLabelValues("Begin Transaction", "Transaction").Inc()
		metrics.UserErrorsTotal.WithLabelValues(erro.ServerErrorType).Inc()
		return nil, &ServiceResponse{Success: false, Errors: erro.ServerError(erro.UserServiceUnavalaible)}
	}
	metrics.UserDBQueriesTotal.WithLabelValues("Begin Transaction").Inc()
	return tx, nil
}
func (as *UserService) rollbackTransaction(tx *sql.Tx, traceid string, place string) {
	if tx == nil {
		return
	}
	maxAttempts := 3
	attempt := 0
	for attempt < maxAttempts {
		attempt++
		err := as.Dbtxmanager.RollbackTx(tx)
		metrics.UserDBQueriesTotal.WithLabelValues("Rollback Transaction").Inc()
		if err == nil {
			msg := fmt.Sprintf("Successful rollback on attempt %d", attempt)
			as.LogProducer.NewUserLog(kafka.LogLevelInfo, place, traceid, msg)
			return
		}
		metrics.UserErrorsTotal.WithLabelValues("InternalServerError").Inc()
		metrics.UserDBErrorsTotal.WithLabelValues("Rollback Transaction", "Transaction").Inc()
		if errors.Is(err, sql.ErrTxDone) {
			as.LogProducer.NewUserLog(kafka.LogLevelWarn, place, traceid, "Transaction is already completed, skipping rollback")
			return
		}
		fmterr := fmt.Sprintf("Error rolling back transaction on attempt %d: %v", attempt, err)
		as.LogProducer.NewUserLog(kafka.LogLevelWarn, place, traceid, fmterr)
		if attempt == maxAttempts {
			as.LogProducer.NewUserLog(kafka.LogLevelError, place, traceid, "Failed to rollback transaction after all attempts")
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
}
func (as *UserService) commitTransaction(tx *sql.Tx, traceid string, place string) error {
	if tx == nil {
		return fmt.Errorf("Transaction is not active")
	}
	maxAttempts := 3
	attempt := 0
	for attempt < maxAttempts {
		attempt++
		err := as.Dbtxmanager.CommitTx(tx)
		metrics.UserDBQueriesTotal.WithLabelValues("Commit Transaction").Inc()
		if err == nil {
			msg := fmt.Sprintf("Successful commit on attempt %d", attempt)
			as.LogProducer.NewUserLog(kafka.LogLevelInfo, place, traceid, msg)
			return nil
		}
		metrics.UserErrorsTotal.WithLabelValues("InternalServerError").Inc()
		metrics.UserDBErrorsTotal.WithLabelValues("Commit Transaction", "Transaction").Inc()
		if errors.Is(err, sql.ErrTxDone) {
			as.LogProducer.NewUserLog(kafka.LogLevelWarn, place, traceid, "Transaction is already completed, skipping commit")
			return nil
		}
		fmterr := fmt.Sprintf("Error commiting back transaction on attempt %d: %v", attempt, err)
		as.LogProducer.NewUserLog(kafka.LogLevelWarn, place, traceid, fmterr)
		if attempt == maxAttempts {
			as.LogProducer.NewUserLog(kafka.LogLevelError, place, traceid, "Failed to commit transaction after all attempts")
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	return fmt.Errorf("Failed to commit transaction after all attempts")
}
func checkContext(ctx context.Context, place string, traceID string, kafkaprod LogProducer) *ServiceResponse {
	select {
	case <-ctx.Done():
		fmterr := fmt.Sprintf("Context cancelled before gRPC-Request: %v", ctx.Err())
		kafkaprod.NewUserLog(kafka.LogLevelError, place, traceID, fmterr)
		metrics.UserErrorsTotal.WithLabelValues(erro.ServerErrorType).Inc()
		return &ServiceResponse{Success: false, Errors: erro.ServerError(erro.RequestTimedOut)}
	default:
		return nil
	}
}
func retryOperationGrpc[T any](ctx context.Context, operation func(context.Context) (T, error), traceID string, place string, kafkaprod LogProducer) (T, *ServiceResponse) {
	var response T
	var err error
	for i := 1; i <= 3; i++ {
		md := metadata.Pairs("traceID", traceID)
		ctx = metadata.NewOutgoingContext(ctx, md)
		if ctxresponse := checkContext(ctx, place, traceID, kafkaprod); ctxresponse != nil {
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
				kafkaprod.NewUserLog(kafka.LogLevelWarn, place, traceID, st.Message())
				metrics.UserErrorsTotal.WithLabelValues(erro.ClientErrorType).Inc()
				return response, &ServiceResponse{
					Success: false,
					Errors:  erro.ClientError(st.Message()),
				}
			}
		} else {
			kafkaprod.NewUserLog(kafka.LogLevelInfo, place, traceID, "Successful gRPC-request to Session-Service")
			return response, nil
		}
	}
	kafkaprod.NewUserLog(kafka.LogLevelError, place, traceID, "All retry attempts failed")
	return response, &ServiceResponse{Success: false, Errors: erro.ServerError(erro.SessionServiceUnavalaible)}
}

func (as *UserService) requestToDB(response *repository.RepositoryResponse, traceid string) (*repository.RepositoryResponse, *ServiceResponse) {
	if !response.Success && response.Errors != nil {
		switch response.Errors.Type {
		case erro.ServerErrorType:
			metrics.UserErrorsTotal.WithLabelValues(erro.ServerErrorType).Inc()
			as.LogProducer.NewUserLog(kafka.LogLevelError, response.Place, traceid, response.Errors.Message)
			response.Errors.Message = erro.UserServiceUnavalaible
			return response, &ServiceResponse{Success: false, Errors: response.Errors}

		case erro.ClientErrorType:
			metrics.UserErrorsTotal.WithLabelValues(erro.ClientErrorType).Inc()
			as.LogProducer.NewUserLog(kafka.LogLevelWarn, response.Place, traceid, response.Errors.Message)
			return response, &ServiceResponse{Success: false, Errors: response.Errors}
		}
	}

	as.LogProducer.NewUserLog(kafka.LogLevelInfo, response.Place, traceid, response.SuccessMessage)
	return response, nil
}
func (as *UserService) parsingUserId(useridstr string, traceid string, place string) (uuid.UUID, error) {
	userid, err := uuid.Parse(useridstr)
	if err != nil {
		fmterr := fmt.Sprintf("UUID-parse Error: %v", err)
		as.LogProducer.NewUserLog(kafka.LogLevelWarn, place, traceid, fmterr)
		metrics.UserErrorsTotal.WithLabelValues(erro.ClientErrorType).Inc()
		return uuid.Nil, err
	}
	return userid, nil
}
func (as *UserService) updateAndCommit(ctx context.Context, tx *sql.Tx, userid uuid.UUID, updateType string, traceid string, place string, args ...interface{}) *ServiceResponse {
	bdresponse, serviceresponse := as.requestToDB(as.Dbrepo.UpdateUserData(ctx, tx, userid, updateType, args...), traceid)
	if serviceresponse != nil {
		as.rollbackTransaction(tx, traceid, place)
		return serviceresponse
	}
	_, serviceresponse = as.requestToDB(as.CacheUserRepos.DeleteProfileCache(ctx, userid.String()), traceid)
	if serviceresponse != nil {
		as.rollbackTransaction(tx, traceid, place)
		return serviceresponse
	}
	err := as.commitTransaction(tx, traceid, place)
	if err != nil {
		as.rollbackTransaction(tx, traceid, place)
		return &ServiceResponse{Success: false, Errors: erro.ServerError(erro.UserServiceUnavalaible)}
	}
	as.LogProducer.NewUserLog(kafka.LogLevelInfo, place, traceid, "Transaction was successfully committed and user's data updates")
	return &ServiceResponse{Success: bdresponse.Success}
}
