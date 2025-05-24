package handlers

import (
	"fmt"
	"net/http"

	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/erro"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/handlers/response"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/kafka"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/model"
)

func (h *Handler) Registration(w http.ResponseWriter, r *http.Request) {
	const place = Registration
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
	if !serviceResponse(regresponse, r, w, traceID, place, h.KafkaProducer) {
		return
	}
	respdata := map[string]any{"UserID": regresponse.UserId}
	response.AddSessionCookie(w, regresponse.SessionId, regresponse.ExpireSession)
	msg := fmt.Sprintf("Person with id %v has successfully registered", regresponse.UserId)
	h.KafkaProducer.NewUserLog(kafka.LogLevelInfo, place, traceID, msg)
	response.SendResponse(r.Context(), w, true, respdata, nil, http.StatusOK, traceID, place, h.KafkaProducer)
}

func (h *Handler) Authentication(w http.ResponseWriter, r *http.Request) {
	const place = Authentication
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
	if !serviceResponse(auresponse, r, w, traceID, place, h.KafkaProducer) {
		return
	}
	respdata := map[string]any{"UserID": auresponse.UserId}
	response.AddSessionCookie(w, auresponse.SessionId, auresponse.ExpireSession)
	msg := fmt.Sprintf("Person with id %v has successfully authenticated", auresponse.UserId)
	h.KafkaProducer.NewUserLog(kafka.LogLevelInfo, place, traceID, msg)
	response.SendResponse(r.Context(), w, true, respdata, nil, http.StatusOK, traceID, place, h.KafkaProducer)
}

func (h *Handler) DeleteAccount(w http.ResponseWriter, r *http.Request) {
	const place = DeleteAccount
	defer r.Body.Close()
	maparesponse := make(map[string]string)
	traceID := r.Context().Value("traceID").(string)
	if !checkMethod(r, w, http.MethodDelete, traceID, maparesponse, place, h.KafkaProducer) {
		return
	}
	flag, persondata := getPersonality(r, w, traceID, maparesponse, place, h.KafkaProducer)
	if !flag {
		return
	}
	data := make(map[string]string)
	if !getAllData(r, w, traceID, maparesponse, place, &data, h.KafkaProducer) {
		return
	}
	password, ok := data["password"]
	if !ok || password == "" {
		h.KafkaProducer.NewUserLog(kafka.LogLevelError, place, traceID, "Password is missing or empty")
		maparesponse["ClientError"] = erro.ErrorUnmarshal.Error()
		response.SendResponse(r.Context(), w, false, nil, maparesponse, http.StatusBadRequest, traceID, place, h.KafkaProducer)
		return
	}
	defer r.Body.Close()
	delresponse := h.Services.DeleteAccount(r.Context(), persondata["sessionID"], persondata["userID"], string(password))
	if !serviceResponse(delresponse, r, w, traceID, place, h.KafkaProducer) {
		return
	}
	response.DeleteSessionCookie(w)
	msg := fmt.Sprintf("Person with id %v has successfully deleted account", persondata["userID"])
	h.KafkaProducer.NewUserLog(kafka.LogLevelInfo, place, traceID, msg)
	response.SendResponse(r.Context(), w, true, nil, nil, http.StatusOK, traceID, place, h.KafkaProducer)
}
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	const place = Logout
	defer r.Body.Close()
	maparesponse := make(map[string]string)
	traceID := r.Context().Value("traceID").(string)
	if !checkMethod(r, w, http.MethodDelete, traceID, maparesponse, place, h.KafkaProducer) {
		return
	}
	flag, persondata := getPersonality(r, w, traceID, maparesponse, place, h.KafkaProducer)
	if !flag {
		return
	}
	logresponse := h.Services.Logout(r.Context(), persondata["sessionID"])
	if !serviceResponse(logresponse, r, w, traceID, place, h.KafkaProducer) {
		return
	}
	response.DeleteSessionCookie(w)
	msg := fmt.Sprintf("Person with id %v has successfully logout", persondata["userID"])
	h.KafkaProducer.NewUserLog(kafka.LogLevelInfo, place, traceID, msg)
	response.SendResponse(r.Context(), w, true, nil, nil, http.StatusOK, traceID, place, h.KafkaProducer)
}
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	const place = Update
	defer r.Body.Close()
	maparesponse := make(map[string]string)
	traceID := r.Context().Value("traceID").(string)
	if !checkMethod(r, w, http.MethodPatch, traceID, maparesponse, place, h.KafkaProducer) {
		return
	}
	flag, persondata := getPersonality(r, w, traceID, maparesponse, place, h.KafkaProducer)
	if !flag {
		return
	}
	data := make(map[string]string)
	if !getAllData(r, w, traceID, maparesponse, place, &data, h.KafkaProducer) {
		return
	}
	updateresponse := h.Services.UpdateAccount(r.Context(), persondata, data["update_type"])
	if !serviceResponse(updateresponse, r, w, traceID, place, h.KafkaProducer) {
		return
	}
	msg := fmt.Sprintf("Person with id %v has successfully update his %s", persondata["userID"], data["update_type"])
	h.KafkaProducer.NewUserLog(kafka.LogLevelInfo, place, traceID, msg)
	response.SendResponse(r.Context(), w, true, nil, nil, http.StatusOK, traceID, place, h.KafkaProducer)
}
