package response

import (
	"encoding/json"
	"log"
	"net/http"
)

type HTTPResponse struct {
	Success bool              `json:"success"`
	Errors  map[string]string `json:"errors"`
	Data    map[string]any    `json:"data,omitempty"`
	Status  int               `json:"status"`
}

func NewSuccessResponse(data map[string]any, status int) HTTPResponse {
	if data == nil {
		data = make(map[string]any)
	}
	return HTTPResponse{
		Success: true,
		Data:    data,
		Status:  status,
	}
}
func NewErrorResponse(errors map[string]string, status int) HTTPResponse {
	if errors == nil {
		errors = make(map[string]string)
	}
	return HTTPResponse{
		Success: false,
		Errors:  errors,
		Status:  status,
	}
}
func SendResponse(w http.ResponseWriter, resp HTTPResponse, reqid string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.Status)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("[ERROR] [UserManagement] [RequestID: %s]: Failed to encode response: %v", reqid, err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(NewErrorResponse(map[string]string{
			"Internal": "InternalServerError",
		}, resp.Status))
	}
	log.Printf("[INFO] [UserManagement] [RequestID: %s]: Succesfull send response to client", reqid)
}
func ConvertErrorsToString(errors map[string]error) map[string]string {
	stringMap := make(map[string]string)
	for key, err := range errors {
		if err != nil {
			stringMap[key] = err.Error()
		}
	}
	return stringMap
}
