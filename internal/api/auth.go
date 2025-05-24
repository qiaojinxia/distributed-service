package api

import (
	"context"
	"distributed-service/internal/model"
	"distributed-service/internal/service"
	"distributed-service/pkg/auth"
	"distributed-service/pkg/logger"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

// AuthHandler handles authentication requests
type AuthHandler struct {
	userService service.UserService
	jwtManager  *auth.JWTManager
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(userService service.UserService, jwtManager *auth.JWTManager) *AuthHandler {
	return &AuthHandler{
		userService: userService,
		jwtManager:  jwtManager,
	}
}

// Register godoc
// @Summary Register a new user
// @Description Register a new user account
// @Tags auth
// @Accept json
// @Produce json
// @Param request body model.RegisterRequest true "Registration request"
// @Success 201 {object} model.AuthResponse
// @Failure 400 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Router /api/v1/auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	ctx := c.MustGet("ctx").(context.Context)
	var req model.RegisterRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	user, err := h.userService.Register(ctx, &req)
	if err != nil {
		logger.Error(ctx, "Registration failed", logger.Error_(err))
		if errors.Is(err, service.ErrUserExists) {
			c.JSON(http.StatusConflict, ErrorResponse{Error: "User already exists"})
		} else {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		}
		return
	}

	// Generate JWT token
	token, err := h.jwtManager.GenerateToken(ctx, user.ID, user.Username)
	if err != nil {
		logger.Error(ctx, "Failed to generate token", logger.Error_(err))
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to generate token"})
		return
	}

	response := model.AuthResponse{
		Token: token,
		User:  *user,
	}

	c.JSON(http.StatusCreated, response)
}

// Login godoc
// @Summary User login
// @Description Authenticate user and return JWT token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body model.LoginRequest true "Login request"
// @Success 200 {object} model.AuthResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/v1/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	ctx := c.MustGet("ctx").(context.Context)
	var req model.LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	user, err := h.userService.Login(ctx, &req)
	if err != nil {
		logger.Error(ctx, "Login failed", logger.Error_(err))
		if errors.Is(err, service.ErrInvalidCredentials) {
			c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Invalid credentials"})
		} else {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		}
		return
	}

	// Generate JWT token
	token, err := h.jwtManager.GenerateToken(ctx, user.ID, user.Username)
	if err != nil {
		logger.Error(ctx, "Failed to generate token", logger.Error_(err))
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to generate token"})
		return
	}

	response := model.AuthResponse{
		Token: token,
		User:  *user,
	}

	c.JSON(http.StatusOK, response)
}

// RefreshToken godoc
// @Summary Refresh JWT token
// @Description Refresh an existing JWT token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body model.RefreshTokenRequest true "Refresh token request"
// @Success 200 {object} model.AuthResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/v1/auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	ctx := c.MustGet("ctx").(context.Context)
	var req model.RefreshTokenRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	// Refresh the token
	newToken, err := h.jwtManager.RefreshToken(ctx, req.Token)
	if err != nil {
		logger.Error(ctx, "Failed to refresh token", logger.Error_(err))
		if errors.Is(err, auth.ErrTokenExpired) || errors.Is(err, auth.ErrInvalidToken) {
			c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Invalid or expired token"})
		} else {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		}
		return
	}

	// Get user info from token
	claims, err := h.jwtManager.ValidateToken(ctx, newToken)
	if err != nil {
		logger.Error(ctx, "Failed to validate new token", logger.Error_(err))
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to validate token"})
		return
	}

	// Get user details
	user, err := h.userService.GetByID(ctx, claims.UserID)
	if err != nil {
		logger.Error(ctx, "Failed to get user", logger.Error_(err))
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to get user"})
		return
	}

	response := model.AuthResponse{
		Token: newToken,
		User:  *user,
	}

	c.JSON(http.StatusOK, response)
}

// ChangePassword godoc
// @Summary Change user password
// @Description Change the password of the authenticated user
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body ChangePasswordRequest true "Change password request"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/v1/auth/change-password [post]
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	ctx := c.MustGet("ctx").(context.Context)
	userID := c.GetUint("user_id")

	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	err := h.userService.ChangePassword(ctx, userID, req.OldPassword, req.NewPassword)
	if err != nil {
		logger.Error(ctx, "Failed to change password", logger.Error_(err))
		if errors.Is(err, service.ErrInvalidPassword) {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid old password"})
		} else {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{Message: "Password changed successfully"})
}

// ChangePasswordRequest represents a change password request
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required" example:"oldpassword123"`
	NewPassword string `json:"new_password" binding:"required,min=6" example:"newpassword123"`
}

// SuccessResponse represents a success response
type SuccessResponse struct {
	Message string `json:"message"`
}
