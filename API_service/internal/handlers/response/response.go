package response

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/niktin06sash/MicroserviceProject/API_service/internal/kafka"
	"github.com/niktin06sash/MicroserviceProject/API_service/internal/metrics"
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
	start := c.MustGet("starttime").(time.Time)
	response := HTTPResponse{
		Success: success,
		Data:    data,
		Errors:  errors,
		Status:  status,
	}
	c.JSON(status, response)
	kafkaprod.NewAPILog(c.Request, kafka.LogLevelInfo, place, traceid, "Succesfull send response to client")
	duration := time.Since(start).Seconds()
	metrics.APIRequestDuration.WithLabelValues(place, c.Request.URL.Path).Observe(duration)
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
