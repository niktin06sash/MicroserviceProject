package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/brokers/kafka"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/handlers/response"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/model"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/repository"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/service"
)

type Handler struct {
	Services    UserService
	Middlewares MiddlewareService
	LogProducer LogProducer
}

func NewHandler(services UserService, middleware MiddlewareService, logproducer LogProducer) *Handler {
	return &Handler{Services: services, Middlewares: middleware, LogProducer: logproducer}
}
func (h *Handler) Registration(w http.ResponseWriter, r *http.Request) {
	const place = Registration
	defer r.Body.Close()
	traceID := r.Context().Value("traceID").(string)
	if !h.checkMethod(r, w, http.MethodPost, traceID, place) {
		return
	}
	var regreq model.RegistrationRequest
	if !getAllData(r, w, traceID, place, &regreq, h.LogProducer) {
		return
	}
	regresponse := h.Services.RegistrateAndLogin(r.Context(), &regreq)
	if !h.serviceResponse(regresponse, r, w, traceID, place) {
		return
	}
	userID := regresponse.Data[repository.KeyUserID].(string)
	sessionID := regresponse.Data[service.KeySessionID].(string)
	expiresession := regresponse.Data[service.KeyExpirySession].(time.Time)
	response.AddSessionCookie(w, sessionID, expiresession)
	msg := fmt.Sprintf("Person with id %v has successfully registered", userID)
	h.LogProducer.NewUserLog(kafka.LogLevelInfo, place, traceID, msg)
	response.OkResponse(r, w, http.StatusOK, map[string]any{response.KeyMessage: "You have successfully registered"}, traceID, place, h.LogProducer)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	const place = Authentication
	defer r.Body.Close()
	traceID := r.Context().Value("traceID").(string)
	if !h.checkMethod(r, w, http.MethodPost, traceID, place) {
		return
	}
	var aureq model.AuthenticationRequest
	if !getAllData(r, w, traceID, place, &aureq, h.LogProducer) {
		return
	}
	auresponse := h.Services.AuthenticateAndLogin(r.Context(), &aureq)
	if !h.serviceResponse(auresponse, r, w, traceID, place) {
		return
	}
	userID := auresponse.Data[repository.KeyUserID].(string)
	sessionID := auresponse.Data[service.KeySessionID].(string)
	expiresession := auresponse.Data[service.KeyExpirySession].(time.Time)
	response.AddSessionCookie(w, sessionID, expiresession)
	msg := fmt.Sprintf("Person with id %v has successfully login", userID)
	h.LogProducer.NewUserLog(kafka.LogLevelInfo, place, traceID, msg)
	response.OkResponse(r, w, http.StatusOK, map[string]any{response.KeyMessage: "You have successfully login"}, traceID, place, h.LogProducer)
}

func (h *Handler) DeleteAccount(w http.ResponseWriter, r *http.Request) {
	const place = DeleteAccount
	defer r.Body.Close()
	traceID := r.Context().Value("traceID").(string)
	if !h.checkMethod(r, w, http.MethodDelete, traceID, place) {
		return
	}
	persondata := make(map[string]string)
	if !h.getPersonality(r, w, traceID, place, persondata) {
		return
	}
	var delreq model.DeletionRequest
	if !getAllData(r, w, traceID, place, &delreq, h.LogProducer) {
		return
	}
	defer r.Body.Close()
	delresponse := h.Services.DeleteAccount(r.Context(), &delreq, persondata["sessionID"], persondata["userID"])
	if !h.serviceResponse(delresponse, r, w, traceID, place) {
		return
	}
	response.DeleteSessionCookie(w)
	msg := fmt.Sprintf("Person with id %v has successfully deleted account", persondata["userID"])
	h.LogProducer.NewUserLog(kafka.LogLevelInfo, place, traceID, msg)
	response.OkResponse(r, w, http.StatusOK, map[string]any{response.KeyMessage: "You have successfully deleted account"}, traceID, place, h.LogProducer)
}
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	const place = Logout
	defer r.Body.Close()
	traceID := r.Context().Value("traceID").(string)
	if !h.checkMethod(r, w, http.MethodDelete, traceID, place) {
		return
	}
	persondata := make(map[string]string)
	if !h.getPersonality(r, w, traceID, place, persondata) {
		return
	}
	logresponse := h.Services.Logout(r.Context(), persondata["sessionID"])
	if !h.serviceResponse(logresponse, r, w, traceID, place) {
		return
	}
	response.DeleteSessionCookie(w)
	msg := fmt.Sprintf("Person with id %v has successfully logout", persondata["userID"])
	h.LogProducer.NewUserLog(kafka.LogLevelInfo, place, traceID, msg)
	response.OkResponse(r, w, http.StatusOK, map[string]any{response.KeyMessage: "You have successfully logout"}, traceID, place, h.LogProducer)
}
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	const place = Update
	defer r.Body.Close()
	traceID := r.Context().Value("traceID").(string)
	if !h.checkMethod(r, w, http.MethodPatch, traceID, place) {
		return
	}
	persondata := make(map[string]string)
	if !h.getPersonality(r, w, traceID, place, persondata) {
		return
	}
	var updatereq model.UpdateRequest
	if !getAllData(r, w, traceID, place, &updatereq, h.LogProducer) {
		return
	}
	if !h.getQueryParameters(r, w, traceID, place, persondata) {
		return
	}
	updateresponse := h.Services.UpdateAccount(r.Context(), &updatereq, persondata["userID"], persondata["update_type"])
	if !h.serviceResponse(updateresponse, r, w, traceID, place) {
		return
	}
	msg := fmt.Sprintf("Person with id %v has successfully updated his %s", persondata["userID"], persondata["update_type"])
	h.LogProducer.NewUserLog(kafka.LogLevelInfo, place, traceID, msg)
	response.OkResponse(r, w, http.StatusOK, map[string]any{response.KeyMessage: fmt.Sprintf("You have successfully updated your %v", persondata["update_type"])}, traceID, place, h.LogProducer)
}
func (h *Handler) MyProfile(w http.ResponseWriter, r *http.Request) {
	const place = MyProfile
	defer r.Body.Close()
	traceID := r.Context().Value("traceID").(string)
	if !h.checkMethod(r, w, http.MethodGet, traceID, place) {
		return
	}
	persondata := make(map[string]string)
	if !h.getPersonality(r, w, traceID, place, persondata) {
		return
	}
	myprofileresponse := h.Services.GetMyProfile(r.Context(), persondata["userID"])
	if !h.serviceResponse(myprofileresponse, r, w, traceID, place) {
		return
	}
	msg := fmt.Sprintf("Person with id %v has successfully received his account data", persondata["userID"])
	h.LogProducer.NewUserLog(kafka.LogLevelInfo, place, traceID, msg)
	response.OkResponse(r, w, http.StatusOK, myprofileresponse.Data, traceID, place, h.LogProducer)
}
func (h *Handler) GetUserById(w http.ResponseWriter, r *http.Request) {
	const place = GetUserById
	defer r.Body.Close()
	traceID := r.Context().Value("traceID").(string)
	if !h.checkMethod(r, w, http.MethodGet, traceID, place) {
		return
	}
	persondata := make(map[string]string)
	if !h.getPersonality(r, w, traceID, place, persondata) {
		return
	}
	if !h.getDinamicParameters(r, w, traceID, place, persondata) {
		return
	}
	getprofileresponse := h.Services.GetProfileById(r.Context(), persondata["userID"], persondata["getID"])
	if !h.serviceResponse(getprofileresponse, r, w, traceID, place) {
		return
	}
	msg := fmt.Sprintf("Person with id %v has successfully received data of person with id %v", persondata["userID"], persondata["getID"])
	h.LogProducer.NewUserLog(kafka.LogLevelInfo, place, traceID, msg)
	response.OkResponse(r, w, http.StatusOK, getprofileresponse.Data, traceID, place, h.LogProducer)
}
