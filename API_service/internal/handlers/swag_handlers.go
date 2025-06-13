package handlers

import (
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/niktin06sash/MicroserviceProject/API_service/internal/brokers/kafka"
	"github.com/niktin06sash/MicroserviceProject/API_service/internal/erro"
	"github.com/niktin06sash/MicroserviceProject/API_service/internal/handlers/response"
	_ "github.com/niktin06sash/MicroserviceProject/API_service/internal/handlers/response"
	"github.com/niktin06sash/MicroserviceProject/API_service/internal/metrics"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// @Summary Register a new user
// @Description Register a new user by sending user data to the target service.
// @Tags User Management
// @Accept json
// @Produce json
// @Param input body response.PersonReg true "User registration data"
// @Success 200 {object} response.HTTPResponse "User successfully registered"
// @Failure 400 {object} response.HTTPResponse "Invalid input data"
// @Failure 403 {object} response.HTTPResponse "Forbidden"
// @Failure 429 {object} response.HTTPResponse "Too many requests"
// @Failure 500 {object} response.HTTPResponse "Internal server error"
// @Router /api/auth/register [post]
func (h *Handler) Registration(c *gin.Context) {
	h.ProxyHTTP(c)
}

// @Summary Authenticate a user
// @Description Authenticate a user by sending credentials to the target service.
// @Tags User Management
// @Accept json
// @Produce json
// @Param input body response.PersonAuth true "User credentials"
// @Success 200 {object} response.HTTPResponse "User successfully authenticated"
// @Failure 400 {object} response.HTTPResponse "Invalid input data"
// @Failure 403 {object} response.HTTPResponse "Forbidden"
// @Failure 429 {object} response.HTTPResponse "Too many requests"
// @Failure 500 {object} response.HTTPResponse "Internal server error"
// @Router /api/auth/login [post]
func (h *Handler) Login(c *gin.Context) {
	h.ProxyHTTP(c)
}

// @Summary Delete a user
// @Description Delete a user by sending a DELETE request with session and user ID.
// @Tags User Management
// @Accept json
// @Produce json
// @Param input body response.PersonDelete true "User credentials"
// @Success 200 {object} response.HTTPResponse "User successfully deleted"
// @Failure 400 {object} response.HTTPResponse "Invalid input data"
// @Failure 401 {object} response.HTTPResponse "Unauthorized"
// @Failure 429 {object} response.HTTPResponse "Too many requests"
// @Failure 500 {object} response.HTTPResponse "Internal server error"
// @Router /api/users/del [delete]
func (h *Handler) DeleteUser(c *gin.Context) {
	h.ProxyHTTP(c)
}

// @Summary Logout a user
// @Description Logout a user by sending a request with session.
// @Tags User Management
// @Produce json
// @Success 200 {object} response.HTTPResponse "User successfully logout"
// @Failure 401 {object} response.HTTPResponse "Unauthorized"
// @Failure 429 {object} response.HTTPResponse "Too many requests"
// @Failure 500 {object} response.HTTPResponse "Internal server error"
// @Router /api/users/logout [delete]
func (h *Handler) Logout(c *gin.Context) {
	h.ProxyHTTP(c)
}

// @Summary Update a user's profile
// @Description Update a user's profile by sending a request with session.
// @Tags User Management
// @Accept json
// @Produce json
// @Param name query string false "Flag to update name"
// @Param email query string false "Flag to update email"
// @Param password query string false "Flag to update password"
// @Param input body response.PersonUpdate true "User update data"
// @Success 200 {object} response.HTTPResponse "User successfully updated his profile"
// @Failure 400 {object} response.HTTPResponse "Invalid input data"
// @Failure 401 {object} response.HTTPResponse "Unauthorized"
// @Failure 429 {object} response.HTTPResponse "Too many requests"
// @Failure 500 {object} response.HTTPResponse "Internal server error"
// @Router /api/me/update [patch]
func (h *Handler) Update(c *gin.Context) {
	h.ProxyHTTP(c)
}

// @Summary Received a user's profile data
// @Description Received a user's profile data by sending a request with session.
// @Tags User Management
// @Produce json
// @Success 200 {object} response.HTTPResponse "User successfully received his profile data"
// @Failure 401 {object} response.HTTPResponse "Unauthorized"
// @Failure 429 {object} response.HTTPResponse "Too many requests"
// @Failure 500 {object} response.HTTPResponse "Internal server error"
// @Router /api/me [get]
func (h *Handler) MyProfile(c *gin.Context) {

}

// @Summary Find userprofile by ID in path parameters
// @Description Retrieves a user's profile data by sending a request with the userID as a path parameter.
// @Tags User Management
// @Produce json
// @Param id path string true "userID in UUID format"
// @Success 200 {object} response.HTTPResponse "User profile data successfully retrieved"
// @Failure 400 {object} response.HTTPResponse "Invalid input data"
// @Failure 401 {object} response.HTTPResponse "Unauthorized"
// @Failure 429 {object} response.HTTPResponse "Too many requests"
// @Failure 500 {object} response.HTTPResponse "Internal server error"
// @Router /api/users/{id} [get]
func (h *Handler) GetUserProfileById(c *gin.Context) {

}

// @Summary Find photo by ID's in path parameters
// @Description Retrieves a user's photo by sending a request with the userID, photoID as a path parameter.
// @Tags Photo Service
// @Produce json
// @Param id path string true "userID in UUID format"
// @Param photo_id path string true "photoID in UUID format"
// @Success 200 {object} response.HTTPResponse "User's photo successfully retrieved"
// @Failure 400 {object} response.HTTPResponse "Invalid input data"
// @Failure 401 {object} response.HTTPResponse "Unauthorized"
// @Failure 429 {object} response.HTTPResponse "Too many requests"
// @Failure 500 {object} response.HTTPResponse "Internal server error"
// @Router /api/users/{id}/photos/{photo_id} [get]
func (h *Handler) GetPhotoById(c *gin.Context) {
}

// @Summary Delete photo by ID's in path parameters
// @Description Retrieves a own photo by sending a request with the photoID as a path parameter.
// @Tags Photo Service
// @Produce json
// @Param photo_id path string true "photoID in UUID format"
// @Success 200 {object} response.HTTPResponse "Photo successfully deleted"
// @Failure 400 {object} response.HTTPResponse "Invalid input data"
// @Failure 401 {object} response.HTTPResponse "Unauthorized"
// @Failure 429 {object} response.HTTPResponse "Too many requests"
// @Failure 500 {object} response.HTTPResponse "Internal server error"
// @Router /api/me/photos/{photo_id} [delete]
func (h *Handler) DeletePhoto(c *gin.Context) {
}

// @Summary Upload user photo
// @Description Uploads a new photo for user
// @Tags Photo Service
// @Accept multipart/form-data
// @Produce json
// @Param photo formData file true "Image file (JPG/PNG)"
// @Success 200 {object} response.HTTPResponse "Photo successfully uploaded"
// @Failure 400 {object} response.HTTPResponse "Invalid input data"
// @Failure 401 {object} response.HTTPResponse "Unauthorized"
// @Failure 429 {object} response.HTTPResponse "Too many requests"
// @Failure 500 {object} response.HTTPResponse "Internal server error"
// @Router /api/me/photos [post]
func (h *Handler) LoadPhoto(c *gin.Context) {
	const place = LoadPhoto
	traceID := c.MustGet("traceID").(string)
	userid := c.MustGet("userID").(string)
	file, _, err := c.Request.FormFile("photo")
	if err != nil {
		h.logproducer.NewAPILog(c.Request, kafka.LogLevelWarn, place, traceID, erro.RequiredFormPhoto)
		response.BadResponse(c, http.StatusBadRequest, erro.RequiredFormPhoto, traceID, place, h.logproducer)
		c.Abort()
		metrics.APIErrorsTotal.WithLabelValues(erro.ClientErrorType).Inc()
		return
	}
	bytes, err := io.ReadAll(file)
	if err != nil {
		h.logproducer.NewAPILog(c.Request, kafka.LogLevelWarn, place, traceID, fmt.Sprintf("Failed readAll: %v", err))
		response.BadResponse(c, http.StatusBadRequest, erro.APIServiceUnavalaible, traceID, place, h.logproducer)
		c.Abort()
		metrics.APIErrorsTotal.WithLabelValues(erro.ServerErrorType).Inc()
		return
	}
	protoresponse, err := h.photoclient.LoadPhoto(c.Request.Context(), userid, bytes)
	if err == nil && protoresponse != nil && protoresponse.Status {
		response.OkResponse(c, http.StatusOK, map[string]any{response.KeyMessage: protoresponse.Message, response.KeyPhotoID: protoresponse.PhotoId}, traceID, place, h.logproducer)
		return
	}
	st, _ := status.FromError(err)
	switch st.Code() {
	case codes.Internal, codes.Canceled, codes.Unavailable:
		response.BadResponse(c, http.StatusInternalServerError, erro.PhotoServiceUnavalaible, traceID, place, h.logproducer)
		c.Abort()
		return
	default:
		response.BadResponse(c, http.StatusBadRequest, st.Message(), traceID, place, h.logproducer)
		c.Abort()
		return
	}
}

// @Summary Find own photo by ID's in path parameters
// @Description Retrieves a own photo by sending a request with the photoID as a path parameter.
// @Tags Photo Service
// @Produce json
// @Param photo_id path string true "photoID in UUID format"
// @Success 200 {object} response.HTTPResponse "Own photo successfully retrieved"
// @Failure 400 {object} response.HTTPResponse "Invalid input data"
// @Failure 401 {object} response.HTTPResponse "Unauthorized"
// @Failure 429 {object} response.HTTPResponse "Too many requests"
// @Failure 500 {object} response.HTTPResponse "Internal server error"
// @Router /api/me/photos/{photo_id} [get]
func (h *Handler) GetMyPhotoById(c *gin.Context) {
}
