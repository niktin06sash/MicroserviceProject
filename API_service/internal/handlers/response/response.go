package response

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/niktin06sash/MicroserviceProject/API_service/internal/kafka"
)

// swagger:model HTTPResponse
type HTTPResponse struct {
	Success bool              `json:"success"`
	Errors  map[string]string `json:"errors"`
	Data    map[string]any    `json:"data,omitempty"`
	Status  int               `json:"status"`
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

func SendResponse(c *gin.Context, status int, success bool, data map[string]any, errors map[string]string, traceid string, place string, kafkaprod kafka.KafkaProducerService) {
	response := HTTPResponse{
		Success: success,
		Data:    data,
		Errors:  errors,
		Status:  status,
	}
	c.JSON(status, response)
	kafkaprod.NewAPILog(kafka.APILog{
		Level:     kafka.LogLevelInfo,
		Place:     place,
		TraceID:   traceid,
		IP:        c.Request.RemoteAddr,
		Method:    c.Request.Method,
		Path:      c.Request.URL.Path,
		Timestamp: time.Now().Format(time.RFC3339),
		Message:   "Succesfull send response to client",
	})
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
