package middleware

import (
	"context"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/niktin06sash/MicroserviceProject/API_service/internal/client"
	"github.com/niktin06sash/MicroserviceProject/API_service/internal/erro"
	"github.com/niktin06sash/MicroserviceProject/API_service/internal/handlers/response"
	"github.com/niktin06sash/MicroserviceProject/SessionManagement_service/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func retryAuthorized(c *gin.Context, client client.GrpcClientService, sessionID string, traceID string, place string) (*proto.ValidateSessionResponse, erro.ErrorInterface) {
	var err error
	var protoresponse *proto.ValidateSessionResponse
	for i := 1; i <= 3; i++ {
		ctx := context.WithValue(c.Request.Context(), "traceID", traceID)
		md := metadata.Pairs("traceID", traceID)
		ctx = metadata.NewOutgoingContext(ctx, md)
		if err = response.CheckContext(ctx, traceID, place); err != nil {
			logRequest(c.Request, place, traceID, true, "Context cancelled before operation")
			return nil, erro.CustomError{ErrorName: "Context Error", ErrorType: erro.ServerErrorType}
		}
		protoresponse, err = client.ValidateSession(ctx, sessionID)
		if err == nil && protoresponse.Success {
			logRequest(c.Request, place, traceID, false, "Successful gRPC request")
			return protoresponse, nil
		}
		if err != nil {
			st, _ := status.FromError(err)
			trymessage := fmt.Sprintf("Operation attempt %d failed: %v", i, st.Message())
			logRequest(c.Request, place, traceID, true, trymessage)
			switch st.Code() {
			case codes.Internal, codes.Unavailable, codes.Canceled:
				logRequest(c.Request, place, traceID, true, "Server unavailable, retrying...")
				time.Sleep(time.Duration(i) * time.Second)
				continue
			default:
				return protoresponse, erro.CustomError{ErrorName: "Invalid Session Data", ErrorType: erro.ClientErrorType}
			}
		}
	}
	logRequest(c.Request, place, traceID, true, "All retry attempts failed")
	return protoresponse, erro.CustomError{ErrorName: "All retry attempts failed", ErrorType: erro.ServerErrorType}
}
func retryAuthorized_Not(c *gin.Context, client client.GrpcClientService, sessionID string, traceID string, place string) (*proto.ValidateSessionResponse, erro.ErrorInterface) {
	var err error
	var protoresponse *proto.ValidateSessionResponse
	for i := 1; i <= 3; i++ {
		ctx := context.WithValue(c.Request.Context(), "traceID", traceID)
		md := metadata.Pairs("traceID", traceID)
		ctx = metadata.NewOutgoingContext(ctx, md)
		if err = response.CheckContext(ctx, traceID, place); err != nil {
			logRequest(c.Request, place, traceID, true, "Context cancelled before operation")
			return nil, erro.CustomError{ErrorName: "Context Error", ErrorType: erro.ServerErrorType}
		}
		protoresponse, err = client.ValidateSession(ctx, sessionID)
		if err != nil {
			st, _ := status.FromError(err)
			switch st.Code() {
			case codes.Internal, codes.Unavailable, codes.Canceled:
				trymessage := fmt.Sprintf("Operation attempt %d failed: %v", i, st.Message())
				logRequest(c.Request, place, traceID, true, trymessage)
				time.Sleep(time.Duration(i) * time.Second)
				continue
			default:
				return protoresponse, nil
			}
		}
		if protoresponse.Success {
			return protoresponse, erro.CustomError{
				ErrorName: "Already authorized",
				ErrorType: erro.ClientErrorType,
			}
		}
	}
	logRequest(c.Request, place, traceID, true, "All retry attempts failed")
	return protoresponse, erro.CustomError{ErrorName: "All retry attempts failed", ErrorType: erro.ServerErrorType}
}
