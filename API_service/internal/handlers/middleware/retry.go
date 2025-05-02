package middleware

import (
	"context"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/niktin06sash/MicroserviceProject/API_service/internal/erro"
	"github.com/niktin06sash/MicroserviceProject/API_service/internal/handlers/response"
	"github.com/niktin06sash/MicroserviceProject/API_service/internal/kafka"
	"github.com/niktin06sash/MicroserviceProject/SessionManagement_service/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func retryAuthorized(c *gin.Context, middleware *Middleware, sessionID string, traceID string, place string) (*proto.ValidateSessionResponse, *erro.CustomError) {
	var err error
	var protoresponse *proto.ValidateSessionResponse
	for i := 1; i <= 3; i++ {
		ctx := context.WithValue(c.Request.Context(), "traceID", traceID)
		md := metadata.Pairs("traceID", traceID)
		ctx = metadata.NewOutgoingContext(ctx, md)
		if err = response.CheckContext(ctx, traceID, place); err != nil {
			middleware.KafkaProducer.NewAPILog(kafka.APILog{
				Level:     kafka.LogLevelError,
				Place:     place,
				TraceID:   traceID,
				IP:        c.Request.RemoteAddr,
				Method:    c.Request.Method,
				Path:      c.Request.URL.Path,
				Timestamp: time.Now().Format(time.RFC3339),
				Message:   "Context cancelled before operation",
			})
			return nil, &erro.CustomError{ErrorName: "Context Error", ErrorType: erro.ServerErrorType}
		}
		protoresponse, err = middleware.grpcClient.ValidateSession(ctx, sessionID)
		if err == nil && protoresponse.Success {
			middleware.KafkaProducer.NewAPILog(kafka.APILog{
				Level:     kafka.LogLevelInfo,
				Place:     place,
				TraceID:   traceID,
				IP:        c.Request.RemoteAddr,
				Method:    c.Request.Method,
				Path:      c.Request.URL.Path,
				Timestamp: time.Now().Format(time.RFC3339),
				Message:   "Successful gRPC request",
			})
			return protoresponse, nil
		}
		if err != nil {
			st, _ := status.FromError(err)
			trymessage := fmt.Sprintf("Operation attempt %d failed: %v", i, st.Message())
			middleware.KafkaProducer.NewAPILog(kafka.APILog{
				Level:     kafka.LogLevelWarn,
				Place:     place,
				TraceID:   traceID,
				IP:        c.Request.RemoteAddr,
				Method:    c.Request.Method,
				Path:      c.Request.URL.Path,
				Timestamp: time.Now().Format(time.RFC3339),
				Message:   trymessage,
			})
			switch st.Code() {
			case codes.Internal, codes.Unavailable, codes.Canceled:
				logmsg := fmt.Sprintf("Server unavailable:(%s), retrying...", err)
				middleware.KafkaProducer.NewAPILog(kafka.APILog{
					Level:     kafka.LogLevelWarn,
					Place:     place,
					TraceID:   traceID,
					IP:        c.Request.RemoteAddr,
					Method:    c.Request.Method,
					Path:      c.Request.URL.Path,
					Timestamp: time.Now().Format(time.RFC3339),
					Message:   logmsg,
				})
				time.Sleep(time.Duration(i) * time.Second)
				continue
			default:
				return protoresponse, &erro.CustomError{ErrorName: "Invalid Session Data", ErrorType: erro.ClientErrorType}
			}
		}
	}
	middleware.KafkaProducer.NewAPILog(kafka.APILog{
		Level:     kafka.LogLevelError,
		Place:     place,
		TraceID:   traceID,
		IP:        c.Request.RemoteAddr,
		Method:    c.Request.Method,
		Path:      c.Request.URL.Path,
		Timestamp: time.Now().Format(time.RFC3339),
		Message:   "All retry attempts failed",
	})
	return protoresponse, &erro.CustomError{ErrorName: "All retry attempts failed", ErrorType: erro.ServerErrorType}
}
func retryAuthorized_Not(c *gin.Context, middleware *Middleware, sessionID string, traceID string, place string) (*proto.ValidateSessionResponse, *erro.CustomError) {
	var err error
	var protoresponse *proto.ValidateSessionResponse
	for i := 1; i <= 3; i++ {
		ctx := context.WithValue(c.Request.Context(), "traceID", traceID)
		md := metadata.Pairs("traceID", traceID)
		ctx = metadata.NewOutgoingContext(ctx, md)
		if err = response.CheckContext(ctx, traceID, place); err != nil {
			middleware.KafkaProducer.NewAPILog(kafka.APILog{
				Level:     kafka.LogLevelError,
				Place:     place,
				TraceID:   traceID,
				IP:        c.Request.RemoteAddr,
				Method:    c.Request.Method,
				Path:      c.Request.URL.Path,
				Timestamp: time.Now().Format(time.RFC3339),
				Message:   "Context cancelled before operation",
			})
			return nil, &erro.CustomError{ErrorName: "Context Error", ErrorType: erro.ServerErrorType}
		}
		protoresponse, err = middleware.grpcClient.ValidateSession(ctx, sessionID)
		if err != nil {
			st, _ := status.FromError(err)
			switch st.Code() {
			case codes.Internal, codes.Unavailable, codes.Canceled:
				trymessage := fmt.Sprintf("Operation attempt %d failed: %v, (%s)", i, st.Message(), err)
				middleware.KafkaProducer.NewAPILog(kafka.APILog{
					Level:     kafka.LogLevelWarn,
					Place:     place,
					TraceID:   traceID,
					IP:        c.Request.RemoteAddr,
					Method:    c.Request.Method,
					Path:      c.Request.URL.Path,
					Timestamp: time.Now().Format(time.RFC3339),
					Message:   trymessage,
				})
				time.Sleep(time.Duration(i) * time.Second)
				continue
			default:
				return protoresponse, nil
			}
		}
		if protoresponse.Success {
			return protoresponse, &erro.CustomError{
				ErrorName: "Already authorized",
				ErrorType: erro.ClientErrorType,
			}
		}
	}
	middleware.KafkaProducer.NewAPILog(kafka.APILog{
		Level:     kafka.LogLevelError,
		Place:     place,
		TraceID:   traceID,
		IP:        c.Request.RemoteAddr,
		Method:    c.Request.Method,
		Path:      c.Request.URL.Path,
		Timestamp: time.Now().Format(time.RFC3339),
		Message:   "All retry attempts failed",
	})
	return protoresponse, &erro.CustomError{ErrorName: "All retry attempts failed", ErrorType: erro.ServerErrorType}
}
