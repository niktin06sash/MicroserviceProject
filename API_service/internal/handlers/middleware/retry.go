package middleware

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/niktin06sash/MicroserviceProject/API_service/internal/brokers/kafka"
	"github.com/niktin06sash/MicroserviceProject/API_service/internal/erro"
	"github.com/niktin06sash/MicroserviceProject/API_service/internal/handlers/response"
	"github.com/niktin06sash/MicroserviceProject/API_service/internal/metrics"
	"github.com/niktin06sash/MicroserviceProject/SessionManagement_service/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func (middleware *Middleware) retryAuthorized(c *gin.Context, sessionID string, traceID string, place string) (*proto.ValidateSessionResponse, map[string]string) {
	var err error
	var protoresponse *proto.ValidateSessionResponse
	for i := 1; i <= 3; i++ {
		md := metadata.Pairs("traceID", traceID, "flagvalidate", "true")
		ctx := metadata.NewOutgoingContext(c.Request.Context(), md)
		if err = response.CheckContext(c, place, traceID, middleware.logproducer); err != nil {
			return nil, map[string]string{erro.ErrorType: erro.ServerErrorType, erro.ErrorMessage: erro.RequestTimedOut}
		}
		protoresponse, err = middleware.grpcClient.ValidateSession(ctx, sessionID)
		if err == nil && protoresponse != nil && protoresponse.Success {
			return protoresponse, nil
		}
		if err != nil {
			st, _ := status.FromError(err)
			fmterr := fmt.Sprintf("Operation attempt %d failed", i)
			middleware.logproducer.NewAPILog(c.Request, kafka.LogLevelWarn, place, traceID, fmterr)
			switch st.Code() {
			case codes.Internal, codes.Unavailable, codes.Canceled:
				middleware.logproducer.NewAPILog(c.Request, kafka.LogLevelWarn, place, traceID, "Session-Service is unavailable, retrying...")
				metrics.APIErrorsTotal.WithLabelValues(erro.ServerErrorType).Inc()
				time.Sleep(time.Duration(i) * time.Second)
				continue
			default:
				middleware.logproducer.NewAPILog(c.Request, kafka.LogLevelWarn, place, traceID, st.Message())
				metrics.APIErrorsTotal.WithLabelValues(erro.ClientErrorType).Inc()
				return nil, map[string]string{erro.ErrorType: erro.ClientErrorType, erro.ErrorMessage: st.Message()}
			}
		}
	}
	middleware.logproducer.NewAPILog(c.Request, kafka.LogLevelError, place, traceID, "All retry attempts failed")
	return nil, map[string]string{erro.ErrorType: erro.ServerErrorType, erro.ErrorMessage: erro.SessionServiceUnavalaible}
}
func (middleware *Middleware) retryAuthorized_Not(c *gin.Context, sessionID string, traceID string, place string) (*proto.ValidateSessionResponse, map[string]string) {
	var err error
	var protoresponse *proto.ValidateSessionResponse
	for i := 1; i <= 3; i++ {
		md := metadata.Pairs("traceID", traceID, "flagvalidate", "false")
		ctx := metadata.NewOutgoingContext(c.Request.Context(), md)
		if err = response.CheckContext(c, place, traceID, middleware.logproducer); err != nil {
			return nil, map[string]string{erro.ErrorType: erro.ServerErrorType, erro.ErrorMessage: erro.RequestTimedOut}
		}
		protoresponse, err = middleware.grpcClient.ValidateSession(ctx, sessionID)
		if err == nil && protoresponse != nil && protoresponse.Success {
			return protoresponse, nil
		}
		if err != nil {
			st, _ := status.FromError(err)
			fmterr := fmt.Sprintf("Operation attempt %d failed", i)
			middleware.logproducer.NewAPILog(c.Request, kafka.LogLevelWarn, place, traceID, fmterr)
			switch st.Code() {
			case codes.Internal, codes.Unavailable, codes.Canceled:
				middleware.logproducer.NewAPILog(c.Request, kafka.LogLevelWarn, place, traceID, "Session-Service is unavailable, retrying...")
				metrics.APIErrorsTotal.WithLabelValues(erro.ServerErrorType).Inc()
				time.Sleep(time.Duration(i) * time.Second)
				continue
			default:
				middleware.logproducer.NewAPILog(c.Request, kafka.LogLevelWarn, place, traceID, st.Message())
				metrics.APIErrorsTotal.WithLabelValues("ClientError").Inc()
				return nil, map[string]string{erro.ErrorType: erro.ClientErrorType, erro.ErrorMessage: st.Message()}
			}
		}
	}
	middleware.logproducer.NewAPILog(c.Request, kafka.LogLevelError, place, traceID, "All retry attempts failed")
	return nil, map[string]string{erro.ErrorType: erro.ServerErrorType, erro.ErrorMessage: erro.SessionServiceUnavalaible}
}
