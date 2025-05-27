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

func SendResponse(c *gin.Context, status int, response HTTPResponse, traceid string, place string, kafkaprod kafka.KafkaProducerService) {
	start := c.MustGet("starttime").(time.Time)
	c.JSON(status, response)
	kafkaprod.NewAPILog(c.Request, kafka.LogLevelInfo, place, traceid, "Succesfull send response to client")
	duration := time.Since(start).Seconds()
	metrics.APITotalSuccessfulRequests.WithLabelValues(place, metrics.NormalizePath((c.Request.URL.Path))).Inc()
	metrics.APIRequestDuration.WithLabelValues(place, metrics.NormalizePath((c.Request.URL.Path))).Observe(duration)
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
