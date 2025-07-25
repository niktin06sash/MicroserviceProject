package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/niktin06sash/MicroserviceProject/API_service/internal/brokers/kafka"
	"github.com/niktin06sash/MicroserviceProject/API_service/internal/erro"
	"github.com/niktin06sash/MicroserviceProject/API_service/internal/handlers/response"
	"github.com/niktin06sash/MicroserviceProject/API_service/internal/metrics"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (h *Handler) asyncHTTPRequest(c *gin.Context, target string, place string, httpresponseChan chan response.HTTPResponse) error {
	traceID := c.MustGet("traceID").(string)
	userid := c.MustGet("userID").(string)
	sessionid := c.MustGet("sessionID").(string)
	httprequest, err := http.NewRequestWithContext(c.Request.Context(), http.MethodGet, target, c.Request.Body)
	if err != nil {
		h.logproducer.NewAPILog(c.Request, kafka.LogLevelError, place, traceID, fmt.Sprintf("New http-request create error: %v", err))
		metrics.APIErrorsTotal.WithLabelValues(erro.ServerErrorType).Inc()
		return fmt.Errorf(erro.APIServiceUnavalaible)
	}
	httprequest.Header.Set("X-User-ID", userid)
	httprequest.Header.Set("X-Trace-ID", traceID)
	httprequest.Header.Set("X-Session-ID", sessionid)
	httpresponse, err := h.httpclient.Do(httprequest)
	if err != nil {
		h.logproducer.NewAPILog(c.Request, kafka.LogLevelError, place, traceID, fmt.Sprintf("Http-request error: %v", err))
		metrics.APIErrorsTotal.WithLabelValues(erro.ServerErrorType).Inc()
		return fmt.Errorf(erro.UserServiceUnavalaible)
	}
	defer httpresponse.Body.Close()
	data, err := io.ReadAll(httpresponse.Body)
	if err != nil {
		h.logproducer.NewAPILog(c.Request, kafka.LogLevelError, place, traceID, fmt.Sprintf("ReadAll http-response data error: %v", err))
		metrics.APIErrorsTotal.WithLabelValues(erro.ServerErrorType).Inc()
		return fmt.Errorf(erro.APIServiceUnavalaible)
	}
	var userresponse response.HTTPResponse
	err = json.Unmarshal(data, &userresponse)
	if err != nil {
		h.logproducer.NewAPILog(c.Request, kafka.LogLevelError, place, traceID, fmt.Sprintf("Unmarshal http-response data error: %v", err))
		metrics.APIErrorsTotal.WithLabelValues(erro.ServerErrorType).Inc()
		return fmt.Errorf(erro.APIServiceUnavalaible)
	}
	httpresponseChan <- userresponse
	return nil
}

func asyncgRPCRequest[T any](context context.Context, operation func(context.Context) (T, error), protoresponseChan chan T) error {
	protoresponse, err := operation(context)
	if err != nil {
		st, _ := status.FromError(err)
		switch st.Code() {
		case codes.Canceled, codes.Unavailable:
			return fmt.Errorf(erro.PhotoServiceUnavalaible)
		default:
			return err
		}
	}
	protoresponseChan <- protoresponse
	return nil
}
func (h *Handler) handleGrpcError(c *gin.Context, traceID, place string, err error) {
	st, _ := status.FromError(err)
	switch st.Code() {
	case codes.Canceled, codes.Unavailable:
		response.BadResponse(c, http.StatusInternalServerError, erro.ServerError(erro.PhotoServiceUnavalaible), traceID, place, h.logproducer)
	case codes.Internal:
		response.BadResponse(c, http.StatusInternalServerError, erro.ServerError(st.Message()), traceID, place, h.logproducer)
	default:
		response.BadResponse(c, http.StatusBadRequest, erro.ClientError(st.Message()), traceID, place, h.logproducer)
	}
}
func (h *Handler) badHttpResponse(c *gin.Context, traceID, place string, userresponse response.HTTPResponse) bool {
	if !userresponse.Success {
		var customError *erro.CustomError
		if errors.As(userresponse.Errors, &customError) {
			switch customError.Type {
			case erro.ServerErrorType:
				response.BadResponse(c, http.StatusInternalServerError, userresponse.Errors, traceID, place, h.logproducer)
			case erro.ClientErrorType:
				response.BadResponse(c, http.StatusBadRequest, userresponse.Errors, traceID, place, h.logproducer)
			}
		}
		return true
	}
	return false
}
