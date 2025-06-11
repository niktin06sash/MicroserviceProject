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
		response.SendResponse(r, w, response.HTTPResponse{Success: false, Data: nil, Errors: map[string]string{erro.ErrorType: erro.ClientErrorType, erro.ErrorMessage: erro.ErrorInvalidReqMethod}}, http.StatusBadRequest, traceID, place, h.LogProducer)
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
		response.SendResponse(r, w, response.HTTPResponse{Success: false, Data: nil, Errors: map[string]string{erro.ErrorType: erro.ClientErrorType, erro.ErrorMessage: erro.ErrorInvalidDataReq}}, http.StatusBadRequest, traceID, place, logproducer)
		metrics.UserErrorsTotal.WithLabelValues(erro.ClientErrorType).Inc()
		return false
	}
	err = json.Unmarshal(datafromperson, parser)
	if err != nil {
		fmterr := fmt.Sprintf("Unmarshal Error: %v", err)
		logproducer.NewUserLog(kafka.LogLevelWarn, place, traceID, fmterr)
		response.SendResponse(r, w, response.HTTPResponse{Success: false, Data: nil, Errors: map[string]string{erro.ErrorType: erro.ClientErrorType, erro.ErrorMessage: erro.ErrorInvalidDataReq}}, http.StatusBadRequest, traceID, place, logproducer)
		metrics.UserErrorsTotal.WithLabelValues(erro.ClientErrorType).Inc()
		return false
	}
	return true
}
func (h *Handler) getPersonality(r *http.Request, w http.ResponseWriter, traceID string, place string, personmapa map[string]string) bool {
	sessionID, ok := r.Context().Value("sessionID").(string)
	if !ok {
		h.LogProducer.NewUserLog(kafka.LogLevelError, place, traceID, "Session ID not found in context")
		response.SendResponse(r, w, response.HTTPResponse{Success: false, Data: nil, Errors: map[string]string{erro.ErrorType: erro.ServerErrorType, erro.ErrorMessage: erro.UserServiceUnavalaible}}, http.StatusInternalServerError, traceID, place, h.LogProducer)
		metrics.UserErrorsTotal.WithLabelValues(erro.ServerErrorType).Inc()
		return false
	}
	personmapa["sessionID"] = sessionID
	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		h.LogProducer.NewUserLog(kafka.LogLevelError, place, traceID, "User ID not found in context")
		response.SendResponse(r, w, response.HTTPResponse{Success: false, Data: nil, Errors: map[string]string{erro.ErrorType: erro.ServerErrorType, erro.ErrorMessage: erro.UserServiceUnavalaible}}, http.StatusInternalServerError, traceID, place, h.LogProducer)
		metrics.UserErrorsTotal.WithLabelValues(erro.ServerErrorType).Inc()
		return false
	}
	personmapa["userID"] = userID
	return true
}
func (h *Handler) serviceResponse(resp *service.ServiceResponse, r *http.Request, w http.ResponseWriter, traceID string, place string) bool {
	if !resp.Success {
		switch resp.Errors[erro.ErrorType] {
		case erro.ClientErrorType:
			response.SendResponse(r, w, response.HTTPResponse{Success: false, Data: nil, Errors: resp.Errors}, http.StatusBadRequest, traceID, place, h.LogProducer)
		case erro.ServerErrorType:
			response.SendResponse(r, w, response.HTTPResponse{Success: false, Data: nil, Errors: resp.Errors}, http.StatusInternalServerError, traceID, place, h.LogProducer)
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
			response.SendResponse(r, w, response.HTTPResponse{Success: false, Data: nil, Errors: map[string]string{erro.ErrorType: erro.ClientErrorType, erro.ErrorMessage: erro.ErrorInvalidCountQueryParameters}}, http.StatusBadRequest, traceID, place, h.LogProducer)
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
			response.SendResponse(r, w, response.HTTPResponse{Success: false, Data: nil, Errors: map[string]string{erro.ErrorType: erro.ClientErrorType, erro.ErrorMessage: erro.ErrorInvalidQueryParameter}}, http.StatusBadRequest, traceID, place, h.LogProducer)
			metrics.UserErrorsTotal.WithLabelValues(erro.ClientErrorType).Inc()
			return false
		}
		personmapa["update_type"] = updateType
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
			response.SendResponse(r, w, response.HTTPResponse{Success: false, Data: nil, Errors: map[string]string{erro.ErrorType: erro.ClientErrorType, erro.ErrorMessage: erro.ErrorInvalidDinamicParameter}}, http.StatusBadRequest, traceID, place, h.LogProducer)
			metrics.UserErrorsTotal.WithLabelValues(erro.ClientErrorType).Inc()
			return false
		}
		personmapa["getID"] = userID
		return true
	}
	return false
}
