package api

import (
	"context"
	"distributed-service/internal/model"
	"distributed-service/internal/service"
	"distributed-service/pkg/logger"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// UserHandler handles HTTP requests for users
type UserHandler struct {
	userService service.UserService
}

// NewUserHandler creates a new user handler
func NewUserHandler(userService service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

// Create godoc
// @Summary Create user
// @Description Create a new user
// @Tags users
// @Accept json
// @Produce json
// @Param user body model.User true "User object"
// @Success 201 {object} model.User
// @Failure 400 {object} ErrorResponse
// @Router /api/v1/users [post]
func (h *UserHandler) Create(c *gin.Context) {
	ctx := c.MustGet("ctx").(context.Context)
	var user model.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	if err := h.userService.Create(ctx, &user); err != nil {
		logger.Error(ctx, "Failed to create user", logger.Error_(err))
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, user)
}

// GetByID godoc
// @Summary Get user by ID
// @Description Get user details by ID
// @Tags users
// @Produce json
// @Param id path int true "User ID"
// @Success 200 {object} model.User
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/users/{id} [get]
func (h *UserHandler) GetByID(c *gin.Context) {
	ctx := c.MustGet("ctx").(context.Context)
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	user, err := h.userService.GetByID(ctx, uint(id))
	if err != nil {
		logger.Error(ctx, "Failed to get user", logger.Error_(err))
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "User not found"})
		return
	}

	c.JSON(http.StatusOK, user)
}

// Delete godoc
// @Summary Delete user
// @Description Delete user by ID
// @Tags users
// @Produce json
// @Param id path int true "User ID"
// @Success 204 "No Content"
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/users/{id} [delete]
func (h *UserHandler) Delete(c *gin.Context) {
	ctx := c.MustGet("ctx").(context.Context)
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	if err := h.userService.Delete(ctx, uint(id)); err != nil {
		logger.Error(ctx, "Failed to delete user", logger.Error_(err))
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "User not found"})
		return
	}

	c.Status(http.StatusNoContent)
}

// GetMe godoc
// @Summary Get current user info
// @Description Get the profile information of the authenticated user
// @Tags users
// @Produce json
// @Security BearerAuth
// @Success 200 {object} model.User
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/users/me [get]
func (h *UserHandler) GetMe(c *gin.Context) {
	ctx := c.MustGet("ctx").(context.Context)
	userID := c.GetUint("user_id")

	user, err := h.userService.GetByID(ctx, userID)
	if err != nil {
		logger.Error(ctx, "Failed to get current user", logger.Error_(err))
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "User not found"})
		return
	}

	c.JSON(http.StatusOK, user)
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error string `json:"error"`
}
