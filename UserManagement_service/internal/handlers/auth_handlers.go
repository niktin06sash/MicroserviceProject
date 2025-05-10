package handlers

import (
	"fmt"
	"log"
	"net/http"

	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/erro"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/handlers/response"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/kafka"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/model"

	"github.com/google/uuid"
)

func (h *Handler) Registration(w http.ResponseWriter, r *http.Request) {
	var place = "Registration"
	defer r.Body.Close()
	maparesponse := make(map[string]string)
	traceID := r.Context().Value("traceID").(string)
	if !checkMethod(r, w, http.MethodPost, traceID, maparesponse, place, h.KafkaProducer) {
		return
	}
	var newperk model.Person
	if !getAllData(r, w, traceID, maparesponse, place, &newperk, h.KafkaProducer) {
		return
	}
	regresponse := h.Services.RegistrateAndLogin(r.Context(), &newperk)
	if !serviceResponse(regresponse, r, w, traceID, place) {
		return
	}
	respdata := map[string]any{"UserID": regresponse.UserId}
	response.AddSessionCookie(w, regresponse.SessionId, regresponse.ExpireSession)
	response.SendResponse(r.Context(), w, true, respdata, nil, http.StatusOK, traceID, place)
	msg := fmt.Sprintf("Person with id %v has successfully registered", regresponse.UserId)
	h.KafkaProducer.NewUserLog(kafka.LogLevelInfo, place, traceID, msg)
}

func (h *Handler) Authentication(w http.ResponseWriter, r *http.Request) {
	var place = "Authentication"
	defer r.Body.Close()
	maparesponse := make(map[string]string)
	traceID := r.Context().Value("traceID").(string)
	if !checkMethod(r, w, http.MethodPost, traceID, maparesponse, place, h.KafkaProducer) {
		return
	}
	var newperk model.Person
	if !getAllData(r, w, traceID, maparesponse, place, &newperk, h.KafkaProducer) {
		return
	}
	auresponse := h.Services.AuthenticateAndLogin(r.Context(), &newperk)
	if !serviceResponse(auresponse, r, w, traceID, place) {
		return
	}
	respdata := map[string]any{"UserID": auresponse.UserId}
	response.AddSessionCookie(w, auresponse.SessionId, auresponse.ExpireSession)
	response.SendResponse(r.Context(), w, true, respdata, nil, http.StatusOK, traceID, place)
	msg := fmt.Sprintf("Person with id %v has successfully registered", auresponse.UserId)
	h.KafkaProducer.NewUserLog(kafka.LogLevelInfo, place, traceID, msg)
}

func (h *Handler) DeleteAccount(w http.ResponseWriter, r *http.Request) {
	var place = "DeleteAccount"
	defer r.Body.Close()
	maparesponse := make(map[string]string)
	traceID := r.Context().Value("traceID").(string)
	flag, sessionID, userIDStr := getUserIdAndSession(r, w, traceID, maparesponse, place, h.KafkaProducer)
	if !flag {
		return
	}
	//сделать перевод в uuid в бизнес-логике - добавить в другой слой
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		log.Printf("[ERROR] [UserManagement] [TraceID: %s] DeleteAccount: Invalid User ID format: %v", traceID, err)
		maparesponse["InternalServerError"] = erro.ErrorMissingUserID.Error()
		response.SendResponse(r.Context(), w, false, nil, maparesponse, http.StatusInternalServerError, traceID, "DeleteAccount")
		return
	}
	if !checkMethod(r, w, http.MethodDelete, traceID, maparesponse, place, h.KafkaProducer) {
		return
	}
	data := make(map[string]string)
	if !getAllData(r, w, traceID, maparesponse, place, data, h.KafkaProducer) {
		return
	}
	password, ok := data["password"]
	if !ok || password == "" {
		h.KafkaProducer.NewUserLog(kafka.LogLevelError, place, traceID, "Password is missing or empty")
		maparesponse["ClientError"] = erro.ErrorUnmarshal.Error()
		response.SendResponse(r.Context(), w, false, nil, maparesponse, http.StatusBadRequest, traceID, place)
		return
	}
	defer r.Body.Close()
	delresponse := h.Services.DeleteAccount(r.Context(), sessionID, userID, string(password))
	if !serviceResponse(delresponse, r, w, traceID, place) {
		return
	}
	response.DeleteSessionCookie(w)
	response.SendResponse(r.Context(), w, true, nil, nil, http.StatusOK, traceID, place)
	msg := fmt.Sprintf("Person with id %v has successfully registered", delresponse.UserId)
	h.KafkaProducer.NewUserLog(kafka.LogLevelInfo, place, traceID, msg)
}
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	var place = "Logout"
	defer r.Body.Close()
	maparesponse := make(map[string]string)
	traceID := r.Context().Value("traceID").(string)
	flag, sessionID, userIDStr := getUserIdAndSession(r, w, traceID, maparesponse, place, h.KafkaProducer)
	if !flag {
		return
	}
	if !checkMethod(r, w, http.MethodDelete, traceID, maparesponse, place, h.KafkaProducer) {
		return
	}
	logresponse := h.Services.Logout(r.Context(), sessionID)
	if !serviceResponse(logresponse, r, w, traceID, place) {
		return
	}
	response.DeleteSessionCookie(w)
	response.SendResponse(r.Context(), w, true, nil, nil, http.StatusOK, traceID, "Logout")
	msg := fmt.Sprintf("Person with id %v has successfully logout", userIDStr)
	h.KafkaProducer.NewUserLog(kafka.LogLevelInfo, place, traceID, msg)
}
