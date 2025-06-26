package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/niktin06sash/MicroserviceProject/API_service/internal/brokers/kafka"
	"github.com/niktin06sash/MicroserviceProject/API_service/internal/erro"
	"github.com/niktin06sash/MicroserviceProject/API_service/internal/handlers/response"
	"github.com/niktin06sash/MicroserviceProject/API_service/internal/metrics"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (h *Handler) badGrpcResponse(c *gin.Context, traceID, place string, err error) {
	st, _ := status.FromError(err)
	switch st.Code() {
	case codes.Canceled, codes.Unavailable:
		h.logproducer.NewAPILog(c.Request, kafka.LogLevelError, st.Message(), traceID, place)
		response.BadResponse(c, http.StatusInternalServerError, erro.ServerError(erro.PhotoServiceUnavalaible), traceID, place, h.logproducer)
	case codes.Internal:
		h.logproducer.NewAPILog(c.Request, kafka.LogLevelError, st.Message(), traceID, place)
		response.BadResponse(c, http.StatusInternalServerError, erro.ServerError(st.Message()), traceID, place, h.logproducer)
	default:
		h.logproducer.NewAPILog(c.Request, kafka.LogLevelWarn, st.Message(), traceID, place)
		response.BadResponse(c, http.StatusBadRequest, erro.ClientError(st.Message()), traceID, place, h.logproducer)
	}
}
func (h *Handler) badHttpResponse(c *gin.Context, traceID, place string, userresponse response.HTTPResponse) bool {
	if !userresponse.Success {
		if userresponse.Errors.Type == erro.ServerErrorType {
			response.BadResponse(c, http.StatusInternalServerError, userresponse.Errors, traceID, place, h.logproducer)
		} else {
			response.BadResponse(c, http.StatusBadRequest, userresponse.Errors, traceID, place, h.logproducer)
		}
		return true
	}
	return false
}
func (h *Handler) asyncHTTPRequest(c *gin.Context, target string, place string, httpresponseChan chan response.HTTPResponse) error {
	traceID := c.MustGet("traceID").(string)
	userid := c.MustGet("userID").(string)
	sessionid := c.MustGet("sessionID").(string)
	deadline, ok := c.Request.Context().Deadline()
	if !ok {
		h.logproducer.NewAPILog(c.Request, kafka.LogLevelWarn, place, traceID, "Failed to get deadline from context")
		deadline = time.Now().Add(15 * time.Second)
	}
	httprequest, err := http.NewRequestWithContext(c.Request.Context(), http.MethodGet, target, c.Request.Body)
	if err != nil {
		h.logproducer.NewAPILog(c.Request, kafka.LogLevelError, place, traceID, fmt.Sprintf("New http-request create error: %v", err))
		metrics.APIErrorsTotal.WithLabelValues(erro.ServerErrorType).Inc()
		return fmt.Errorf(erro.APIServiceUnavalaible)
	}
	httprequest.Header.Set("X-Deadline", deadline.Format(time.RFC3339))
	httprequest.Header.Set("X-User-ID", userid)
	httprequest.Header.Set("X-Trace-ID", traceID)
	httprequest.Header.Set("X-Session-ID", sessionid)
	httpresponse, err := http.DefaultClient.Do(httprequest)
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
func (h *Handler) asyncBadResponse(c *gin.Context, traceID string, place string, err error) {
	if _, ok := status.FromError(err); ok {
		h.badGrpcResponse(c, traceID, place, err)
		return
	}
	response.BadResponse(c, http.StatusInternalServerError, erro.ServerError(err.Error()), traceID, place, h.logproducer)
}
