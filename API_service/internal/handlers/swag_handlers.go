package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/niktin06sash/MicroserviceProject/API_service/internal/brokers/kafka"
	"github.com/niktin06sash/MicroserviceProject/API_service/internal/erro"
	"github.com/niktin06sash/MicroserviceProject/API_service/internal/handlers/response"
	_ "github.com/niktin06sash/MicroserviceProject/API_service/internal/handlers/response"
	"github.com/niktin06sash/MicroserviceProject/API_service/internal/metrics"
	pb "github.com/niktin06sash/MicroserviceProject/Photo_service/proto"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
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
	const place = GetUserProfileById
	traceID := c.MustGet("traceID").(string)
	userid := c.MustGet("userID").(string)
	sessionid := c.MustGet("sessionID").(string)
	deadline, ok := c.Request.Context().Deadline()
	if !ok {
		h.logproducer.NewAPILog(c.Request, kafka.LogLevelWarn, place, traceID, "Failed to get deadline from context")
		deadline = time.Now().Add(15 * time.Second)
	}
	md := metadata.Pairs("traceID", traceID)
	ctx := metadata.NewOutgoingContext(c.Request.Context(), md)
	normalizedPath := metrics.NormalizePath(c.Request.URL.Path)
	target, ok := h.Routes[normalizedPath]
	if !ok {
		response.BadResponse(c, http.StatusBadRequest, erro.PageNotFound, traceID, place, h.logproducer)
		metrics.APIErrorsTotal.WithLabelValues(erro.ClientErrorType).Inc()
		return
	}
	paramuserid := c.Param("id")
	targetid := target + "/" + paramuserid
	err := response.CheckContext(c, place, traceID, h.logproducer)
	if err != nil {
		response.BadResponse(c, http.StatusInternalServerError, erro.RequestTimedOut, traceID, place, h.logproducer)
		return
	}
	g, ctx := errgroup.WithContext(ctx)
	httpresponseChan := make(chan response.HTTPResponse, 1)
	protoresponseChan := make(chan *pb.GetPhotosResponse, 1)
	defer close(httpresponseChan)
	defer close(protoresponseChan)
	g.Go(func() error {
		httprequest, err := http.NewRequest(http.MethodGet, targetid, c.Request.Body)
		if err != nil {
			h.logproducer.NewAPILog(c.Request, kafka.LogLevelError, place, traceID, fmt.Sprintf("New http-request create error: %v"))
			metrics.APIErrorsTotal.WithLabelValues(erro.ServerErrorType).Inc()
			return fmt.Errorf(erro.APIServiceUnavalaible)
		}
		httprequest.Header.Set("X-Deadline", deadline.Format(time.RFC3339))
		httprequest.Header.Set("X-User-ID", userid)
		httprequest.Header.Set("X-Trace-ID", traceID)
		httprequest.Header.Set("X-Session-ID", sessionid)
		httpresponse, err := http.DefaultClient.Do(httprequest)
		if err != nil {
			h.logproducer.NewAPILog(c.Request, kafka.LogLevelError, place, traceID, fmt.Sprintf("Http-request error: %v"))
			metrics.APIErrorsTotal.WithLabelValues(erro.ServerErrorType).Inc()
			return fmt.Errorf(erro.UserServiceUnavalaible)
		}
		defer httpresponse.Body.Close()
		data, err := io.ReadAll(httpresponse.Body)
		if err != nil {
			h.logproducer.NewAPILog(c.Request, kafka.LogLevelError, place, traceID, fmt.Sprintf("ReadAll http-response data error: %v"))
			metrics.APIErrorsTotal.WithLabelValues(erro.ServerErrorType).Inc()
			return fmt.Errorf(erro.APIServiceUnavalaible)
		}
		var userresponse response.HTTPResponse
		err = json.Unmarshal(data, &userresponse)
		if err != nil {
			h.logproducer.NewAPILog(c.Request, kafka.LogLevelError, place, traceID, fmt.Sprintf("Unmarshal http-response data error: %v"))
			metrics.APIErrorsTotal.WithLabelValues(erro.ServerErrorType).Inc()
			return fmt.Errorf(erro.APIServiceUnavalaible)
		}
		httpresponseChan <- userresponse
		return nil
	})
	g.Go(func() error {
		protoresponse, err := h.photoclient.GetPhotos(ctx, paramuserid)
		if err != nil {
			st, _ := status.FromError(err)
			switch st.Code() {
			case codes.Canceled, codes.Unavailable:
				return fmt.Errorf(erro.PhotoServiceUnavalaible)
			default:
				return err
			}
		}
		protoresponseChan <- protoresponse
		return nil
	})
	if err := g.Wait(); err != nil {
		if _, ok := status.FromError(err); ok {
			h.badGrpcResponse(c, traceID, place, err)
			return
		}
		response.BadResponse(c, http.StatusInternalServerError, err.Error(), traceID, place, h.logproducer)
		return
	}
	userresponse := <-httpresponseChan
	photoresp := <-protoresponseChan
	if !userresponse.Success {
		typ := userresponse.Errors[erro.ErrorType]
		if typ == erro.ServerErrorType {
			response.BadResponse(c, http.StatusInternalServerError, userresponse.Errors[erro.ErrorMessage], traceID, place, h.logproducer)
		} else {
			response.BadResponse(c, http.StatusBadRequest, userresponse.Errors[erro.ErrorMessage], traceID, place, h.logproducer)
		}
		return
	}
	totalresp := userresponse.Data
	totalresp[response.KeyPhotos] = photoresp.Photos
	response.OkResponse(c, http.StatusOK, totalresp, traceID, place, h.logproducer)
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
	const place = GetPhotoById
	traceID := c.MustGet("traceID").(string)
	userid := c.Param("id")
	photoid := c.Param("photo_id")
	md := metadata.Pairs("traceID", traceID)
	ctx := metadata.NewOutgoingContext(c.Request.Context(), md)
	err := response.CheckContext(c, place, traceID, h.logproducer)
	if err != nil {
		response.BadResponse(c, http.StatusInternalServerError, erro.RequestTimedOut, traceID, place, h.logproducer)
		return
	}
	protoresponse, err := h.photoclient.GetPhoto(ctx, userid, photoid)
	if err == nil && protoresponse != nil && protoresponse.Status {
		response.OkResponse(c, http.StatusOK, map[string]any{response.KeyPhoto: protoresponse.Photo}, traceID, place, h.logproducer)
		return
	}
	h.badGrpcResponse(c, traceID, place, err)
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
	const place = DeletePhoto
	traceID := c.MustGet("traceID").(string)
	userid := c.MustGet("userID").(string)
	photoid := c.Param("photo_id")
	md := metadata.Pairs("traceID", traceID)
	ctx := metadata.NewOutgoingContext(c.Request.Context(), md)
	err := response.CheckContext(c, place, traceID, h.logproducer)
	if err != nil {
		response.BadResponse(c, http.StatusInternalServerError, erro.RequestTimedOut, traceID, place, h.logproducer)
		return
	}
	protoresponse, err := h.photoclient.DeletePhoto(ctx, userid, photoid)
	if err == nil && protoresponse != nil && protoresponse.Status {
		response.OkResponse(c, http.StatusOK, map[string]any{response.KeyMessage: protoresponse.Message}, traceID, place, h.logproducer)
		return
	}
	h.badGrpcResponse(c, traceID, place, err)
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
	md := metadata.Pairs("traceID", traceID)
	ctx := metadata.NewOutgoingContext(c.Request.Context(), md)
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
	err = response.CheckContext(c, place, traceID, h.logproducer)
	if err != nil {
		response.BadResponse(c, http.StatusInternalServerError, erro.RequestTimedOut, traceID, place, h.logproducer)
		return
	}
	protoresponse, err := h.photoclient.LoadPhoto(ctx, userid, bytes)
	if err == nil && protoresponse != nil && protoresponse.Status {
		response.OkResponse(c, http.StatusOK, map[string]any{response.KeyMessage: protoresponse.Message, response.KeyPhotoID: protoresponse.PhotoId}, traceID, place, h.logproducer)
		return
	}
	h.badGrpcResponse(c, traceID, place, err)
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
	const place = GetMyPhotoById
	traceID := c.MustGet("traceID").(string)
	userid := c.MustGet("userID").(string)
	photoid := c.Param("photo_id")
	md := metadata.Pairs("traceID", traceID)
	ctx := metadata.NewOutgoingContext(c.Request.Context(), md)
	err := response.CheckContext(c, place, traceID, h.logproducer)
	if err != nil {
		response.BadResponse(c, http.StatusInternalServerError, erro.RequestTimedOut, traceID, place, h.logproducer)
		return
	}
	protoresponse, err := h.photoclient.GetPhoto(ctx, userid, photoid)
	if err == nil && protoresponse != nil && protoresponse.Status {
		response.OkResponse(c, http.StatusOK, map[string]any{response.KeyPhoto: protoresponse.Photo}, traceID, place, h.logproducer)
		return
	}
	h.badGrpcResponse(c, traceID, place, err)
}

func (h *Handler) badGrpcResponse(c *gin.Context, traceID, place string, err error) {
	st, _ := status.FromError(err)
	switch st.Code() {
	case codes.Canceled, codes.Unavailable:
		response.BadResponse(c, http.StatusInternalServerError, erro.PhotoServiceUnavalaible, traceID, place, h.logproducer)
		metrics.APIErrorsTotal.WithLabelValues(erro.ServerErrorType).Inc()
	case codes.Internal:
		response.BadResponse(c, http.StatusInternalServerError, st.Message(), traceID, place, h.logproducer)
		metrics.APIErrorsTotal.WithLabelValues(erro.ServerErrorType).Inc()
	default:
		response.BadResponse(c, http.StatusBadRequest, st.Message(), traceID, place, h.logproducer)
		metrics.APIErrorsTotal.WithLabelValues(erro.ClientErrorType).Inc()
	}
}
