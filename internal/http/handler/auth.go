package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/luxixing/fx-gin-scaffold/internal/domain"
	"github.com/luxixing/fx-gin-scaffold/internal/http/middleware"
	"go.uber.org/fx"
)

// AuthHandlerParams holds dependencies for AuthHandler
type AuthHandlerParams struct {
	fx.In
	UserService domain.UserService
	AuthService domain.AuthService
}

// AuthHandler handles authentication related requests
type AuthHandler struct {
	userService domain.UserService
	authService domain.AuthService
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(p AuthHandlerParams) *AuthHandler {
	return &AuthHandler{
		userService: p.UserService,
		authService: p.AuthService,
	}
}

// Register handles user registration
// @Summary Register a new user
// @Description Create a new user account
// @Tags auth
// @Accept json
// @Produce json
// @Param request body domain.UserCreateRequest true "User registration data"
// @Success 201 {object} domain.Response{data=domain.AuthResponse}
// @Failure 400 {object} domain.Response{error=domain.Error}
// @Failure 409 {object} domain.Response{error=domain.Error}
// @Failure 500 {object} domain.Response{error=domain.Error}
// @Router /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req domain.UserCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse(
			domain.NewErrorWithDetails(domain.ErrCodeValidation, "Invalid request body", err.Error()),
		))
		return
	}

	user, err := h.userService.Register(c.Request.Context(), &req)
	if err != nil {
		if domainErr, ok := err.(*domain.Error); ok {
			c.JSON(domain.HTTPStatusFromError(domainErr), domain.NewErrorResponse(domainErr))
		} else {
			c.JSON(http.StatusInternalServerError, domain.NewErrorResponse(domain.ErrInternalServer))
		}
		return
	}

	// Generate token for the new user
	token, err := h.authService.GenerateToken(&domain.User{
		ID:    user.ID,
		Email: user.Email,
		Role:  user.Role,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse(domain.ErrInternalServer))
		return
	}

	response := &domain.AuthResponse{
		Token: token,
		User:  user,
	}

	c.JSON(http.StatusCreated, domain.NewSuccessResponse(response))
}

// Login handles user authentication
// @Summary Authenticate user
// @Description Login with email and password
// @Tags auth
// @Accept json
// @Produce json
// @Param request body domain.UserLoginRequest true "User login credentials"
// @Success 200 {object} domain.Response{data=domain.AuthResponse}
// @Failure 400 {object} domain.Response{error=domain.Error}
// @Failure 401 {object} domain.Response{error=domain.Error}
// @Failure 500 {object} domain.Response{error=domain.Error}
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req domain.UserLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse(
			domain.NewErrorWithDetails(domain.ErrCodeValidation, "Invalid request body", err.Error()),
		))
		return
	}

	token, user, err := h.userService.Login(c.Request.Context(), &req)
	if err != nil {
		if domainErr, ok := err.(*domain.Error); ok {
			c.JSON(domain.HTTPStatusFromError(domainErr), domain.NewErrorResponse(domainErr))
		} else {
			c.JSON(http.StatusInternalServerError, domain.NewErrorResponse(domain.ErrInternalServer))
		}
		return
	}

	response := &domain.AuthResponse{
		Token: token,
		User:  user,
	}

	c.JSON(http.StatusOK, domain.NewSuccessResponse(response))
}

// RefreshToken handles token refresh
// @Summary Refresh JWT token
// @Description Refresh an existing JWT token
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} domain.Response{data=map[string]string}
// @Failure 401 {object} domain.Response{error=domain.Error}
// @Failure 500 {object} domain.Response{error=domain.Error}
// @Router /auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	token := middleware.ExtractToken(c)
	if token == "" {
		c.JSON(http.StatusUnauthorized, domain.NewErrorResponse(domain.ErrUnauthorized))
		return
	}

	newToken, err := h.authService.RefreshToken(c.Request.Context(), token)
	if err != nil {
		if domainErr, ok := err.(*domain.Error); ok {
			c.JSON(domain.HTTPStatusFromError(domainErr), domain.NewErrorResponse(domainErr))
		} else {
			c.JSON(http.StatusInternalServerError, domain.NewErrorResponse(domain.ErrInternalServer))
		}
		return
	}

	response := map[string]string{
		"token": newToken,
	}

	c.JSON(http.StatusOK, domain.NewSuccessResponse(response))
}

// GetProfile handles getting current user profile
// @Summary Get current user profile
// @Description Get the profile of the currently authenticated user
// @Tags auth
// @Produce json
// @Security BearerAuth
// @Success 200 {object} domain.Response{data=domain.UserResponse}
// @Failure 401 {object} domain.Response{error=domain.Error}
// @Failure 500 {object} domain.Response{error=domain.Error}
// @Router /auth/profile [get]
func (h *AuthHandler) GetProfile(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, domain.NewErrorResponse(domain.ErrUnauthorized))
		return
	}

	user, err := h.userService.GetProfile(c.Request.Context(), userID)
	if err != nil {
		if domainErr, ok := err.(*domain.Error); ok {
			c.JSON(domain.HTTPStatusFromError(domainErr), domain.NewErrorResponse(domainErr))
		} else {
			c.JSON(http.StatusInternalServerError, domain.NewErrorResponse(domain.ErrInternalServer))
		}
		return
	}

	c.JSON(http.StatusOK, domain.NewSuccessResponse(user))
}

// UpdateProfile handles updating current user profile
// @Summary Update current user profile
// @Description Update the profile of the currently authenticated user
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body domain.UserUpdateRequest true "User update data"
// @Success 200 {object} domain.Response{data=domain.UserResponse}
// @Failure 400 {object} domain.Response{error=domain.Error}
// @Failure 401 {object} domain.Response{error=domain.Error}
// @Failure 500 {object} domain.Response{error=domain.Error}
// @Router /auth/profile [put]
func (h *AuthHandler) UpdateProfile(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, domain.NewErrorResponse(domain.ErrUnauthorized))
		return
	}

	var req domain.UserUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse(
			domain.NewErrorWithDetails(domain.ErrCodeValidation, "Invalid request body", err.Error()),
		))
		return
	}

	user, err := h.userService.UpdateProfile(c.Request.Context(), userID, &req)
	if err != nil {
		if domainErr, ok := err.(*domain.Error); ok {
			c.JSON(domain.HTTPStatusFromError(domainErr), domain.NewErrorResponse(domainErr))
		} else {
			c.JSON(http.StatusInternalServerError, domain.NewErrorResponse(domain.ErrInternalServer))
		}
		return
	}

	c.JSON(http.StatusOK, domain.NewSuccessResponse(user))
}