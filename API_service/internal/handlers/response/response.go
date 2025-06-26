package response

import (
	"fmt"
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
	Errors  *erro.CustomError `json:"errors,omitempty"`
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
const KeyPhoto = "photo"
const KeyPhotos = "photos"

func OkResponse(c *gin.Context, status int, data map[string]any, traceid, place string, logproducer LogProducer) {
	sendResponse(c, status, HTTPResponse{Data: data, Success: true}, traceid, place, logproducer)
}
func BadResponse(c *gin.Context, status int, err *erro.CustomError, traceid string, place string, logproducer LogProducer) {
	sendResponse(c, status, HTTPResponse{Success: false, Errors: err}, traceid, place, logproducer)
}
func sendResponse(c *gin.Context, status int, response HTTPResponse, traceid string, place string, logproducer LogProducer) {
	start := c.MustGet("starttime").(time.Time)
	duration := time.Since(start).Seconds()
	err := CheckContext(c, place, traceid, logproducer)
	if err != nil {
		badresp := HTTPResponse{
			Success: false,
			Errors:  erro.ServerError(erro.RequestTimedOut),
		}
		c.JSON(http.StatusInternalServerError, badresp)
		metrics.APITotalBadRequests.WithLabelValues(place, metrics.NormalizePath(c.Request.URL.Path)).Inc()
		metrics.APIBadRequestDuration.WithLabelValues(place, metrics.NormalizePath(c.Request.URL.Path)).Observe(duration)
		return
	}
	c.JSON(status, response)
	if response.Success {
		metrics.APITotalSuccessfulRequests.WithLabelValues(place, metrics.NormalizePath((c.Request.URL.Path))).Inc()
		metrics.APISuccessfulRequestDuration.WithLabelValues(place, metrics.NormalizePath((c.Request.URL.Path))).Observe(duration)
	} else {
		metrics.APITotalBadRequests.WithLabelValues(place, metrics.NormalizePath((c.Request.URL.Path))).Inc()
		metrics.APIBadRequestDuration.WithLabelValues(place, metrics.NormalizePath((c.Request.URL.Path))).Observe(duration)
	}
	logproducer.NewAPILog(c.Request, kafka.LogLevelInfo, place, traceid, "Succesfull send response to client")
}
func CheckContext(c *gin.Context, place string, traceID string, logproducer LogProducer) error {
	select {
	case <-c.Request.Context().Done():
		err := c.Request.Context().Err()
		logproducer.NewAPILog(c.Request, kafka.LogLevelError, place, traceID, fmt.Sprintf("Context's error: %v", err))
		metrics.APIErrorsTotal.WithLabelValues(erro.ServerErrorType).Inc()
		return err
	default:
		return nil
	}
}
