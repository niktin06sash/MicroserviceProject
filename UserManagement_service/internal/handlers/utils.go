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
		mapa["ClientError"] = "Invalid request method"
		response.SendResponse(r.Context(), w, false, nil, mapa, http.StatusBadRequest, traceID, place, kafkaproducer)
		metrics.UserErrorsTotal.WithLabelValues("ClientError").Inc()
		return false
	}
	return true
}
func getAllData[T any](r *http.Request, w http.ResponseWriter, traceID string, mapa map[string]string, place string, parser T, kafkaproducer kafka.KafkaProducerService) bool {
	datafromperson, err := io.ReadAll(r.Body)
	if err != nil {
		fmterr := fmt.Sprintf("ReadAll Error: %v", err)
		kafkaproducer.NewUserLog(kafka.LogLevelWarn, place, traceID, fmterr)
		mapa["ClientError"] = erro.ErrorReadAll.Error()
		response.SendResponse(r.Context(), w, false, nil, mapa, http.StatusBadRequest, traceID, place, kafkaproducer)
		metrics.UserErrorsTotal.WithLabelValues("ClientError").Inc()
		return false
	}
	err = json.Unmarshal(datafromperson, parser)
	if err != nil {
		fmterr := fmt.Sprintf("Unmarshal Error: %v", err)
		kafkaproducer.NewUserLog(kafka.LogLevelWarn, place, traceID, fmterr)
		mapa["ClientError"] = erro.ErrorUnmarshal.Error()
		response.SendResponse(r.Context(), w, false, nil, mapa, http.StatusBadRequest, traceID, place, kafkaproducer)
		metrics.UserErrorsTotal.WithLabelValues("ClientError").Inc()
		return false
	}
	return true
}
func getPersonality(r *http.Request, w http.ResponseWriter, traceID string, mapa map[string]string, place string, kafkaproducer kafka.KafkaProducerService) (bool, map[string]string) {
	personmapa := make(map[string]string)
	sessionID, ok := r.Context().Value("sessionID").(string)
	if !ok {
		kafkaproducer.NewUserLog(kafka.LogLevelError, place, traceID, "Session ID not found in context")
		mapa["InternalServerError"] = erro.UserServiceUnavalaible
		response.SendResponse(r.Context(), w, false, nil, mapa, http.StatusInternalServerError, traceID, place, kafkaproducer)
		metrics.UserErrorsTotal.WithLabelValues("InternalServerError").Inc()
		return false, nil
	}
	personmapa["sessionID"] = sessionID
	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		kafkaproducer.NewUserLog(kafka.LogLevelError, place, traceID, "User ID not found in context")
		mapa["InternalServerError"] = erro.UserServiceUnavalaible
		response.SendResponse(r.Context(), w, false, nil, mapa, http.StatusInternalServerError, traceID, place, kafkaproducer)
		metrics.UserErrorsTotal.WithLabelValues("InternalServerError").Inc()
		return false, nil
	}
	personmapa["userID"] = userID
	return true, personmapa
}
func serviceResponse(resp *service.ServiceResponse, r *http.Request, w http.ResponseWriter, traceID string, place string, kafkaproducer kafka.KafkaProducerService) bool {
	if !resp.Success {
		stringMap := response.ConvertErrorsToString(resp.Errors)
		switch resp.Type {
		case erro.ClientErrorType:
			response.SendResponse(r.Context(), w, false, nil, stringMap, http.StatusBadRequest, traceID, place, kafkaproducer)
		case erro.ServerErrorType:
			response.SendResponse(r.Context(), w, false, nil, stringMap, http.StatusInternalServerError, traceID, place, kafkaproducer)
		}
		return false
	}
	return true
}
func getQueryParameters(r *http.Request, w http.ResponseWriter, traceId string, mapa map[string]string, place string, kafkaproducer kafka.KafkaProducerService) (bool, map[string]string) {
	personmapa := make(map[string]string)
	switch place {
	case Update:
		query := r.URL.Query()
		updateName := query.Get("name")
		updatePassword := query.Get("password")
		updateEmail := query.Get("email")
		if (updateName != "" && updatePassword != "") || (updateName != "" && updateEmail != "") || (updatePassword != "" && updateEmail != "") {
			kafkaproducer.NewUserLog(kafka.LogLevelWarn, place, traceId, "More than one parameter in a query-request")
			mapa["ClientError"] = "More than one parameter in a query-request"
			response.SendResponse(r.Context(), w, false, nil, mapa, http.StatusBadRequest, traceId, place, kafkaproducer)
			metrics.UserErrorsTotal.WithLabelValues("ClientError").Inc()
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
			kafkaproducer.NewUserLog(kafka.LogLevelWarn, place, traceId, "Invalid query parameter")
			mapa["ClientError"] = "Invalid query parameter"
			response.SendResponse(r.Context(), w, false, nil, mapa, http.StatusBadRequest, traceId, place, kafkaproducer)
			metrics.UserErrorsTotal.WithLabelValues("ClientError").Inc()
			return false, nil
		}
		personmapa["update_type"] = updateType
		return true, personmapa
	}
	return false, nil
}
