package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/erro"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/handlers/response"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/kafka"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/metrics"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/service"
)

func checkMethod(r *http.Request, w http.ResponseWriter, expectedMethod string, traceID string, mapa map[string]string, place string, kafkaproducer kafka.KafkaProducerService) bool {
	if r.Method != expectedMethod {
		fmterr := fmt.Sprintf("Invalid request method(expected %s but it was sent %v)", expectedMethod, r.Method)
		kafkaproducer.NewUserLog(kafka.LogLevelWarn, place, traceID, fmterr)
		mapa[erro.ClientErrorType] = "Invalid request method"
		response.SendResponse(r.Context(), w, response.HTTPResponse{Success: false, Data: nil, Errors: mapa}, http.StatusBadRequest, traceID, place, kafkaproducer)
		metrics.UserErrorsTotal.WithLabelValues(erro.ClientErrorType).Inc()
		return false
	}
	return true
}
func getAllData[T any](r *http.Request, w http.ResponseWriter, traceID string, mapa map[string]string, place string, parser T, kafkaproducer kafka.KafkaProducerService) bool {
	datafromperson, err := io.ReadAll(r.Body)
	if err != nil {
		fmterr := fmt.Sprintf("ReadAll Error: %v", err)
		kafkaproducer.NewUserLog(kafka.LogLevelWarn, place, traceID, fmterr)
		mapa[erro.ClientErrorType] = erro.ErrorReadAllConst
		response.SendResponse(r.Context(), w, response.HTTPResponse{Success: false, Data: nil, Errors: mapa}, http.StatusBadRequest, traceID, place, kafkaproducer)
		metrics.UserErrorsTotal.WithLabelValues(erro.ClientErrorType).Inc()
		return false
	}
	err = json.Unmarshal(datafromperson, parser)
	if err != nil {
		fmterr := fmt.Sprintf("Unmarshal Error: %v", err)
		kafkaproducer.NewUserLog(kafka.LogLevelWarn, place, traceID, fmterr)
		mapa[erro.ClientErrorType] = erro.ErrorUnmarshalConst
		response.SendResponse(r.Context(), w, response.HTTPResponse{Success: false, Data: nil, Errors: mapa}, http.StatusBadRequest, traceID, place, kafkaproducer)
		metrics.UserErrorsTotal.WithLabelValues(erro.ClientErrorType).Inc()
		return false
	}
	return true
}
func getPersonality(r *http.Request, w http.ResponseWriter, traceID string, mapa map[string]string, place string, kafkaproducer kafka.KafkaProducerService) (bool, map[string]string) {
	personmapa := make(map[string]string)
	sessionID, ok := r.Context().Value("sessionID").(string)
	if !ok {
		kafkaproducer.NewUserLog(kafka.LogLevelError, place, traceID, "Session ID not found in context")
		mapa[erro.ServerErrorType] = erro.UserServiceUnavalaible
		response.SendResponse(r.Context(), w, response.HTTPResponse{Success: false, Data: nil, Errors: mapa}, http.StatusInternalServerError, traceID, place, kafkaproducer)
		metrics.UserErrorsTotal.WithLabelValues(erro.ServerErrorType).Inc()
		return false, nil
	}
	personmapa["sessionID"] = sessionID
	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		kafkaproducer.NewUserLog(kafka.LogLevelError, place, traceID, "User ID not found in context")
		mapa[erro.ServerErrorType] = erro.UserServiceUnavalaible
		response.SendResponse(r.Context(), w, response.HTTPResponse{Success: false, Data: nil, Errors: mapa}, http.StatusInternalServerError, traceID, place, kafkaproducer)
		metrics.UserErrorsTotal.WithLabelValues(erro.ServerErrorType).Inc()
		return false, nil
	}
	personmapa["userID"] = userID
	return true, personmapa
}
func serviceResponse(resp *service.ServiceResponse, r *http.Request, w http.ResponseWriter, traceID string, place string, kafkaproducer kafka.KafkaProducerService) bool {
	if !resp.Success {
		switch resp.ErrorType {
		case erro.ClientErrorType:
			response.SendResponse(r.Context(), w, response.HTTPResponse{Success: false, Data: nil, Errors: resp.Errors}, http.StatusBadRequest, traceID, place, kafkaproducer)
		case erro.ServerErrorType:
			response.SendResponse(r.Context(), w, response.HTTPResponse{Success: false, Data: nil, Errors: resp.Errors}, http.StatusInternalServerError, traceID, place, kafkaproducer)
		}
		return false
	}
	return true
}
func getQueryParameters(r *http.Request, w http.ResponseWriter, traceID string, mapa map[string]string, place string, kafkaproducer kafka.KafkaProducerService) (bool, map[string]string) {
	personmapa := make(map[string]string)
	switch place {
	case Update:
		query := r.URL.Query()
		updateName := query.Get("name")
		updatePassword := query.Get("password")
		updateEmail := query.Get("email")
		if (updateName != "" && updatePassword != "") || (updateName != "" && updateEmail != "") || (updatePassword != "" && updateEmail != "") {
			kafkaproducer.NewUserLog(kafka.LogLevelWarn, place, traceID, "More than one parameter in a query-request")
			mapa[erro.ClientErrorType] = "More than one parameter in a query-request"
			response.SendResponse(r.Context(), w, response.HTTPResponse{Success: false, Data: nil, Errors: mapa}, http.StatusBadRequest, traceID, place, kafkaproducer)
			metrics.UserErrorsTotal.WithLabelValues(erro.ClientErrorType).Inc()
			return false, nil
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
			kafkaproducer.NewUserLog(kafka.LogLevelWarn, place, traceID, "Invalid query parameter")
			mapa[erro.ClientErrorType] = "Invalid query parameter"
			response.SendResponse(r.Context(), w, response.HTTPResponse{Success: false, Data: nil, Errors: mapa}, http.StatusBadRequest, traceID, place, kafkaproducer)
			metrics.UserErrorsTotal.WithLabelValues(erro.ClientErrorType).Inc()
			return false, nil
		}
		personmapa["update_type"] = updateType
		return true, personmapa
	}
	return false, nil
}
