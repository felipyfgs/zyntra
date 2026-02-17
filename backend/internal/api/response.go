package api

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// APIResponse is the standard response structure for all API endpoints
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *APIError   `json:"error,omitempty"`
	Meta    *Meta       `json:"meta,omitempty"`
}

// APIError represents an error response
type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// Meta contains pagination and other metadata
type Meta struct {
	Page       int   `json:"page,omitempty"`
	PerPage    int   `json:"per_page,omitempty"`
	Total      int64 `json:"total,omitempty"`
	TotalPages int   `json:"total_pages,omitempty"`
}

// Common error codes
const (
	ErrCodeBadRequest      = "BAD_REQUEST"
	ErrCodeUnauthorized    = "UNAUTHORIZED"
	ErrCodeForbidden       = "FORBIDDEN"
	ErrCodeNotFound        = "NOT_FOUND"
	ErrCodeConflict        = "CONFLICT"
	ErrCodeValidation      = "VALIDATION_ERROR"
	ErrCodeInternal        = "INTERNAL_ERROR"
	ErrCodeRateLimited     = "RATE_LIMITED"
	ErrCodeInvalidAPIKey   = "INVALID_API_KEY"
	ErrCodeExpiredToken    = "EXPIRED_TOKEN"
	ErrCodeInvalidToken    = "INVALID_TOKEN"
)

// Success sends a successful response with data
func Success(c echo.Context, data interface{}) error {
	return c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data:    data,
	})
}

// SuccessWithMeta sends a successful response with data and metadata
func SuccessWithMeta(c echo.Context, data interface{}, meta *Meta) error {
	return c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data:    data,
		Meta:    meta,
	})
}

// Created sends a 201 response with created resource
func Created(c echo.Context, data interface{}) error {
	return c.JSON(http.StatusCreated, APIResponse{
		Success: true,
		Data:    data,
	})
}

// NoContent sends a 204 response
func NoContent(c echo.Context) error {
	return c.NoContent(http.StatusNoContent)
}

// Error sends an error response
func Error(c echo.Context, status int, code, message string) error {
	return c.JSON(status, APIResponse{
		Success: false,
		Error: &APIError{
			Code:    code,
			Message: message,
		},
	})
}

// ErrorWithDetails sends an error response with additional details
func ErrorWithDetails(c echo.Context, status int, code, message, details string) error {
	return c.JSON(status, APIResponse{
		Success: false,
		Error: &APIError{
			Code:    code,
			Message: message,
			Details: details,
		},
	})
}

// BadRequest sends a 400 response
func BadRequest(c echo.Context, message string) error {
	return Error(c, http.StatusBadRequest, ErrCodeBadRequest, message)
}

// Unauthorized sends a 401 response
func Unauthorized(c echo.Context, message string) error {
	return Error(c, http.StatusUnauthorized, ErrCodeUnauthorized, message)
}

// Forbidden sends a 403 response
func Forbidden(c echo.Context, message string) error {
	return Error(c, http.StatusForbidden, ErrCodeForbidden, message)
}

// NotFound sends a 404 response
func NotFound(c echo.Context, message string) error {
	return Error(c, http.StatusNotFound, ErrCodeNotFound, message)
}

// Conflict sends a 409 response
func Conflict(c echo.Context, message string) error {
	return Error(c, http.StatusConflict, ErrCodeConflict, message)
}

// ValidationError sends a 422 response
func ValidationError(c echo.Context, message string) error {
	return Error(c, http.StatusUnprocessableEntity, ErrCodeValidation, message)
}

// InternalError sends a 500 response
func InternalError(c echo.Context, message string) error {
	return Error(c, http.StatusInternalServerError, ErrCodeInternal, message)
}

// RateLimited sends a 429 response
func RateLimited(c echo.Context) error {
	return Error(c, http.StatusTooManyRequests, ErrCodeRateLimited, "Rate limit exceeded")
}

// CalculateTotalPages calculates total pages for pagination
func CalculateTotalPages(total int64, perPage int) int {
	if perPage <= 0 {
		return 0
	}
	pages := int(total) / perPage
	if int(total)%perPage > 0 {
		pages++
	}
	return pages
}

// NewMeta creates a new Meta with pagination info
func NewMeta(page, perPage int, total int64) *Meta {
	return &Meta{
		Page:       page,
		PerPage:    perPage,
		Total:      total,
		TotalPages: CalculateTotalPages(total, perPage),
	}
}
