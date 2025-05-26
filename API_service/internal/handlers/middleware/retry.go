package middleware

import (
	"context"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/niktin06sash/MicroserviceProject/API_service/internal/erro"
	"github.com/niktin06sash/MicroserviceProject/API_service/internal/handlers/response"
	"github.com/niktin06sash/MicroserviceProject/API_service/internal/kafka"
	"github.com/niktin06sash/MicroserviceProject/API_service/internal/metrics"
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
			metrics.APIErrorsTotal.WithLabelValues(erro.ServerErrorType).Inc()
			return nil, &erro.CustomError{Message: erro.RequestTimedOut, Type: erro.ServerErrorType}
		}
		protoresponse, err = middleware.grpcClient.ValidateSession(ctx, sessionID)
		if err == nil && protoresponse != nil && protoresponse.Success {
			return protoresponse, nil
		}
		if err != nil {
			st, _ := status.FromError(err)
			fmterr := fmt.Sprintf("Operation attempt %d failed", i)
			middleware.KafkaProducer.NewAPILog(c.Request, kafka.LogLevelWarn, place, traceID, fmterr)
			switch st.Code() {
			case codes.Internal, codes.Unavailable, codes.Canceled:
				middleware.KafkaProducer.NewAPILog(c.Request, kafka.LogLevelWarn, place, traceID, "Session-Service is unavailable, retrying...")
				metrics.APIErrorsTotal.WithLabelValues(erro.ServerErrorType).Inc()
				time.Sleep(time.Duration(i) * time.Second)
				continue
			default:
				middleware.KafkaProducer.NewAPILog(c.Request, kafka.LogLevelWarn, place, traceID, st.Message())
				metrics.APIErrorsTotal.WithLabelValues(erro.ClientErrorType).Inc()
				return nil, &erro.CustomError{Message: st.Message(), Type: erro.ClientErrorType}
			}
		}
	}
	middleware.KafkaProducer.NewAPILog(c.Request, kafka.LogLevelError, place, traceID, "All retry attempts failed")
	return nil, &erro.CustomError{Message: erro.SessionServiceUnavalaible, Type: erro.ServerErrorType}
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
			metrics.APIErrorsTotal.WithLabelValues(erro.ServerErrorType).Inc()
			return nil, &erro.CustomError{Message: "Request timed out", Type: erro.ServerErrorType}
		}
		protoresponse, err = middleware.grpcClient.ValidateSession(ctx, sessionID)
		if err == nil && protoresponse != nil && protoresponse.Success {
			return protoresponse, nil
		}
		if err != nil {
			st, _ := status.FromError(err)
			fmterr := fmt.Sprintf("Operation attempt %d failed", i)
			middleware.KafkaProducer.NewAPILog(c.Request, kafka.LogLevelWarn, place, traceID, fmterr)
			switch st.Code() {
			case codes.Internal, codes.Unavailable, codes.Canceled:
				middleware.KafkaProducer.NewAPILog(c.Request, kafka.LogLevelWarn, place, traceID, "Session-Service is unavailable, retrying...")
				metrics.APIErrorsTotal.WithLabelValues(erro.ServerErrorType).Inc()
				time.Sleep(time.Duration(i) * time.Second)
				continue
			default:
				middleware.KafkaProducer.NewAPILog(c.Request, kafka.LogLevelWarn, place, traceID, st.Message())
				metrics.APIErrorsTotal.WithLabelValues("ClientError").Inc()
				return nil, &erro.CustomError{Message: st.Message(), Type: erro.ClientErrorType}
			}
		}
	}
	middleware.KafkaProducer.NewAPILog(c.Request, kafka.LogLevelError, place, traceID, "All retry attempts failed")
	return nil, &erro.CustomError{Message: erro.SessionServiceUnavalaible, Type: erro.ServerErrorType}
}
