package response

import (
	"context"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
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

func SendResponse(c *gin.Context, status int, success bool, data map[string]any, errors map[string]string, traceid string, place string) {
	err := CheckContext(c, traceid, place)
	if err != nil {
		response := HTTPResponse{
			Success: false,
			Data:    nil,
			Errors:  map[string]string{"InternalServerError": "Context deadline exceeded"},
			Status:  http.StatusInternalServerError,
		}
		c.JSON(http.StatusInternalServerError, response)
		return
	}
	response := HTTPResponse{
		Success: success,
		Data:    data,
		Errors:  errors,
		Status:  status,
	}
	c.JSON(status, response)
	log.Printf("[INFO] [API-Service] [%s] [TraceID: %s]: Succesfull send response to client", traceid, place)
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
