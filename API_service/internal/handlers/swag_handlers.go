package handlers

import "github.com/gin-gonic/gin"

//@Summary Register a new user
//@Description Register a new user by sending user data to the target service.
//@Tags User Management
//@Accept json
//@Produce json
//@Param input body PersonReg true "User registration data"
//@Success 200 {object} map[string]string "User successfully registered"
//@Failure 400 {object} map[string]string "Invalid input data"
//@Failure 403 {object} map[string]string "Forbidden"
//@Failure 500 {object} map[string]string "Internal server error"
//@Router /api/reg [post]
func (h *Handler) Registration(c *gin.Context) {
	h.ProxyHTTP(c)
}

// @Summary Authenticate a user
// @Description Authenticate a user by sending credentials to the target service.
// @Tags User Management
// @Accept json
// @Produce json
// @Param input body PersonAuth true "User credentials"
// @Success 200 {object} HTTPResponse "Authentication successful"
// @Failure 400 {object} HTTPResponse "Invalid input data"
// @Failure 403 {object} HTTPResponse "Forbidden"
// @Failure 500 {object} HTTPResponse "Internal server error"
// @Router /api/auth [post]
func (h *Handler) Authentication(c *gin.Context) {
	h.ProxyHTTP(c)
}

// @Summary Delete a user
// @Description Delete a user by sending a DELETE request with session and user ID.
// @Tags User Management
// @Accept json
// @Produce json
// @Param session cookie string true "Session ID for authorization"
// @Param input body PersonDelete true "User credentials"
// @Success 200 {object} HTTPResponse "User successfully deleted"
// @Failure 400 {object} HTTPResponse"Invalid input data"
// @Failure 401 {object} HTTPResponse "Unathorized"
// @Failure 500 {object} HTTPResponse "Internal server error"
// @Router /api/del [delete]
func (h *Handler) DeleteUser(c *gin.Context) {
	h.ProxyHTTP(c)
}
