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
	traceID := r.Context().Value("traceID").(string)
	if r.Method != http.MethodPost {
		log.Printf("[ERROR] [UserManagement] [TraceID: %s] Registration: Invalid request method(expected Post but it was sent %v)", traceID, r.Method)
		maparesponse["ClientError"] = erro.ErrorNotPost.Error()
		br := response.NewErrorResponse(maparesponse, http.StatusBadRequest)
		response.SendResponse(w, br, traceID, "Registration", r.Context())
		return
	}
	datafromperson, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("[ERROR] [UserManagement] [TraceID: %s] Registration: ReadAll Error: %v", traceID, err)
		maparesponse["ClientError"] = erro.ErrorReadAll.Error()
		br := response.NewErrorResponse(maparesponse, http.StatusBadRequest)
		response.SendResponse(w, br, traceID, "Registration", r.Context())
		return
	}
	var newperk model.Person
	err = json.Unmarshal(datafromperson, &newperk)
	if err != nil {
		log.Printf("[ERROR] [UserManagement] [TraceID: %s] Registration: Unmarshal Error: %v", traceID, err)
		maparesponse["ClientError"] = erro.ErrorUnmarshal.Error()
		br := response.NewErrorResponse(maparesponse, http.StatusBadRequest)
		response.SendResponse(w, br, traceID, "Registration", r.Context())
		return
	}
	regresponse := h.services.RegistrateAndLogin(r.Context(), &newperk)
	if !regresponse.Success {
		stringMap := response.ConvertErrorsToString(regresponse.Errors)
		switch regresponse.Type {
		case erro.ClientErrorType:
			br := response.NewErrorResponse(stringMap, http.StatusBadRequest)
			response.SendResponse(w, br, traceID, "Registration", r.Context())
		case erro.ServerErrorType:
			br := response.NewErrorResponse(stringMap, http.StatusInternalServerError)
			response.SendResponse(w, br, traceID, "Registration", r.Context())
		default:
			log.Printf("[ERROR] [UserManagement] [TraceID: %s] Registration: Unknown error type", traceID)
			maparesponse["InternalServerError"] = erro.ErrorInternalServer.Error()
			br := response.NewErrorResponse(maparesponse, http.StatusInternalServerError)
			response.SendResponse(w, br, traceID, "Registration", r.Context())
		}
		return
	}

	respdata := map[string]any{"UserID": regresponse.UserId}
	response.AddSessionCookie(w, regresponse.SessionId, regresponse.ExpireSession)
	br := response.NewSuccessResponse(respdata, http.StatusOK)
	response.SendResponse(w, br, traceID, "Registration", r.Context())
	log.Printf("[INFO] [UserManagement] [TraceID: %s] Registration: Person with id: %v has successfully registered", traceID, regresponse.UserId)
}

func (h *Handler) Authentication(w http.ResponseWriter, r *http.Request) {
	maparesponse := make(map[string]string)
	traceID := r.Context().Value("traceID").(string)
	if r.Method != http.MethodPost {
		log.Printf("[ERROR] [UserManagement] [TraceID: %s] Authentication: Invalid request method(expected Post but it was sent %v)", traceID, r.Method)
		maparesponse["ClientError"] = erro.ErrorNotPost.Error()
		br := response.NewErrorResponse(maparesponse, http.StatusBadRequest)
		response.SendResponse(w, br, traceID, "Authentication", r.Context())
		return
	}
	datafromperson, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("[ERROR] [UserManagement] [TraceID: %s] Authentication: ReadAll Error: %v", traceID, err)
		maparesponse["ClientError"] = erro.ErrorReadAll.Error()
		br := response.NewErrorResponse(maparesponse, http.StatusBadRequest)
		response.SendResponse(w, br, traceID, "Authentication", r.Context())
		return
	}
	var newperk model.Person
	err = json.Unmarshal(datafromperson, &newperk)
	if err != nil {
		log.Printf("[ERROR] [UserManagement] [TraceID: %s] Authentication: Unmarshal Error: %v", traceID, err)
		maparesponse["ClientError"] = erro.ErrorUnmarshal.Error()
		br := response.NewErrorResponse(maparesponse, http.StatusBadRequest)
		response.SendResponse(w, br, traceID, "Authentication", r.Context())
		return
	}
	auresponse := h.services.AuthenticateAndLogin(r.Context(), &newperk)
	if !auresponse.Success {
		stringMap := response.ConvertErrorsToString(auresponse.Errors)
		switch auresponse.Type {
		case erro.ClientErrorType:
			br := response.NewErrorResponse(stringMap, http.StatusBadRequest)
			response.SendResponse(w, br, traceID, "Authentication", r.Context())
		case erro.ServerErrorType:
			br := response.NewErrorResponse(stringMap, http.StatusInternalServerError)
			response.SendResponse(w, br, traceID, "Authentication", r.Context())
		default:
			log.Printf("[ERROR] [UserManagement] [TraceID: %s] Authentication: Unknown error type", traceID)
			maparesponse["InternalServerError"] = erro.ErrorInternalServer.Error()
			br := response.NewErrorResponse(maparesponse, http.StatusInternalServerError)
			response.SendResponse(w, br, traceID, "Authentication", r.Context())
		}
		return
	}

	respdata := map[string]any{"UserID": auresponse.UserId}
	response.AddSessionCookie(w, auresponse.SessionId, auresponse.ExpireSession)
	br := response.NewSuccessResponse(respdata, http.StatusOK)
	response.SendResponse(w, br, traceID, "Authentication", r.Context())
	log.Printf("[INFO] [UserManagement] [TraceID: %s] Authentication: Person with id: %v has successfully authenticated", traceID, auresponse.UserId)
}

func (h *Handler) DeleteAccount(w http.ResponseWriter, r *http.Request) {
	maparesponse := make(map[string]string)
	traceID := r.Context().Value("traceID").(string)
	sessionID, ok := r.Context().Value("sessionID").(string)
	if !ok {
		log.Printf("[ERROR] [UserManagement] [TraceID: %s] DeleteAccount: Session ID not found in context", traceID)
		maparesponse["InternalServerError"] = erro.ErrorMissingSessionID.Error()
		br := response.NewErrorResponse(maparesponse, http.StatusInternalServerError)
		response.SendResponse(w, br, traceID, "DeleteAccount", r.Context())
		return
	}
	userIDStr, ok := r.Context().Value("userID").(string)
	if !ok {
		log.Printf("[ERROR] [UserManagement] [TraceID: %s] DeleteAccount: User ID not found in context", traceID)
		maparesponse["InternalServerError"] = erro.ErrorMissingUserID.Error()
		br := response.NewErrorResponse(maparesponse, http.StatusInternalServerError)
		response.SendResponse(w, br, traceID, "DeleteAccount", r.Context())
		return
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		log.Printf("[ERROR] [UserManagement] [TraceID: %s] DeleteAccount: Invalid User ID format: %v", traceID, err)
		maparesponse["InternalServerError"] = erro.ErrorMissingUserID.Error()
		br := response.NewErrorResponse(maparesponse, http.StatusInternalServerError)
		response.SendResponse(w, br, traceID, "DeleteAccount", r.Context())
		return
	}
	if r.Method != http.MethodDelete {
		log.Printf("[ERROR] [UserManagement] [TraceID: %s] DeleteAccount: Invalid request method(expected Delete but it was sent %v)", traceID, r.Method)
		maparesponse["ClientError"] = erro.ErrorNotDelete.Error()
		br := response.NewErrorResponse(maparesponse, http.StatusBadRequest)
		response.SendResponse(w, br, traceID, "DeleteAccount", r.Context())
		return
	}
	passwordBytes, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("[ERROR] [UserManagement] [TraceID: %s] DeleteAccount: ReadAll Error: %v", traceID, err)
		maparesponse["ClientError"] = erro.ErrorReadAll.Error()
		br := response.NewErrorResponse(maparesponse, http.StatusBadRequest)
		response.SendResponse(w, br, traceID, "DeleteAccount", r.Context())
		return
	}
	defer r.Body.Close()
	var data map[string]string
	if err := json.Unmarshal(passwordBytes, &data); err != nil {
		log.Printf("[ERROR] [UserManagement] [TraceID: %s] DeleteAccount: Invalid password format or empty password", traceID)
		maparesponse["ClientError"] = erro.ErrorInvalidPassword.Error()
		br := response.NewErrorResponse(maparesponse, http.StatusBadRequest)
		response.SendResponse(w, br, traceID, "DeleteAccount", r.Context())
		return
	}
	password, ok := data["password"]
	if !ok || password == "" {
		log.Printf("[ERROR] [UserManagement] [TraceID: %s] DeleteAccount: Password is missing or empty", traceID)
		maparesponse["ClientError"] = erro.ErrorInvalidPassword.Error()
		br := response.NewErrorResponse(maparesponse, http.StatusBadRequest)
		response.SendResponse(w, br, traceID, "DeleteAccount", r.Context())
		return
	}
	defer r.Body.Close()
	delresponse := h.services.DeleteAccount(r.Context(), sessionID, userID, string(password))
	if !delresponse.Success {
		stringMap := response.ConvertErrorsToString(delresponse.Errors)
		switch delresponse.Type {
		case erro.ClientErrorType:
			br := response.NewErrorResponse(stringMap, http.StatusBadRequest)
			response.SendResponse(w, br, traceID, "DeleteAccount", r.Context())
		case erro.ServerErrorType:
			br := response.NewErrorResponse(stringMap, http.StatusInternalServerError)
			response.SendResponse(w, br, traceID, "DeleteAccount", r.Context())
		default:
			log.Printf("[ERROR] [UserManagement] [TraceID: %s] DeleteAccount: Unknown error type", traceID)
			maparesponse["InternalServerError"] = erro.ErrorInternalServer.Error()
			br := response.NewErrorResponse(maparesponse, http.StatusInternalServerError)
			response.SendResponse(w, br, traceID, "DeleteAccount", r.Context())
		}
		return
	}
	br := response.NewSuccessResponse(nil, http.StatusOK)
	response.DeleteSessionCookie(w)
	response.SendResponse(w, br, traceID, "DeleteAccount", r.Context())
	log.Printf("[INFO] [UserManagement] [TraceID: %s] DeleteAccount: Person with id: %v has successfully delete account with all data", traceID, userID)
}
