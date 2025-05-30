package handlers

import (
	"fmt"
	"net/http"

	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/handlers/response"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/kafka"
)

func (h *Handler) MyFriends(w http.ResponseWriter, r *http.Request) {
	const place = MyFriends
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
	myfriendsresponse := h.Services.GetMyFriends(r.Context(), persondata["userID"])
	if !serviceResponse(myfriendsresponse, r, w, traceID, place, h.KafkaProducer) {
		return
	}
	msg := fmt.Sprintf("Person with id %v has successfully received his friends list", persondata["userID"])
	h.KafkaProducer.NewUserLog(kafka.LogLevelInfo, place, traceID, msg)
	response.SendResponse(r, w, response.HTTPResponse{Success: true, Data: myfriendsresponse.Data}, http.StatusOK, traceID, place, h.KafkaProducer)
}
