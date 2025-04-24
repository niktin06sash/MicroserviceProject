package response

import (
	"context"
	"log"

	"github.com/gin-gonic/gin"
)

type HTTPResponse struct {
	Success bool              `json:"success"`
	Errors  map[string]string `json:"errors"`
	Data    map[string]any    `json:"data,omitempty"`
	Status  int               `json:"status"`
}

type PersonReg struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type PersonAuth struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type PersonDelete struct {
	Password string `json:"password"`
}

func SendResponse(c *gin.Context, status int, success bool, data map[string]any, errors map[string]string) {
	response := HTTPResponse{
		Success: success,
		Data:    data,
		Errors:  errors,
		Status:  status,
	}
	c.JSON(status, response)
}
func CheckContext(ctx context.Context, traceID string, place string) error {
	select {
	case <-ctx.Done():
		err := ctx.Err()
		log.Printf("[ERROR] [API-Service] [%s] [TraceID: %s] ContextError: %s", place, traceID, err)
		return err
	default:
		return nil
	}
}
