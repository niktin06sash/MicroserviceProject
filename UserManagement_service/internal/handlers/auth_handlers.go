package handlers

import (
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/erro"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/handlers/response"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/model"

	"github.com/google/uuid"
)

func (h *Handler) Registration(w http.ResponseWriter, r *http.Request) {
	maparesponse := make(map[string]string)
	requestID := r.Context().Value("requestID").(string)
	if r.Method != http.MethodPost {
		log.Printf("[ERROR] [UserManagement] [RequestID: %s] Registration: Invalid request method(expected Post but it was sent %v)", requestID, r.Method)
		maparesponse["Method"] = erro.ErrorNotPost.Error()
		br := response.NewErrorResponse(maparesponse, http.StatusMethodNotAllowed)
		response.SendResponse(w, br)
		return
	}
	datafromperson, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("[ERROR] [UserManagement] [RequestID: %s] Registration: ReadAll Error: %v", requestID, err)
		maparesponse["ReadAll"] = erro.ErrorReadAll.Error()
		br := response.NewErrorResponse(maparesponse, http.StatusBadRequest)
		response.SendResponse(w, br)
		return
	}
	var newperk model.Person
	err = json.Unmarshal(datafromperson, &newperk)
	if err != nil {
		log.Printf("[ERROR] [UserManagement] [RequestID: %s] Registration: Unmarshal Error: %v", requestID, err)
		maparesponse["Unmarshal"] = erro.ErrorUnmarshal.Error()
		br := response.NewErrorResponse(maparesponse, http.StatusBadRequest)
		response.SendResponse(w, br)
		return
	}
	regresponse := h.services.RegistrateAndLogin(r.Context(), &newperk)
	if !regresponse.Success {
		stringMap := response.ConvertErrorsToString(regresponse.Errors)
		switch regresponse.Type {
		case erro.ClientErrorType:
			br := response.NewErrorResponse(stringMap, http.StatusBadRequest)
			response.SendResponse(w, br)
		case erro.ServerErrorType:
			br := response.NewErrorResponse(stringMap, http.StatusInternalServerError)
			response.SendResponse(w, br)
		default:
			log.Printf("[ERROR] [UserManagement] [RequestID: %s] Registration: Unknown error type", requestID)
			maparesponse["InternalServer"] = erro.ErrorInternalServer.Error()
			br := response.NewErrorResponse(maparesponse, http.StatusInternalServerError)
			response.SendResponse(w, br)
		}
		return
	}

	respdata := map[string]any{"SessionID": regresponse.SessionId, "ExpiryTime": regresponse.ExpireSession}
	br := response.NewSuccessResponse(respdata, http.StatusOK)
	response.SendResponse(w, br)
	log.Printf("[INFO] [UserManagement] [RequestID: %s] Registration: Person with id: %v has successfully registered", requestID, regresponse.UserId)
}

func (h *Handler) Authentication(w http.ResponseWriter, r *http.Request) {
	maparesponse := make(map[string]string)
	requestID := r.Context().Value("requestID").(string)
	if r.Method != http.MethodPost {
		log.Printf("[ERROR] [UserManagement] [RequestID: %s] Authentication: Invalid request method(expected Post but it was sent %v)", requestID, r.Method)
		maparesponse["Method"] = erro.ErrorNotPost.Error()
		br := response.NewErrorResponse(maparesponse, http.StatusMethodNotAllowed)
		response.SendResponse(w, br)
		return
	}
	datafromperson, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("[ERROR] [UserManagement] [RequestID: %s] Authentication: ReadAll Error: %v", requestID, err)
		maparesponse["ReadAll"] = erro.ErrorReadAll.Error()
		br := response.NewErrorResponse(maparesponse, http.StatusBadRequest)
		response.SendResponse(w, br)
		return
	}
	var newperk model.Person
	err = json.Unmarshal(datafromperson, &newperk)
	if err != nil {
		log.Printf("[ERROR] [UserManagement] [RequestID: %s] Authentication: Unmarshal Error: %v", requestID, err)
		maparesponse["Unmarshal"] = erro.ErrorUnmarshal.Error()
		br := response.NewErrorResponse(maparesponse, http.StatusInternalServerError)
		response.SendResponse(w, br)
		return
	}
	auresponse := h.services.AuthenticateAndLogin(r.Context(), &newperk)
	if !auresponse.Success {
		stringMap := response.ConvertErrorsToString(auresponse.Errors)
		switch auresponse.Type {
		case erro.ClientErrorType:
			br := response.NewErrorResponse(stringMap, http.StatusBadRequest)
			response.SendResponse(w, br)
		case erro.ServerErrorType:
			br := response.NewErrorResponse(stringMap, http.StatusInternalServerError)
			response.SendResponse(w, br)
		default:
			log.Printf("[ERROR] [UserManagement] [RequestID: %s] Authentication: Unknown error type", requestID)
			maparesponse["InternalServer"] = erro.ErrorInternalServer.Error()
			br := response.NewErrorResponse(maparesponse, http.StatusInternalServerError)
			response.SendResponse(w, br)
		}
		return
	}

	respdata := map[string]any{"SessionID": auresponse.SessionId, "ExpiryTime": auresponse.ExpireSession}
	br := response.NewSuccessResponse(respdata, http.StatusOK)
	response.SendResponse(w, br)
	log.Printf("[INFO] [UserManagement] [RequestID: %s] Authentication: Person with id: %v has successfully authenticated", requestID, auresponse.UserId)
}

func (h *Handler) DeleteAccount(w http.ResponseWriter, r *http.Request) {
	maparesponse := make(map[string]string)
	requestID := r.Context().Value("requestID").(string)
	sessionID, ok := r.Context().Value("sessionID").(string)
	if !ok {
		log.Printf("[ERROR] [UserManagement] [RequestID: %s] DeleteAccount: Session ID not found in context", requestID)
		maparesponse["SessionId"] = erro.ErrorMissingSessionID.Error()
		br := response.NewErrorResponse(maparesponse, http.StatusInternalServerError)
		response.SendResponse(w, br)
		return
	}
	userID, ok := r.Context().Value("userID").(uuid.UUID)
	if !ok {
		log.Printf("[ERROR] [UserManagement] [RequestID: %s] DeleteAccount: User ID not found in context", requestID)
		maparesponse["UserId"] = erro.ErrorMissingUserID.Error()
		br := response.NewErrorResponse(maparesponse, http.StatusInternalServerError)
		response.SendResponse(w, br)
		return
	}
	if r.Method != http.MethodDelete {
		log.Printf("[ERROR] [UserManagement] [RequestID: %s] DeleteAccount: Invalid request method(expected Delete but it was sent %v)", requestID, r.Method)
		maparesponse["Method"] = erro.ErrorNotDelete.Error()
		br := response.NewErrorResponse(maparesponse, http.StatusMethodNotAllowed)
		response.SendResponse(w, br)
		return
	}
	passwordBytes, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("[ERROR] [UserManagement] [RequestID: %s] DeleteAccount: ReadAll Error: %v", requestID, err)
		maparesponse["ReadAll"] = erro.ErrorReadAll.Error()
		br := response.NewErrorResponse(maparesponse, http.StatusBadRequest)
		response.SendResponse(w, br)
		return
	}
	defer r.Body.Close()
	var password string
	if err := json.Unmarshal(passwordBytes, &password); err != nil || password == "" {
		log.Printf("[ERROR] [UserManagement] [RequestID: %s] DeleteAccount: Invalid password format or empty password", requestID)
		maparesponse["Password"] = erro.ErrorInvalidPassword.Error()
		br := response.NewErrorResponse(maparesponse, http.StatusBadRequest)
		response.SendResponse(w, br)
		return
	}
	defer r.Body.Close()
	delresponse := h.services.DeleteAccount(r.Context(), sessionID, userID, string(password))
	if !delresponse.Success {
		stringMap := response.ConvertErrorsToString(delresponse.Errors)
		switch delresponse.Type {
		case erro.ClientErrorType:
			br := response.NewErrorResponse(stringMap, http.StatusBadRequest)
			response.SendResponse(w, br)
		case erro.ServerErrorType:
			br := response.NewErrorResponse(stringMap, http.StatusBadRequest)
			response.SendResponse(w, br)
		default:
			log.Printf("[ERROR] [UserManagement] [RequestID: %s] DeleteAccount: Unknown error type", requestID)
			maparesponse["InternalServer"] = erro.ErrorInternalServer.Error()
			br := response.NewErrorResponse(maparesponse, http.StatusInternalServerError)
			response.SendResponse(w, br)
		}
		return
	}
	br := response.NewSuccessResponse(nil, http.StatusOK)
	response.SendResponse(w, br)
	log.Printf("[INFO] [UserManagement] [RequestID: %s] DeleteAccount: Person with id: %v has successfully delete account with all data", requestID, userID)
}
