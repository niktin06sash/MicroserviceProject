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
	defer r.Body.Close()
	maparesponse := make(map[string]string)
	traceID := r.Context().Value("traceID").(string)
	if r.Method != http.MethodPost {
		log.Printf("[ERROR] [UserManagement] [TraceID: %s] Registration: Invalid request method(expected Post but it was sent %v)", traceID, r.Method)
		maparesponse["ClientError"] = erro.ErrorNotPost.Error()
		response.SendResponse(r.Context(), w, false, nil, maparesponse, http.StatusBadRequest, traceID, "Registration")
		return
	}
	datafromperson, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("[ERROR] [UserManagement] [TraceID: %s] Registration: ReadAll Error: %v", traceID, err)
		maparesponse["ClientError"] = erro.ErrorReadAll.Error()
		response.SendResponse(r.Context(), w, false, nil, maparesponse, http.StatusBadRequest, traceID, "Registration")
		return
	}
	var newperk model.Person
	err = json.Unmarshal(datafromperson, &newperk)
	if err != nil {
		log.Printf("[ERROR] [UserManagement] [TraceID: %s] Registration: Unmarshal Error: %v", traceID, err)
		maparesponse["ClientError"] = erro.ErrorUnmarshal.Error()
		response.SendResponse(r.Context(), w, false, nil, maparesponse, http.StatusBadRequest, traceID, "Registration")
		return
	}
	regresponse := h.Services.RegistrateAndLogin(r.Context(), &newperk)
	if !regresponse.Success {
		stringMap := response.ConvertErrorsToString(regresponse.Errors)
		switch regresponse.Type {
		case erro.ClientErrorType:
			response.SendResponse(r.Context(), w, false, nil, stringMap, http.StatusBadRequest, traceID, "Registration")
		case erro.ServerErrorType:
			response.SendResponse(r.Context(), w, false, nil, stringMap, http.StatusInternalServerError, traceID, "Registration")
		default:
			log.Printf("[ERROR] [UserManagement] [TraceID: %s] Registration: Unknown error type", traceID)
			maparesponse["InternalServerError"] = erro.ErrorInternalServer.Error()
			response.SendResponse(r.Context(), w, false, nil, maparesponse, http.StatusInternalServerError, traceID, "Registration")
		}
		return
	}

	respdata := map[string]any{"UserID": regresponse.UserId}
	response.AddSessionCookie(w, regresponse.SessionId, regresponse.ExpireSession)
	response.SendResponse(r.Context(), w, true, respdata, nil, http.StatusOK, traceID, "Registration")
	log.Printf("[INFO] [UserManagement] [TraceID: %s] Registration: Person with id: %v has successfully registered", traceID, regresponse.UserId)
}

func (h *Handler) Authentication(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	maparesponse := make(map[string]string)
	traceID := r.Context().Value("traceID").(string)
	if r.Method != http.MethodPost {
		log.Printf("[ERROR] [UserManagement] [TraceID: %s] Authentication: Invalid request method(expected Post but it was sent %v)", traceID, r.Method)
		maparesponse["ClientError"] = erro.ErrorNotPost.Error()
		response.SendResponse(r.Context(), w, false, nil, maparesponse, http.StatusBadRequest, traceID, "Authentication")
		return
	}
	datafromperson, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("[ERROR] [UserManagement] [TraceID: %s] Authentication: ReadAll Error: %v", traceID, err)
		maparesponse["ClientError"] = erro.ErrorReadAll.Error()
		response.SendResponse(r.Context(), w, false, nil, maparesponse, http.StatusBadRequest, traceID, "Authentication")
		return
	}
	var newperk model.Person
	err = json.Unmarshal(datafromperson, &newperk)
	if err != nil {
		log.Printf("[ERROR] [UserManagement] [TraceID: %s] Authentication: Unmarshal Error: %v", traceID, err)
		maparesponse["ClientError"] = erro.ErrorUnmarshal.Error()
		response.SendResponse(r.Context(), w, false, nil, maparesponse, http.StatusBadRequest, traceID, "Authentication")
		return
	}
	auresponse := h.Services.AuthenticateAndLogin(r.Context(), &newperk)
	if !auresponse.Success {
		stringMap := response.ConvertErrorsToString(auresponse.Errors)
		switch auresponse.Type {
		case erro.ClientErrorType:
			response.SendResponse(r.Context(), w, false, nil, stringMap, http.StatusBadRequest, traceID, "Authentication")
		case erro.ServerErrorType:
			response.SendResponse(r.Context(), w, false, nil, stringMap, http.StatusInternalServerError, traceID, "Authentication")
		default:
			log.Printf("[ERROR] [UserManagement] [TraceID: %s] Authentication: Unknown error type", traceID)
			maparesponse["InternalServerError"] = erro.ErrorInternalServer.Error()
			response.SendResponse(r.Context(), w, false, nil, maparesponse, http.StatusInternalServerError, traceID, "Authentication")
		}
		return
	}

	respdata := map[string]any{"UserID": auresponse.UserId}
	response.AddSessionCookie(w, auresponse.SessionId, auresponse.ExpireSession)
	response.SendResponse(r.Context(), w, true, respdata, nil, http.StatusOK, traceID, "Authentication")
	log.Printf("[INFO] [UserManagement] [TraceID: %s] Authentication: Person with id: %v has successfully authenticated", traceID, auresponse.UserId)
}

func (h *Handler) DeleteAccount(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	maparesponse := make(map[string]string)
	traceID := r.Context().Value("traceID").(string)
	sessionID, ok := r.Context().Value("sessionID").(string)
	if !ok {
		log.Printf("[ERROR] [UserManagement] [TraceID: %s] DeleteAccount: Session ID not found in context", traceID)
		maparesponse["InternalServerError"] = erro.ErrorMissingSessionID.Error()
		response.SendResponse(r.Context(), w, false, nil, maparesponse, http.StatusInternalServerError, traceID, "DeleteAccount")
		return
	}
	userIDStr, ok := r.Context().Value("userID").(string)
	if !ok {
		log.Printf("[ERROR] [UserManagement] [TraceID: %s] DeleteAccount: User ID not found in context", traceID)
		maparesponse["InternalServerError"] = erro.ErrorMissingUserID.Error()
		response.SendResponse(r.Context(), w, false, nil, maparesponse, http.StatusInternalServerError, traceID, "DeleteAccount")
		return
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		log.Printf("[ERROR] [UserManagement] [TraceID: %s] DeleteAccount: Invalid User ID format: %v", traceID, err)
		maparesponse["InternalServerError"] = erro.ErrorMissingUserID.Error()
		response.SendResponse(r.Context(), w, false, nil, maparesponse, http.StatusInternalServerError, traceID, "DeleteAccount")
		return
	}
	if r.Method != http.MethodDelete {
		log.Printf("[ERROR] [UserManagement] [TraceID: %s] DeleteAccount: Invalid request method(expected Delete but it was sent %v)", traceID, r.Method)
		maparesponse["ClientError"] = erro.ErrorNotDelete.Error()
		response.SendResponse(r.Context(), w, false, nil, maparesponse, http.StatusBadRequest, traceID, "DeleteAccount")
		return
	}
	passwordBytes, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("[ERROR] [UserManagement] [TraceID: %s] DeleteAccount: ReadAll Error: %v", traceID, err)
		maparesponse["ClientError"] = erro.ErrorReadAll.Error()
		response.SendResponse(r.Context(), w, false, nil, maparesponse, http.StatusBadRequest, traceID, "DeleteAccount")
		return
	}
	var data map[string]string
	if err := json.Unmarshal(passwordBytes, &data); err != nil {
		log.Printf("[ERROR] [UserManagement] [TraceID: %s] DeleteAccount: Invalid password format or empty password", traceID)
		maparesponse["ClientError"] = erro.ErrorUnmarshal.Error()
		response.SendResponse(r.Context(), w, false, nil, maparesponse, http.StatusBadRequest, traceID, "DeleteAccount")
		return
	}
	password, ok := data["password"]
	if !ok || password == "" {
		log.Printf("[ERROR] [UserManagement] [TraceID: %s] DeleteAccount: Password is missing or empty", traceID)
		maparesponse["ClientError"] = erro.ErrorUnmarshal.Error()
		response.SendResponse(r.Context(), w, false, nil, maparesponse, http.StatusBadRequest, traceID, "DeleteAccount")
		return
	}
	defer r.Body.Close()
	delresponse := h.Services.DeleteAccount(r.Context(), sessionID, userID, string(password))
	if !delresponse.Success {
		stringMap := response.ConvertErrorsToString(delresponse.Errors)
		switch delresponse.Type {
		case erro.ClientErrorType:
			response.SendResponse(r.Context(), w, false, nil, stringMap, http.StatusBadRequest, traceID, "DeleteAccount")
		case erro.ServerErrorType:
			response.SendResponse(r.Context(), w, false, nil, stringMap, http.StatusInternalServerError, traceID, "DeleteAccount")
		default:
			log.Printf("[ERROR] [UserManagement] [TraceID: %s] DeleteAccount: Unknown error type", traceID)
			maparesponse["InternalServerError"] = erro.ErrorInternalServer.Error()
			response.SendResponse(r.Context(), w, false, nil, maparesponse, http.StatusInternalServerError, traceID, "DeleteAccount")
		}
		return
	}
	response.DeleteSessionCookie(w)
	response.SendResponse(r.Context(), w, true, nil, nil, http.StatusOK, traceID, "DeleteAccount")
	log.Printf("[INFO] [UserManagement] [TraceID: %s] DeleteAccount: Person with id: %v has successfully delete account with all data", traceID, userID)
}
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	maparesponse := make(map[string]string)
	traceID := r.Context().Value("traceID").(string)
	sessionID, ok := r.Context().Value("sessionID").(string)
	if !ok {
		log.Printf("[ERROR] [UserManagement] [TraceID: %s] Logout: Session ID not found in context", traceID)
		maparesponse["InternalServerError"] = erro.ErrorMissingSessionID.Error()
		response.SendResponse(r.Context(), w, false, nil, maparesponse, http.StatusInternalServerError, traceID, "Logout")
		return
	}
	userIDStr, ok := r.Context().Value("userID").(string)
	if !ok {
		log.Printf("[ERROR] [UserManagement] [TraceID: %s] Logout: User ID not found in context", traceID)
		maparesponse["InternalServerError"] = erro.ErrorMissingUserID.Error()
		response.SendResponse(r.Context(), w, false, nil, maparesponse, http.StatusInternalServerError, traceID, "Logout")
		return
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		log.Printf("[ERROR] [UserManagement] [TraceID: %s] Logout: Invalid User ID format: %v", traceID, err)
		maparesponse["InternalServerError"] = erro.ErrorMissingUserID.Error()
		response.SendResponse(r.Context(), w, false, nil, maparesponse, http.StatusInternalServerError, traceID, "Logout")
		return
	}
	if r.Method != http.MethodDelete {
		log.Printf("[ERROR] [UserManagement] [TraceID: %s] Logout: Invalid request method(expected Delete but it was sent %v)", traceID, r.Method)
		maparesponse["ClientError"] = erro.ErrorNotDelete.Error()
		response.SendResponse(r.Context(), w, false, nil, maparesponse, http.StatusBadRequest, traceID, "Logout")
		return
	}
	logresponse := h.Services.Logout(r.Context(), sessionID)
	if !logresponse.Success {
		stringMap := response.ConvertErrorsToString(logresponse.Errors)
		switch logresponse.Type {
		case erro.ClientErrorType:
			response.SendResponse(r.Context(), w, false, nil, stringMap, http.StatusBadRequest, traceID, "Logout")
		case erro.ServerErrorType:
			response.SendResponse(r.Context(), w, false, nil, stringMap, http.StatusInternalServerError, traceID, "Logout")
		default:
			log.Printf("[ERROR] [UserManagement] [TraceID: %s] Logout: Unknown error type", traceID)
			maparesponse["InternalServerError"] = erro.ErrorInternalServer.Error()
			response.SendResponse(r.Context(), w, false, nil, maparesponse, http.StatusInternalServerError, traceID, "Logout")
		}
		return
	}
	response.DeleteSessionCookie(w)
	response.SendResponse(r.Context(), w, true, nil, nil, http.StatusOK, traceID, "Logout")
	log.Printf("[INFO] [UserManagement] [TraceID: %s] Logout: Person with id: %v has successfully logged out of the account", traceID, userID)
}
