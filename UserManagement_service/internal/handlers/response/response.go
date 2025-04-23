package response

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
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
func SendResponse(w http.ResponseWriter, resp HTTPResponse, traceid string, place string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.Status)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("[ERROR] [UserManagement] [TraceID: %s] %s: Failed to encode response: %v", traceid, place, err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(NewErrorResponse(map[string]string{
			"InternalServerError": "EncoderResponse Error",
		}, resp.Status))
	}
	log.Printf("[INFO] [UserManagement] [TraceID: %s] %s: Succesfull send response to client", traceid, place)
}
func AddSessionCookie(w http.ResponseWriter, sessionID string, expireTime time.Time) {
	maxAge := int(time.Until(expireTime).Seconds())
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    sessionID,
		Path:     "/",
		HttpOnly: true,
		Secure:   false, //временное использование так как пока протокол http
		SameSite: http.SameSiteStrictMode,
		MaxAge:   maxAge,
	})
}
func DeleteSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   false, //временное использование так как пока протокол http
		SameSite: http.SameSiteStrictMode,
		MaxAge:   -1,
	})
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
