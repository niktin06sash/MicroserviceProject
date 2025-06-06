package handlers

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/brokers/kafka"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/handlers/response"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/model"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/service"
)

type UserService interface {
	RegistrateAndLogin(ctx context.Context, req *model.RegistrationRequest) *service.ServiceResponse
	AuthenticateAndLogin(ctx context.Context, req *model.AuthenticationRequest) *service.ServiceResponse
	DeleteAccount(ctx context.Context, req *model.DeletionRequest, sessionID string, useridstr string) *service.ServiceResponse
	Logout(ctx context.Context, sessionID string) *service.ServiceResponse
	UpdateAccount(ctx context.Context, req *model.UpdateRequest, useridstr string, updateType string) *service.ServiceResponse
	GetMyProfile(ctx context.Context, useridstr string) *service.ServiceResponse
	GetProfileById(ctx context.Context, useridstr string, findidstr string) *service.ServiceResponse
}

func (h *Handler) Registration(w http.ResponseWriter, r *http.Request) {
	const place = Registration
	defer r.Body.Close()
	maparesponse := make(map[string]string)
	traceID := r.Context().Value("traceID").(string)
	if !checkMethod(r, w, http.MethodPost, traceID, place, maparesponse, h.KafkaProducer) {
		return
	}
	var regreq model.RegistrationRequest
	if !getAllData(r, w, traceID, place, maparesponse, &regreq, h.KafkaProducer) {
		return
	}
	regresponse := h.Services.RegistrateAndLogin(r.Context(), &regreq)
	if !serviceResponse(regresponse, r, w, traceID, place, h.KafkaProducer) {
		return
	}
	userID := regresponse.Data["userID"].(string)
	sessionID := regresponse.Data["sessionID"].(string)
	expiresession := regresponse.Data["expiresession"].(time.Time)
	response.AddSessionCookie(w, sessionID, expiresession)
	msg := fmt.Sprintf("Person with id %v has successfully registered", userID)
	h.KafkaProducer.NewUserLog(kafka.LogLevelInfo, place, traceID, msg)
	response.SendResponse(r, w, response.HTTPResponse{Success: true, Data: map[string]any{"message": "You have successfully registered!"}}, http.StatusOK, traceID, place, h.KafkaProducer)
}

func (h *Handler) Authentication(w http.ResponseWriter, r *http.Request) {
	const place = Authentication
	defer r.Body.Close()
	maparesponse := make(map[string]string)
	traceID := r.Context().Value("traceID").(string)
	if !checkMethod(r, w, http.MethodPost, traceID, place, maparesponse, h.KafkaProducer) {
		return
	}
	var aureq model.AuthenticationRequest
	if !getAllData(r, w, traceID, place, maparesponse, &aureq, h.KafkaProducer) {
		return
	}
	auresponse := h.Services.AuthenticateAndLogin(r.Context(), &aureq)
	if !serviceResponse(auresponse, r, w, traceID, place, h.KafkaProducer) {
		return
	}
	userID := auresponse.Data["userID"].(string)
	sessionID := auresponse.Data["sessionID"].(string)
	expiresession := auresponse.Data["expiresession"].(time.Time)
	response.AddSessionCookie(w, sessionID, expiresession)
	msg := fmt.Sprintf("Person with id %v has successfully authenticated", userID)
	h.KafkaProducer.NewUserLog(kafka.LogLevelInfo, place, traceID, msg)
	response.SendResponse(r, w, response.HTTPResponse{Success: true, Data: map[string]any{"message": "You have successfully authenticated!"}}, http.StatusOK, traceID, place, h.KafkaProducer)
}

func (h *Handler) DeleteAccount(w http.ResponseWriter, r *http.Request) {
	const place = DeleteAccount
	defer r.Body.Close()
	maparesponse := make(map[string]string)
	traceID := r.Context().Value("traceID").(string)
	if !checkMethod(r, w, http.MethodDelete, traceID, place, maparesponse, h.KafkaProducer) {
		return
	}
	persondata := make(map[string]string)
	if !getPersonality(r, w, traceID, place, maparesponse, persondata, h.KafkaProducer) {
		return
	}
	var delreq model.DeletionRequest
	if !getAllData(r, w, traceID, place, maparesponse, &delreq, h.KafkaProducer) {
		return
	}
	defer r.Body.Close()
	delresponse := h.Services.DeleteAccount(r.Context(), &delreq, persondata["sessionID"], persondata["userID"])
	if !serviceResponse(delresponse, r, w, traceID, place, h.KafkaProducer) {
		return
	}
	response.DeleteSessionCookie(w)
	msg := fmt.Sprintf("Person with id %v has successfully deleted account", persondata["userID"])
	h.KafkaProducer.NewUserLog(kafka.LogLevelInfo, place, traceID, msg)
	response.SendResponse(r, w, response.HTTPResponse{Success: true, Data: map[string]any{"message": "You have successfully delete account!"}}, http.StatusOK, traceID, place, h.KafkaProducer)
}
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	const place = Logout
	defer r.Body.Close()
	maparesponse := make(map[string]string)
	traceID := r.Context().Value("traceID").(string)
	if !checkMethod(r, w, http.MethodDelete, traceID, place, maparesponse, h.KafkaProducer) {
		return
	}
	persondata := make(map[string]string)
	if !getPersonality(r, w, traceID, place, maparesponse, persondata, h.KafkaProducer) {
		return
	}
	logresponse := h.Services.Logout(r.Context(), persondata["sessionID"])
	if !serviceResponse(logresponse, r, w, traceID, place, h.KafkaProducer) {
		return
	}
	response.DeleteSessionCookie(w)
	msg := fmt.Sprintf("Person with id %v has successfully logout", persondata["userID"])
	h.KafkaProducer.NewUserLog(kafka.LogLevelInfo, place, traceID, msg)
	response.SendResponse(r, w, response.HTTPResponse{Success: true, Data: map[string]any{"message": "You have successfully logout!"}}, http.StatusOK, traceID, place, h.KafkaProducer)
}
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	const place = Update
	defer r.Body.Close()
	maparesponse := make(map[string]string)
	traceID := r.Context().Value("traceID").(string)
	if !checkMethod(r, w, http.MethodPatch, traceID, place, maparesponse, h.KafkaProducer) {
		return
	}
	persondata := make(map[string]string)
	if !getPersonality(r, w, traceID, place, maparesponse, persondata, h.KafkaProducer) {
		return
	}
	var updatereq model.UpdateRequest
	if !getAllData(r, w, traceID, place, maparesponse, &updatereq, h.KafkaProducer) {
		return
	}
	if !getQueryParameters(r, w, traceID, place, maparesponse, persondata, h.KafkaProducer) {
		return
	}
	updateresponse := h.Services.UpdateAccount(r.Context(), &updatereq, persondata["userID"], persondata["update_type"])
	if !serviceResponse(updateresponse, r, w, traceID, place, h.KafkaProducer) {
		return
	}
	msg := fmt.Sprintf("Person with id %v has successfully update his %s", persondata["userID"], persondata["update_type"])
	h.KafkaProducer.NewUserLog(kafka.LogLevelInfo, place, traceID, msg)
	response.SendResponse(r, w, response.HTTPResponse{Success: true, Data: updateresponse.Data}, http.StatusOK, traceID, place, h.KafkaProducer)
}
func (h *Handler) MyProfile(w http.ResponseWriter, r *http.Request) {
	const place = MyProfile
	defer r.Body.Close()
	maparesponse := make(map[string]string)
	traceID := r.Context().Value("traceID").(string)
	if !checkMethod(r, w, http.MethodGet, traceID, place, maparesponse, h.KafkaProducer) {
		return
	}
	persondata := make(map[string]string)
	if !getPersonality(r, w, traceID, place, maparesponse, persondata, h.KafkaProducer) {
		return
	}
	myprofileresponse := h.Services.GetMyProfile(r.Context(), persondata["userID"])
	if !serviceResponse(myprofileresponse, r, w, traceID, place, h.KafkaProducer) {
		return
	}
	msg := fmt.Sprintf("Person with id %v has successfully received his account data", persondata["userID"])
	h.KafkaProducer.NewUserLog(kafka.LogLevelInfo, place, traceID, msg)
	response.SendResponse(r, w, response.HTTPResponse{Success: true, Data: myprofileresponse.Data}, http.StatusOK, traceID, place, h.KafkaProducer)
}
func (h *Handler) GetUserById(w http.ResponseWriter, r *http.Request) {
	const place = GetUserById
	defer r.Body.Close()
	maparesponse := make(map[string]string)
	traceID := r.Context().Value("traceID").(string)
	if !checkMethod(r, w, http.MethodGet, traceID, place, maparesponse, h.KafkaProducer) {
		return
	}
	persondata := make(map[string]string)
	if !getPersonality(r, w, traceID, place, maparesponse, persondata, h.KafkaProducer) {
		return
	}
	if !getDinamicParameters(r, w, traceID, place, maparesponse, persondata, h.KafkaProducer) {
		return
	}
	getprofileresponse := h.Services.GetProfileById(r.Context(), persondata["userID"], persondata["getID"])
	if !serviceResponse(getprofileresponse, r, w, traceID, place, h.KafkaProducer) {
		return
	}
	msg := fmt.Sprintf("Person with id %v has successfully received data of person with id %v", persondata["userID"], persondata["getID"])
	h.KafkaProducer.NewUserLog(kafka.LogLevelInfo, place, traceID, msg)
	response.SendResponse(r, w, response.HTTPResponse{Success: true, Data: getprofileresponse.Data}, http.StatusOK, traceID, place, h.KafkaProducer)
}
