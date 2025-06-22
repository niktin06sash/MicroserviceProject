package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/brokers/kafka"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/erro"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/handlers/response"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/metrics"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/service"
)

func (h *Handler) checkMethod(r *http.Request, w http.ResponseWriter, expectedMethod string, traceID string, place string) bool {
	if r.Method != expectedMethod {
		fmterr := fmt.Sprintf("Invalid request method(expected %s but it was sent %v)", expectedMethod, r.Method)
		h.LogProducer.NewUserLog(kafka.LogLevelWarn, place, traceID, fmterr)
		response.BadResponse(r, w, http.StatusBadRequest, erro.ErrorInvalidReqMethod, traceID, place, h.LogProducer)
		metrics.UserErrorsTotal.WithLabelValues(erro.ClientErrorType).Inc()
		return false
	}
	return true
}
func getAllData[T any](r *http.Request, w http.ResponseWriter, traceID string, place string, parser T, logproducer LogProducer) bool {
	datafromperson, err := io.ReadAll(r.Body)
	if err != nil {
		fmterr := fmt.Sprintf("ReadAll Error: %v", err)
		logproducer.NewUserLog(kafka.LogLevelWarn, place, traceID, fmterr)
		response.BadResponse(r, w, http.StatusBadRequest, erro.ErrorInvalidDataReq, traceID, place, logproducer)
		metrics.UserErrorsTotal.WithLabelValues(erro.ClientErrorType).Inc()
		return false
	}
	err = json.Unmarshal(datafromperson, parser)
	if err != nil {
		fmterr := fmt.Sprintf("Unmarshal Error: %v", err)
		logproducer.NewUserLog(kafka.LogLevelWarn, place, traceID, fmterr)
		response.BadResponse(r, w, http.StatusBadRequest, erro.ErrorInvalidDataReq, traceID, place, logproducer)
		metrics.UserErrorsTotal.WithLabelValues(erro.ClientErrorType).Inc()
		return false
	}
	return true
}
func (h *Handler) getPersonality(r *http.Request, w http.ResponseWriter, traceID string, place string, personmapa map[string]string) bool {
	sessionID, ok := r.Context().Value(sessionID).(string)
	if !ok {
		h.LogProducer.NewUserLog(kafka.LogLevelError, place, traceID, "Session ID not found in context")
		response.BadResponse(r, w, http.StatusInternalServerError, erro.UserServiceUnavalaible, traceID, place, h.LogProducer)
		metrics.UserErrorsTotal.WithLabelValues(erro.ServerErrorType).Inc()
		return false
	}
	personmapa[sessionID] = sessionID
	userID, ok := r.Context().Value(userID).(string)
	if !ok {
		h.LogProducer.NewUserLog(kafka.LogLevelError, place, traceID, "User ID not found in context")
		response.BadResponse(r, w, http.StatusInternalServerError, erro.UserServiceUnavalaible, traceID, place, h.LogProducer)
		metrics.UserErrorsTotal.WithLabelValues(erro.ServerErrorType).Inc()
		return false
	}
	personmapa[userID] = userID
	return true
}
func (h *Handler) serviceResponse(resp *service.ServiceResponse, r *http.Request, w http.ResponseWriter, traceID string, place string) bool {
	if !resp.Success {
		switch resp.Errors.Type {
		case erro.ClientErrorType:
			response.BadResponse(r, w, http.StatusBadRequest, resp.Errors.Message, traceID, place, h.LogProducer)
		case erro.ServerErrorType:
			response.BadResponse(r, w, http.StatusInternalServerError, resp.Errors.Message, traceID, place, h.LogProducer)
		}
		return false
	}
	return true
}
func (h *Handler) getQueryParameters(r *http.Request, w http.ResponseWriter, traceID string, place string, personmapa map[string]string) bool {
	query := r.URL.Query()
	switch place {
	case Update:
		updateName := query.Get("name")
		updatePassword := query.Get("password")
		updateEmail := query.Get("email")
		if (updateName != "" && updatePassword != "") || (updateName != "" && updateEmail != "") || (updatePassword != "" && updateEmail != "") {
			h.LogProducer.NewUserLog(kafka.LogLevelWarn, place, traceID, "More than one parameter in a query-request")
			response.BadResponse(r, w, http.StatusBadRequest, erro.ErrorInvalidCountQueryParameters, traceID, place, h.LogProducer)
			metrics.UserErrorsTotal.WithLabelValues(erro.ClientErrorType).Inc()
			return false
		}
		var updateType string
		switch {
		case updateName == "1":
			updateType = "name"
		case updatePassword == "1":
			updateType = "password"
		case updateEmail == "1":
			updateType = "email"
		default:
			h.LogProducer.NewUserLog(kafka.LogLevelWarn, place, traceID, "Invalid query parameter")
			response.BadResponse(r, w, http.StatusBadRequest, erro.ErrorInvalidQueryParameter, traceID, place, h.LogProducer)
			metrics.UserErrorsTotal.WithLabelValues(erro.ClientErrorType).Inc()
			return false
		}
		personmapa[update_type] = updateType
		return true
	}
	return false
}
func (h *Handler) getDinamicParameters(r *http.Request, w http.ResponseWriter, traceID string, place string, personmapa map[string]string) bool {
	vars := mux.Vars(r)
	switch place {
	case GetUserById:
		userID, ok := vars["id"]
		if !ok {
			h.LogProducer.NewUserLog(kafka.LogLevelWarn, place, traceID, "Invalid dinamic parameter")
			response.BadResponse(r, w, http.StatusBadRequest, erro.ErrorInvalidDinamicParameter, traceID, place, h.LogProducer)
			metrics.UserErrorsTotal.WithLabelValues(erro.ClientErrorType).Inc()
			return false
		}
		personmapa[getID] = userID
		return true
	}
	return false
}
