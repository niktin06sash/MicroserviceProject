package response

import "github.com/gin-gonic/gin"

type HTTPResponse struct {
	Success bool              `json:"success"`
	Errors  map[string]string `json:"errors"`
	Data    map[string]any    `json:"data,omitempty"`
	Status  int               `json:"status"`
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
