package handlers

import (
	"context"
	"database/sql"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/zyntra/backend/internal/api"
	"github.com/zyntra/backend/internal/auth"
	"github.com/zyntra/backend/internal/middleware"
	"golang.org/x/crypto/bcrypt"
)

// AuthHandler handles authentication endpoints
type AuthHandler struct {
	db         *sql.DB
	jwtService *auth.JWTService
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(db *sql.DB, jwtService *auth.JWTService) *AuthHandler {
	return &AuthHandler{
		db:         db,
		jwtService: jwtService,
	}
}

// LoginRequest represents the login request body
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

// RegisterRequest represents the registration request body
type RegisterRequest struct {
	Name     string `json:"name" validate:"required,min=2"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

// RefreshRequest represents the token refresh request body
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// AuthResponse represents the authentication response
type AuthResponse struct {
	User   UserResponse     `json:"user"`
	Tokens *auth.TokenPair  `json:"tokens"`
}

// UserResponse represents user data in response
type UserResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

// Login handles user login
func (h *AuthHandler) Login(c echo.Context) error {
	var req LoginRequest
	if err := c.Bind(&req); err != nil {
		return api.BadRequest(c, "Invalid request body")
	}

	// Find user by email
	var user struct {
		ID           string
		Name         string
		Email        string
		PasswordHash string
		Role         string
		CreatedAt    time.Time
	}

	query := `SELECT id, name, email, password_hash, role, created_at FROM users WHERE email = $1`
	err := h.db.QueryRowContext(c.Request().Context(), query, req.Email).Scan(
		&user.ID, &user.Name, &user.Email, &user.PasswordHash, &user.Role, &user.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return api.Unauthorized(c, "Invalid email or password")
		}
		return api.InternalError(c, "Database error")
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return api.Unauthorized(c, "Invalid email or password")
	}

	// Generate tokens
	tokens, err := h.jwtService.GenerateTokenPair(user.ID, user.Email, user.Role)
	if err != nil {
		return api.InternalError(c, "Failed to generate tokens")
	}

	return api.Success(c, AuthResponse{
		User: UserResponse{
			ID:        user.ID,
			Name:      user.Name,
			Email:     user.Email,
			Role:      user.Role,
			CreatedAt: user.CreatedAt,
		},
		Tokens: tokens,
	})
}

// Register handles user registration
func (h *AuthHandler) Register(c echo.Context) error {
	var req RegisterRequest
	if err := c.Bind(&req); err != nil {
		return api.BadRequest(c, "Invalid request body")
	}

	// Validate fields
	if req.Name == "" || req.Email == "" || req.Password == "" {
		return api.ValidationError(c, "Name, email and password are required")
	}

	if len(req.Password) < 6 {
		return api.ValidationError(c, "Password must be at least 6 characters")
	}

	// Check if email already exists
	var exists bool
	err := h.db.QueryRowContext(c.Request().Context(), "SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)", req.Email).Scan(&exists)
	if err != nil {
		return api.InternalError(c, "Database error")
	}
	if exists {
		return api.Conflict(c, "Email already registered")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return api.InternalError(c, "Failed to hash password")
	}

	// Insert user
	var userID string
	var createdAt time.Time
	query := `INSERT INTO users (name, email, password_hash, role) VALUES ($1, $2, $3, $4) RETURNING id, created_at`
	err = h.db.QueryRowContext(c.Request().Context(), query, req.Name, req.Email, string(hashedPassword), "operator").Scan(&userID, &createdAt)
	if err != nil {
		return api.InternalError(c, "Failed to create user")
	}

	// Generate tokens
	tokens, err := h.jwtService.GenerateTokenPair(userID, req.Email, "operator")
	if err != nil {
		return api.InternalError(c, "Failed to generate tokens")
	}

	return api.Created(c, AuthResponse{
		User: UserResponse{
			ID:        userID,
			Name:      req.Name,
			Email:     req.Email,
			Role:      "operator",
			CreatedAt: createdAt,
		},
		Tokens: tokens,
	})
}

// RefreshToken handles token refresh
func (h *AuthHandler) RefreshToken(c echo.Context) error {
	var req RefreshRequest
	if err := c.Bind(&req); err != nil {
		return api.BadRequest(c, "Invalid request body")
	}

	tokens, err := h.jwtService.RefreshAccessToken(req.RefreshToken)
	if err != nil {
		if err == auth.ErrExpiredToken {
			return api.Error(c, 401, api.ErrCodeExpiredToken, "Refresh token has expired")
		}
		return api.Error(c, 401, api.ErrCodeInvalidToken, "Invalid refresh token")
	}

	return api.Success(c, map[string]interface{}{
		"tokens": tokens,
	})
}

// GetProfile returns the current user's profile
func (h *AuthHandler) GetProfile(c echo.Context) error {
	user := middleware.GetUser(c)
	if user == nil {
		return api.Unauthorized(c, "Not authenticated")
	}

	var profile UserResponse
	query := `SELECT id, name, email, role, created_at FROM users WHERE id = $1`
	err := h.db.QueryRowContext(c.Request().Context(), query, user.UserID).Scan(
		&profile.ID, &profile.Name, &profile.Email, &profile.Role, &profile.CreatedAt,
	)
	if err != nil {
		return api.InternalError(c, "Failed to get profile")
	}

	return api.Success(c, profile)
}

// ChangePassword handles password change
func (h *AuthHandler) ChangePassword(c echo.Context) error {
	user := middleware.GetUser(c)
	if user == nil {
		return api.Unauthorized(c, "Not authenticated")
	}

	var req struct {
		CurrentPassword string `json:"current_password"`
		NewPassword     string `json:"new_password"`
	}
	if err := c.Bind(&req); err != nil {
		return api.BadRequest(c, "Invalid request body")
	}

	if len(req.NewPassword) < 6 {
		return api.ValidationError(c, "New password must be at least 6 characters")
	}

	// Get current password hash
	var currentHash string
	err := h.db.QueryRowContext(c.Request().Context(), "SELECT password_hash FROM users WHERE id = $1", user.UserID).Scan(&currentHash)
	if err != nil {
		return api.InternalError(c, "Database error")
	}

	// Verify current password
	if err := bcrypt.CompareHashAndPassword([]byte(currentHash), []byte(req.CurrentPassword)); err != nil {
		return api.BadRequest(c, "Current password is incorrect")
	}

	// Hash new password
	newHash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return api.InternalError(c, "Failed to hash password")
	}

	// Update password
	_, err = h.db.ExecContext(c.Request().Context(), "UPDATE users SET password_hash = $1 WHERE id = $2", string(newHash), user.UserID)
	if err != nil {
		return api.InternalError(c, "Failed to update password")
	}

	return api.Success(c, map[string]string{"message": "Password changed successfully"})
}

// APIKeyHandler handles API key management
type APIKeyHandler struct {
	service *auth.APIKeyService
}

// NewAPIKeyHandler creates a new API key handler
func NewAPIKeyHandler(db *sql.DB) *APIKeyHandler {
	return &APIKeyHandler{
		service: auth.NewAPIKeyService(db),
	}
}

// CreateAPIKeyRequest represents the create API key request
type CreateAPIKeyRequest struct {
	Name        string   `json:"name" validate:"required"`
	Permissions []string `json:"permissions"`
	ExpiresIn   *int     `json:"expires_in_days"` // Days until expiration, nil = never
}

// CreateAPIKey creates a new API key
func (h *APIKeyHandler) CreateAPIKey(c echo.Context) error {
	user := middleware.GetUser(c)
	if user == nil {
		return api.Unauthorized(c, "Not authenticated")
	}

	var req CreateAPIKeyRequest
	if err := c.Bind(&req); err != nil {
		return api.BadRequest(c, "Invalid request body")
	}

	if req.Name == "" {
		return api.ValidationError(c, "Name is required")
	}

	// Default to all permissions if none specified
	permissions := req.Permissions
	if len(permissions) == 0 {
		permissions = auth.AllPermissions()
	}

	// Calculate expiration
	var expiresIn *time.Duration
	if req.ExpiresIn != nil {
		d := time.Duration(*req.ExpiresIn) * 24 * time.Hour
		expiresIn = &d
	}

	apiKey, err := h.service.GenerateAPIKey(c.Request().Context(), user.UserID, req.Name, permissions, expiresIn)
	if err != nil {
		return api.InternalError(c, "Failed to create API key")
	}

	return api.Created(c, apiKey)
}

// ListAPIKeys returns all API keys for the user
func (h *APIKeyHandler) ListAPIKeys(c echo.Context) error {
	user := middleware.GetUser(c)
	if user == nil {
		return api.Unauthorized(c, "Not authenticated")
	}

	keys, err := h.service.ListAPIKeys(c.Request().Context(), user.UserID)
	if err != nil {
		return api.InternalError(c, "Failed to list API keys")
	}

	return api.Success(c, keys)
}

// RevokeAPIKey revokes an API key
func (h *APIKeyHandler) RevokeAPIKey(c echo.Context) error {
	user := middleware.GetUser(c)
	if user == nil {
		return api.Unauthorized(c, "Not authenticated")
	}

	keyID := c.Param("id")
	if keyID == "" {
		return api.BadRequest(c, "API key ID is required")
	}

	err := h.service.RevokeAPIKey(context.Background(), keyID, user.UserID)
	if err != nil {
		if err == auth.ErrAPIKeyNotFound {
			return api.NotFound(c, "API key not found")
		}
		return api.InternalError(c, "Failed to revoke API key")
	}

	return api.NoContent(c)
}
