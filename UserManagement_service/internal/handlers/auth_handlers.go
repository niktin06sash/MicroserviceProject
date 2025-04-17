package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/erro"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/model"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/service"

	"github.com/google/uuid"
)

func (h *Handler) Registration(w http.ResponseWriter, r *http.Request) {
	maparesponse := make(map[string]string)
	requestID, ok := r.Context().Value("requestID").(string)
	if !ok {
		log.Println("[ERROR] [UserManagement] Registration: Request ID not found in context")
		maparesponse["RequestId"] = erro.ErrorMissingRequestID.Error()
		badResponse(w, maparesponse, http.StatusInternalServerError)
		return
	}

	if r.Method != http.MethodPost {
		log.Printf("[ERROR] [UserManagement] [RequestID: %s] Registration: Invalid request method(expected Post but it was sent %v)", requestID, r.Method)
		maparesponse["Method"] = erro.ErrorNotPost.Error()
		badResponse(w, maparesponse, http.StatusMethodNotAllowed)

		return
	}
	datafromperson, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("[ERROR] [UserManagement] [RequestID: %s] Registration: ReadAll Error: %v", requestID, err)
		maparesponse["ReadAll"] = erro.ErrorReadAll.Error()
		badResponse(w, maparesponse, http.StatusBadRequest)

		return
	}
	var newperk model.Person
	err = json.Unmarshal(datafromperson, &newperk)
	if err != nil {
		log.Printf("[ERROR] [UserManagement] [RequestID: %s] Registration: Unmarshal Error: %v", requestID, err)
		maparesponse["Unmarshal"] = erro.ErrorUnmarshal.Error()
		badResponse(w, maparesponse, http.StatusInternalServerError)

		return
	}
	regresponse := h.services.RegistrateAndLogin(r.Context(), &newperk)
	if !regresponse.Success {
		stringMap := convertErrorToString(regresponse)
		badResponse(w, stringMap, http.StatusBadRequest)

		return
	}

	w.Header().Set("Content-Type", jsonResponseType)
	w.WriteHeader(http.StatusOK)
	sucresponse := HTTPResponse{
		Success:       true,
		SessionId:     regresponse.SessionId,
		ExpireSession: regresponse.ExpireSession,
	}
	jsonResponse, err := json.Marshal(sucresponse)
	if err != nil {
		log.Printf("Marshal Error: %v", err)
		maparesponse["Marshal"] = erro.ErrorMarshal.Error()
		badResponse(w, maparesponse, http.StatusInternalServerError)

		return
	}
	log.Printf("[INFO] [UserManagement] [RequestID: %s] Registration: Person with id: %v has successfully registered", requestID, regresponse.UserId)

	fmt.Fprint(w, string(jsonResponse))
}

func (h *Handler) Authentication(w http.ResponseWriter, r *http.Request) {
	maparesponse := make(map[string]string)
	requestID, ok := r.Context().Value("requestID").(string)
	if !ok {
		log.Println("[ERROR] [UserManagement] Authentication: Request ID not found in context")
		maparesponse["RequestId"] = erro.ErrorMissingRequestID.Error()
		badResponse(w, maparesponse, http.StatusInternalServerError)
		return
	}

	if r.Method != http.MethodPost {
		log.Printf("[ERROR] [UserManagement] [RequestID: %s] Authentication: Invalid request method(expected Post but it was sent %v)", requestID, r.Method)
		maparesponse["Method"] = erro.ErrorNotPost.Error()
		badResponse(w, maparesponse, http.StatusMethodNotAllowed)

		return
	}
	datafromperson, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("[ERROR] [UserManagement] [RequestID: %s] Authentication: ReadAll Error: %v", requestID, err)
		maparesponse["ReadAll"] = erro.ErrorReadAll.Error()
		badResponse(w, maparesponse, http.StatusBadRequest)

		return
	}
	var newperk model.Person
	err = json.Unmarshal(datafromperson, &newperk)
	if err != nil {
		log.Printf("[ERROR] [UserManagement] [RequestID: %s] Authentication: Unmarshal Error: %v", requestID, err)
		maparesponse["Unmarshal"] = erro.ErrorUnmarshal.Error()
		badResponse(w, maparesponse, http.StatusInternalServerError)

		return
	}
	auresponse := h.services.AuthenticateAndLogin(r.Context(), &newperk)
	if !auresponse.Success {
		stringMap := convertErrorToString(auresponse)
		badResponse(w, stringMap, http.StatusBadRequest)

		return
	}

	w.Header().Set("Content-Type", jsonResponseType)
	w.WriteHeader(http.StatusOK)
	sucresponse := HTTPResponse{
		Success:       true,
		SessionId:     auresponse.SessionId,
		ExpireSession: auresponse.ExpireSession,
	}
	jsonResponse, err := json.Marshal(sucresponse)
	if err != nil {
		log.Printf("[ERROR] [UserManagement] [RequestID: %s] Authentication: Marshal Error: %v", requestID, err)
		maparesponse["Marshal"] = erro.ErrorMarshal.Error()
		badResponse(w, maparesponse, http.StatusInternalServerError)

		return
	}
	log.Printf("[INFO] [UserManagement] [RequestID: %s] Authentication: Person with id: %v has successfully authenticated", requestID, auresponse.UserId)

	fmt.Fprint(w, string(jsonResponse))

}

func (h *Handler) DeleteAccount(w http.ResponseWriter, r *http.Request) {
	maparesponse := make(map[string]string)
	requestID, ok := r.Context().Value("requestID").(string)
	if !ok {
		log.Println("[ERROR] [UserManagement] DeleteAccount: Request ID not found in context")
		maparesponse["RequestId"] = erro.ErrorMissingRequestID.Error()
		badResponse(w, maparesponse, http.StatusInternalServerError)
		return
	}
	sessionID, ok := r.Context().Value("sessionID").(string)
	if !ok {
		log.Printf("[ERROR] [UserManagement] [RequestID: %s] DeleteAccount: Session ID not found in context", requestID)
		maparesponse["SessionId"] = erro.ErrorMissingSessionID.Error()
		badResponse(w, maparesponse, http.StatusInternalServerError)
		return
	}

	userID, ok := r.Context().Value("userID").(uuid.UUID)
	if !ok {
		log.Printf("[ERROR] [UserManagement] [RequestID: %s] DeleteAccount: User ID not found in context", requestID)
		maparesponse["UserId"] = erro.ErrorMissingUserID.Error()
		badResponse(w, maparesponse, http.StatusInternalServerError)
		return
	}
	if r.Method != http.MethodDelete {
		log.Printf("[ERROR] [UserManagement] [RequestID: %s] DeleteAccount: Invalid request method(expected Delete but it was sent %v)", requestID, r.Method)
		maparesponse["Method"] = erro.ErrorNotDelete.Error()
		badResponse(w, maparesponse, http.StatusMethodNotAllowed)
		return
	}
	passwordBytes, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("[ERROR] [UserManagement] [RequestID: %s] DeleteAccount: ReadAll Error: %v", requestID, err)
		maparesponse["ReadAll"] = erro.ErrorReadAll.Error()
		badResponse(w, maparesponse, http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var password string
	if err := json.Unmarshal(passwordBytes, &password); err != nil || password == "" {
		log.Printf("[ERROR] [UserManagement] [RequestID: %s] DeleteAccount: Invalid password format or empty password", requestID)
		maparesponse["Password"] = erro.ErrorInvalidPassword.Error()
		badResponse(w, maparesponse, http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	response := h.services.DeleteAccount(r.Context(), sessionID, userID, string(password))
	if !response.Success {
		stringMap := convertErrorToString(response)
		badResponse(w, stringMap, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", jsonResponseType)
	w.WriteHeader(http.StatusOK)
	sucresponse := HTTPResponse{
		Success: true,
	}
	jsonResponse, err := json.Marshal(sucresponse)
	if err != nil {
		log.Printf("[ERROR] [UserManagement] [RequestID: %s] DeleteAccount: Marshal Error: %v", requestID, err)
		maparesponse["Marshal"] = erro.ErrorMarshal.Error()
		badResponse(w, maparesponse, http.StatusInternalServerError)
		return
	}
	log.Printf("[INFO] [UserManagement] [RequestID: %s] DeleteAccount: Person with id: %v has successfully delete account with all data", requestID, userID)

	fmt.Fprint(w, string(jsonResponse))

}
func badResponse(w http.ResponseWriter, vc map[string]string, statusCode int) {
	response := HTTPResponse{
		Success: false,
		Errors:  vc,
	}
	jsonResponse, err := json.Marshal(response)
	if err != nil {
		log.Printf("[ERROR] [UserManagement] Marshal Error: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", jsonResponseType)
	w.WriteHeader(statusCode)

	_, err = w.Write(jsonResponse)
	if err != nil {
		log.Printf("[ERROR] [UserManagement] Write Error: %v", err)
		return
	}
}
func convertErrorToString(mapa *service.ServiceResponse) map[string]string {
	stringMap := make(map[string]string)
	for key, err := range mapa.Errors {
		if err != nil {
			stringMap[key] = err.Error()
		} else {
			stringMap[key] = "no error"
		}
	}
	return stringMap
}
