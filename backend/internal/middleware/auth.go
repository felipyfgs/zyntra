package middleware

import (
	"context"
	"database/sql"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/zyntra/backend/internal/api"
	"github.com/zyntra/backend/internal/auth"
)

// ContextKey type for context keys
type ContextKey string

const (
	// ContextKeyUser is the key for user data in context
	ContextKeyUser ContextKey = "user"
	// ContextKeyAPIKey is the key for API key data in context
	ContextKeyAPIKey ContextKey = "api_key"
	// ContextKeyAuthType is the key for auth type in context
	ContextKeyAuthType ContextKey = "auth_type"
)

// AuthType represents the authentication method used
type AuthType string

const (
	AuthTypeJWT    AuthType = "jwt"
	AuthTypeAPIKey AuthType = "api_key"
)

// UserContext holds user information in request context
type UserContext struct {
	UserID string
	Email  string
	Role   string
}

// AuthMiddleware handles JWT and API Key authentication
type AuthMiddleware struct {
	jwtService    *auth.JWTService
	apiKeyService *auth.APIKeyService
}

// NewAuthMiddleware creates a new auth middleware
func NewAuthMiddleware(jwtService *auth.JWTService, db *sql.DB) *AuthMiddleware {
	return &AuthMiddleware{
		jwtService:    jwtService,
		apiKeyService: auth.NewAPIKeyService(db),
	}
}

// Authenticate is the main authentication middleware
// Supports both JWT tokens (Authorization: Bearer <token>) and API Keys (X-API-Key: <key>)
func (m *AuthMiddleware) Authenticate(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Try API Key first (X-API-Key header)
		apiKey := c.Request().Header.Get("X-API-Key")
		if apiKey != "" {
			return m.authenticateAPIKey(c, next, apiKey)
		}

		// Try JWT token (Authorization: Bearer <token>)
		authHeader := c.Request().Header.Get("Authorization")
		if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
			token := strings.TrimPrefix(authHeader, "Bearer ")
			return m.authenticateJWT(c, next, token)
		}

		return api.Unauthorized(c, "Authentication required")
	}
}

// authenticateJWT validates JWT token
func (m *AuthMiddleware) authenticateJWT(c echo.Context, next echo.HandlerFunc, token string) error {
	claims, err := m.jwtService.ValidateAccessToken(token)
	if err != nil {
		if err == auth.ErrExpiredToken {
			return api.Error(c, 401, api.ErrCodeExpiredToken, "Token has expired")
		}
		return api.Error(c, 401, api.ErrCodeInvalidToken, "Invalid token")
	}

	// Set user context
	userCtx := &UserContext{
		UserID: claims.UserID,
		Email:  claims.Email,
		Role:   claims.Role,
	}

	ctx := context.WithValue(c.Request().Context(), ContextKeyUser, userCtx)
	ctx = context.WithValue(ctx, ContextKeyAuthType, AuthTypeJWT)
	c.SetRequest(c.Request().WithContext(ctx))

	return next(c)
}

// authenticateAPIKey validates API key
func (m *AuthMiddleware) authenticateAPIKey(c echo.Context, next echo.HandlerFunc, key string) error {
	apiKey, err := m.apiKeyService.ValidateAPIKey(c.Request().Context(), key)
	if err != nil {
		switch err {
		case auth.ErrAPIKeyNotFound:
			return api.Error(c, 401, api.ErrCodeInvalidAPIKey, "Invalid API key")
		case auth.ErrAPIKeyExpired:
			return api.Error(c, 401, api.ErrCodeExpiredToken, "API key has expired")
		case auth.ErrAPIKeyRevoked:
			return api.Error(c, 401, api.ErrCodeInvalidAPIKey, "API key has been revoked")
		default:
			return api.InternalError(c, "Authentication error")
		}
	}

	// Set API key and user context
	userCtx := &UserContext{
		UserID: apiKey.UserID,
	}

	ctx := context.WithValue(c.Request().Context(), ContextKeyUser, userCtx)
	ctx = context.WithValue(ctx, ContextKeyAPIKey, apiKey)
	ctx = context.WithValue(ctx, ContextKeyAuthType, AuthTypeAPIKey)
	c.SetRequest(c.Request().WithContext(ctx))

	return next(c)
}

// RequirePermission middleware checks if API key has required permission
func (m *AuthMiddleware) RequirePermission(permission string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authType := GetAuthType(c)

			// JWT users have all permissions (they are authenticated users)
			if authType == AuthTypeJWT {
				return next(c)
			}

			// Check API key permissions
			apiKey := GetAPIKey(c)
			if apiKey == nil {
				return api.Forbidden(c, "API key required")
			}

			if !apiKey.HasPermission(permission) {
				return api.Forbidden(c, "Insufficient permissions")
			}

			return next(c)
		}
	}
}

// RequireRole middleware checks if user has required role
func RequireRole(roles ...string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			user := GetUser(c)
			if user == nil {
				return api.Unauthorized(c, "Authentication required")
			}

			for _, role := range roles {
				if user.Role == role {
					return next(c)
				}
			}

			return api.Forbidden(c, "Insufficient role")
		}
	}
}

// GetUser returns the authenticated user from context
func GetUser(c echo.Context) *UserContext {
	user, ok := c.Request().Context().Value(ContextKeyUser).(*UserContext)
	if !ok {
		return nil
	}
	return user
}

// GetAPIKey returns the API key from context
func GetAPIKey(c echo.Context) *auth.APIKey {
	apiKey, ok := c.Request().Context().Value(ContextKeyAPIKey).(*auth.APIKey)
	if !ok {
		return nil
	}
	return apiKey
}

// GetAuthType returns the authentication type from context
func GetAuthType(c echo.Context) AuthType {
	authType, ok := c.Request().Context().Value(ContextKeyAuthType).(AuthType)
	if !ok {
		return ""
	}
	return authType
}
