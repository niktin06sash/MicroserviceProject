package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/erro"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/handlers/response"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/kafka"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/metrics"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/service"
)

func checkMethod(r *http.Request, w http.ResponseWriter, expectedMethod string, traceID string, mapa map[string]string, place string, kafkaproducer kafka.KafkaProducerService) bool {
	start := r.Context().Value("starttime").(time.Time)
	if r.Method != expectedMethod {
		fmterr := fmt.Sprintf("Invalid request method(expected %s but it was sent %v)", expectedMethod, r.Method)
		kafkaproducer.NewUserLog(kafka.LogLevelWarn, place, traceID, fmterr)
		mapa["ClientError"] = "Invalid request method"
		response.SendResponse(r.Context(), w, false, nil, mapa, http.StatusBadRequest, traceID, place, kafkaproducer)
		metrics.UserErrorsTotal.WithLabelValues("ClientError").Inc()
		duration := time.Since(start).Seconds()
		metrics.UserRequestDuration.WithLabelValues(place).Observe(duration)
		return false
	}
	return true
}
func getAllData[T any](r *http.Request, w http.ResponseWriter, traceID string, mapa map[string]string, place string, parser T, kafkaproducer kafka.KafkaProducerService) bool {
	start := r.Context().Value("starttime").(time.Time)
	datafromperson, err := io.ReadAll(r.Body)
	if err != nil {
		fmterr := fmt.Sprintf("ReadAll Error: %v", err)
		kafkaproducer.NewUserLog(kafka.LogLevelWarn, place, traceID, fmterr)
		mapa["ClientError"] = erro.ErrorReadAll.Error()
		response.SendResponse(r.Context(), w, false, nil, mapa, http.StatusBadRequest, traceID, place, kafkaproducer)
		metrics.UserErrorsTotal.WithLabelValues("ClientError").Inc()
		duration := time.Since(start).Seconds()
		metrics.UserRequestDuration.WithLabelValues(place).Observe(duration)
		return false
	}
	err = json.Unmarshal(datafromperson, parser)
	if err != nil {
		fmterr := fmt.Sprintf("Unmarshal Error: %v", err)
		kafkaproducer.NewUserLog(kafka.LogLevelWarn, place, traceID, fmterr)
		mapa["ClientError"] = erro.ErrorUnmarshal.Error()
		response.SendResponse(r.Context(), w, false, nil, mapa, http.StatusBadRequest, traceID, place, kafkaproducer)
		metrics.UserErrorsTotal.WithLabelValues("ClientError").Inc()
		duration := time.Since(start).Seconds()
		metrics.UserRequestDuration.WithLabelValues(place).Observe(duration)
		return false
	}
	return true
}
func getUserIdAndSession(r *http.Request, w http.ResponseWriter, traceID string, mapa map[string]string, place string, kafkaproducer kafka.KafkaProducerService) (bool, string, string) {
	start := r.Context().Value("starttime").(time.Time)
	sessionID, ok := r.Context().Value("sessionID").(string)
	if !ok {
		kafkaproducer.NewUserLog(kafka.LogLevelError, place, traceID, "Session ID not found in context")
		mapa["InternalServerError"] = erro.ErrorMissingSessionID.Error()
		response.SendResponse(r.Context(), w, false, nil, mapa, http.StatusInternalServerError, traceID, place, kafkaproducer)
		metrics.UserErrorsTotal.WithLabelValues("InternalServerError").Inc()
		duration := time.Since(start).Seconds()
		metrics.UserRequestDuration.WithLabelValues(place).Observe(duration)
		return false, "", ""
	}
	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		kafkaproducer.NewUserLog(kafka.LogLevelError, place, traceID, "User ID not found in context")
		mapa["InternalServerError"] = erro.ErrorMissingUserID.Error()
		response.SendResponse(r.Context(), w, false, nil, mapa, http.StatusInternalServerError, traceID, place, kafkaproducer)
		metrics.UserErrorsTotal.WithLabelValues("InternalServerError").Inc()
		duration := time.Since(start).Seconds()
		metrics.UserRequestDuration.WithLabelValues(place).Observe(duration)
		return false, "", ""
	}
	return true, sessionID, userID
}
func serviceResponse(resp *service.ServiceResponse, r *http.Request, w http.ResponseWriter, traceID string, place string, kafkaproducer kafka.KafkaProducerService) bool {
	start := r.Context().Value("starttime").(time.Time)
	if !resp.Success {
		stringMap := response.ConvertErrorsToString(resp.Errors)
		switch resp.Type {
		case erro.ClientErrorType:
			response.SendResponse(r.Context(), w, false, nil, stringMap, http.StatusBadRequest, traceID, place, kafkaproducer)
		case erro.ServerErrorType:
			response.SendResponse(r.Context(), w, false, nil, stringMap, http.StatusInternalServerError, traceID, place, kafkaproducer)
		}
		duration := time.Since(start).Seconds()
		metrics.UserRequestDuration.WithLabelValues(place).Observe(duration)
		return false
	}
	return true
}
