package handlers

import (
	"github.com/gin-gonic/gin"
	_ "github.com/niktin06sash/MicroserviceProject/API_service/internal/handlers/response"
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

// @Summary Find user by ID in path parameters
// @Description Retrieves a user's profile data by sending a request with the user's ID as a path parameter.
// @Tags User Management
// @Produce json
// @Param id path string true "User ID in UUID format"
// @Success 200 {object} response.HTTPResponse "User profile data successfully retrieved"
// @Failure 400 {object} response.HTTPResponse "Invalid input data"
// @Failure 401 {object} response.HTTPResponse "Unauthorized"
// @Failure 429 {object} response.HTTPResponse "Too many requests"
// @Failure 500 {object} response.HTTPResponse "Internal server error"
// @Router /api/users/id/{id} [get]
func (h *Handler) GetUserById(c *gin.Context) {

}

func (h *Handler) GetPhotoById(c *gin.Context) {
}

func (h *Handler) DeletePhoto(c *gin.Context) {
}

func (h *Handler) LoadPhoto(c *gin.Context) {
}
func (h *Handler) GetMyPhotoById(c *gin.Context) {
}
