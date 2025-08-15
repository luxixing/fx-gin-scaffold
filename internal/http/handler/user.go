package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/luxixing/fx-gin-scaffold/internal/domain"
	"github.com/luxixing/fx-gin-scaffold/internal/http/middleware"
	"go.uber.org/fx"
)

// UserHandlerParams holds dependencies for UserHandler
type UserHandlerParams struct {
	fx.In
	UserService domain.UserService
}

// UserHandler handles user management requests
type UserHandler struct {
	userService domain.UserService
}

// NewUserHandler creates a new user handler
func NewUserHandler(p UserHandlerParams) *UserHandler {
	return &UserHandler{
		userService: p.UserService,
	}
}

// ListUsers handles listing users with pagination
// @Summary List users
// @Description Get a paginated list of users (admin only)
// @Tags users
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} domain.Response{data=[]domain.UserResponse,meta=domain.Meta}
// @Failure 401 {object} domain.Response{error=domain.Error}
// @Failure 403 {object} domain.Response{error=domain.Error}
// @Failure 500 {object} domain.Response{error=domain.Error}
// @Router /users [get]
func (h *UserHandler) ListUsers(c *gin.Context) {
	var pagination domain.PaginationRequest
	if err := c.ShouldBindQuery(&pagination); err != nil {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse(
			domain.NewErrorWithDetails(domain.ErrCodeValidation, "Invalid pagination parameters", err.Error()),
		))
		return
	}

	users, total, err := h.userService.ListUsers(c.Request.Context(), pagination.GetOffset(), pagination.Limit)
	if err != nil {
		if domainErr, ok := err.(*domain.Error); ok {
			c.JSON(domain.HTTPStatusFromError(domainErr), domain.NewErrorResponse(domainErr))
		} else {
			c.JSON(http.StatusInternalServerError, domain.NewErrorResponse(domain.ErrInternalServer))
		}
		return
	}

	meta := pagination.GetMeta(total)
	c.JSON(http.StatusOK, domain.NewSuccessResponseWithMeta(users, meta))
}

// SearchUsers handles searching users
// @Summary Search users
// @Description Search users by name or email (admin only)
// @Tags users
// @Produce json
// @Security BearerAuth
// @Param q query string true "Search query"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} domain.Response{data=[]domain.UserResponse,meta=domain.Meta}
// @Failure 400 {object} domain.Response{error=domain.Error}
// @Failure 401 {object} domain.Response{error=domain.Error}
// @Failure 403 {object} domain.Response{error=domain.Error}
// @Failure 500 {object} domain.Response{error=domain.Error}
// @Router /users/search [get]
func (h *UserHandler) SearchUsers(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse(
			domain.ValidationError("q", "search query is required"),
		))
		return
	}

	var pagination domain.PaginationRequest
	if err := c.ShouldBindQuery(&pagination); err != nil {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse(
			domain.NewErrorWithDetails(domain.ErrCodeValidation, "Invalid pagination parameters", err.Error()),
		))
		return
	}

	users, total, err := h.userService.SearchUsers(c.Request.Context(), query, pagination.GetOffset(), pagination.Limit)
	if err != nil {
		if domainErr, ok := err.(*domain.Error); ok {
			c.JSON(domain.HTTPStatusFromError(domainErr), domain.NewErrorResponse(domainErr))
		} else {
			c.JSON(http.StatusInternalServerError, domain.NewErrorResponse(domain.ErrInternalServer))
		}
		return
	}

	meta := pagination.GetMeta(total)
	c.JSON(http.StatusOK, domain.NewSuccessResponseWithMeta(users, meta))
}

// GetUser handles getting a specific user
// @Summary Get user by ID
// @Description Get a user by their ID (admin only)
// @Tags users
// @Produce json
// @Security BearerAuth
// @Param id path int true "User ID"
// @Success 200 {object} domain.Response{data=domain.UserResponse}
// @Failure 400 {object} domain.Response{error=domain.Error}
// @Failure 401 {object} domain.Response{error=domain.Error}
// @Failure 403 {object} domain.Response{error=domain.Error}
// @Failure 404 {object} domain.Response{error=domain.Error}
// @Failure 500 {object} domain.Response{error=domain.Error}
// @Router /users/{id} [get]
func (h *UserHandler) GetUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse(
			domain.ValidationError("id", "must be a valid number"),
		))
		return
	}

	user, err := h.userService.GetUser(c.Request.Context(), uint(id))
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

// UpdateUser handles updating a user
// @Summary Update user
// @Description Update a user's information (admin only)
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "User ID"
// @Param request body domain.UserUpdateRequest true "User update data"
// @Success 200 {object} domain.Response{data=domain.UserResponse}
// @Failure 400 {object} domain.Response{error=domain.Error}
// @Failure 401 {object} domain.Response{error=domain.Error}
// @Failure 403 {object} domain.Response{error=domain.Error}
// @Failure 404 {object} domain.Response{error=domain.Error}
// @Failure 500 {object} domain.Response{error=domain.Error}
// @Router /users/{id} [put]
func (h *UserHandler) UpdateUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse(
			domain.ValidationError("id", "must be a valid number"),
		))
		return
	}

	// Prevent users from updating themselves through admin endpoint
	userID, _ := middleware.GetUserID(c)
	if userID == uint(id) {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse(
			domain.NewError(domain.ErrCodeInvalid, "Cannot update your own account through admin endpoint"),
		))
		return
	}

	var req domain.UserUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse(
			domain.NewErrorWithDetails(domain.ErrCodeValidation, "Invalid request body", err.Error()),
		))
		return
	}

	user, err := h.userService.UpdateUser(c.Request.Context(), uint(id), &req)
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

// DeleteUser handles deleting a user
// @Summary Delete user
// @Description Delete a user account (admin only)
// @Tags users
// @Produce json
// @Security BearerAuth
// @Param id path int true "User ID"
// @Success 204 "User deleted successfully"
// @Failure 400 {object} domain.Response{error=domain.Error}
// @Failure 401 {object} domain.Response{error=domain.Error}
// @Failure 403 {object} domain.Response{error=domain.Error}
// @Failure 404 {object} domain.Response{error=domain.Error}
// @Failure 500 {object} domain.Response{error=domain.Error}
// @Router /users/{id} [delete]
func (h *UserHandler) DeleteUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse(
			domain.ValidationError("id", "must be a valid number"),
		))
		return
	}

	// Prevent users from deleting themselves
	userID, _ := middleware.GetUserID(c)
	if userID == uint(id) {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse(
			domain.NewError(domain.ErrCodeInvalid, "Cannot delete your own account"),
		))
		return
	}

	err = h.userService.DeleteUser(c.Request.Context(), uint(id))
	if err != nil {
		if domainErr, ok := err.(*domain.Error); ok {
			c.JSON(domain.HTTPStatusFromError(domainErr), domain.NewErrorResponse(domainErr))
		} else {
			c.JSON(http.StatusInternalServerError, domain.NewErrorResponse(domain.ErrInternalServer))
		}
		return
	}

	c.Status(http.StatusNoContent)
}