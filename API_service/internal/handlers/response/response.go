package response

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/niktin06sash/MicroserviceProject/API_service/internal/brokers/kafka"
	"github.com/niktin06sash/MicroserviceProject/API_service/internal/erro"
	"github.com/niktin06sash/MicroserviceProject/API_service/internal/metrics"
)

// swagger:model HTTPResponse
type HTTPResponse struct {
	Success bool              `json:"success"`
	Errors  map[string]string `json:"errors,omitempty"`
	Data    map[string]any    `json:"data,omitempty"`
}

// swagger:model PersonReg
type PersonReg struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// swagger:model PersonAuth
type PersonAuth struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// swagger:model PersonDelete
type PersonDelete struct {
	Password string `json:"password"`
}

// swagger:model PersonUpdate
type PersonUpdate struct {
	Name         string `json:"name,omitempty"`
	Email        string `json:"email,omitempty"`
	LastPassword string `json:"last_password,omitempty"`
	NewPassword  string `json:"new_password,omitempty"`
}
type LogProducer interface {
	NewAPILog(c *http.Request, level, place, traceid, msg string)
}

const KeyPhotoID = "photoid"
const KeyMessage = "message"

func OkResponse(c *gin.Context, status int, data map[string]any, traceid, place string, logproducer LogProducer) {
	sendResponse(c, status, HTTPResponse{Data: data, Success: true}, traceid, place, logproducer)
}
func BadResponse(c *gin.Context, status int, errormessage string, traceid string, place string, logproducer LogProducer) {
	if status >= 400 && status < 500 {
		sendResponse(c, status, HTTPResponse{Success: false, Errors: map[string]string{erro.ErrorType: erro.ClientErrorType, erro.ErrorMessage: errormessage}}, traceid, place, logproducer)
	} else {
		sendResponse(c, status, HTTPResponse{Success: false, Errors: map[string]string{erro.ErrorType: erro.ServerErrorType, erro.ErrorMessage: errormessage}}, traceid, place, logproducer)
	}
}
func sendResponse(c *gin.Context, status int, response HTTPResponse, traceid string, place string, logproducer LogProducer) {
	start := c.MustGet("starttime").(time.Time)
	c.JSON(status, response)
	logproducer.NewAPILog(c.Request, kafka.LogLevelInfo, place, traceid, "Succesfull send response to client")
	duration := time.Since(start).Seconds()
	metrics.APITotalBadRequests.WithLabelValues(place, metrics.NormalizePath((c.Request.URL.Path))).Inc()
	metrics.APIBadRequestDuration.WithLabelValues(place, metrics.NormalizePath((c.Request.URL.Path))).Observe(duration)
}
func CheckContext(ctx context.Context, traceID string, place string) error {
	select {
	case <-ctx.Done():
		err := ctx.Err()
		return err
	default:
		return nil
	}
}
