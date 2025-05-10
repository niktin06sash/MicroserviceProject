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
		md := metadata.Pairs("traceID", traceID, "flagvalidate", "true")
		ctx = metadata.NewOutgoingContext(ctx, md)
		if err = response.CheckContext(ctx, traceID, place); err != nil {
			fmterr := fmt.Sprintf("Context error: %v", err)
			middleware.KafkaProducer.NewAPILog(c.Request, kafka.LogLevelError, place, traceID, fmterr)
			return nil, &erro.CustomError{ErrorName: "Context Error", ErrorType: erro.ServerErrorType}
		}
		protoresponse, err = middleware.grpcClient.ValidateSession(ctx, sessionID)
		if err == nil && protoresponse.Success {
			return protoresponse, nil
		}
		if err != nil {
			st, _ := status.FromError(err)
			fmterr := fmt.Sprintf("Operation attempt %d failed: %v", i, st.Message())
			middleware.KafkaProducer.NewAPILog(c.Request, kafka.LogLevelWarn, place, traceID, fmterr)
			switch st.Code() {
			case codes.Internal, codes.Unavailable, codes.Canceled:
				fmterr := fmt.Sprintf("Server unavailable:(%s), retrying...", err)
				middleware.KafkaProducer.NewAPILog(c.Request, kafka.LogLevelWarn, place, traceID, fmterr)
				time.Sleep(time.Duration(i) * time.Second)
				continue
			default:
				return nil, &erro.CustomError{ErrorName: "Invalid Session Data", ErrorType: erro.ClientErrorType}
			}
		}
	}
	middleware.KafkaProducer.NewAPILog(c.Request, kafka.LogLevelError, place, traceID, "All retry attempts failed")
	return nil, &erro.CustomError{ErrorName: "All retry attempts failed", ErrorType: erro.ServerErrorType}
}
func retryAuthorized_Not(c *gin.Context, middleware *Middleware, sessionID string, traceID string, place string) (*proto.ValidateSessionResponse, *erro.CustomError) {
	var err error
	var protoresponse *proto.ValidateSessionResponse
	for i := 1; i <= 3; i++ {
		ctx := context.WithValue(c.Request.Context(), "traceID", traceID)
		md := metadata.Pairs("traceID", traceID, "flagvalidate", "false")
		ctx = metadata.NewOutgoingContext(ctx, md)
		if err = response.CheckContext(ctx, traceID, place); err != nil {
			fmterr := fmt.Sprintf("Context error: %v", err)
			middleware.KafkaProducer.NewAPILog(c.Request, kafka.LogLevelError, place, traceID, fmterr)
			return nil, &erro.CustomError{ErrorName: "Context Error", ErrorType: erro.ServerErrorType}
		}
		protoresponse, err = middleware.grpcClient.ValidateSession(ctx, sessionID)
		if err == nil && protoresponse.Success {
			return protoresponse, nil
		}
		if err != nil {
			st, _ := status.FromError(err)
			fmterr := fmt.Sprintf("Operation attempt %d failed: %v", i, st.Message())
			middleware.KafkaProducer.NewAPILog(c.Request, kafka.LogLevelWarn, place, traceID, fmterr)
			switch st.Code() {
			case codes.Internal, codes.Unavailable, codes.Canceled:
				fmterr := fmt.Sprintf("Server unavailable:(%s), retrying...", err)
				middleware.KafkaProducer.NewAPILog(c.Request, kafka.LogLevelWarn, place, traceID, fmterr)
				time.Sleep(time.Duration(i) * time.Second)
				continue
			default:
				return nil, &erro.CustomError{ErrorName: "Already authorized", ErrorType: erro.ClientErrorType}
			}
		}
	}
	middleware.KafkaProducer.NewAPILog(c.Request, kafka.LogLevelError, place, traceID, "All retry attempts failed")
	return nil, &erro.CustomError{ErrorName: "All retry attempts failed", ErrorType: erro.ServerErrorType}
}
