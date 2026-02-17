package handlers

import (
	"database/sql"

	"github.com/labstack/echo/v4"
	"github.com/zyntra/backend/internal/api"
	"github.com/zyntra/backend/internal/middleware"
	"github.com/zyntra/backend/internal/repository"
)

// FilterHandler handles saved filter endpoints
type FilterHandler struct {
	repo *repository.FilterRepository
}

// NewFilterHandler creates a new filter handler
func NewFilterHandler(db *sql.DB) *FilterHandler {
	return &FilterHandler{
		repo: repository.NewFilterRepository(db),
	}
}

// GetFilters returns all saved filters for the user
func (h *FilterHandler) GetFilters(c echo.Context) error {
	user := middleware.GetUser(c)
	if user == nil {
		return api.Unauthorized(c, "Not authenticated")
	}

	filters, err := h.repo.List(c.Request().Context(), user.UserID)
	if err != nil {
		return api.InternalError(c, "Failed to get filters")
	}

	if filters == nil {
		filters = []*repository.SavedFilter{}
	}

	return api.Success(c, filters)
}

// CreateFilterRequest represents the create filter request body
type CreateFilterRequest struct {
	Name  string                  `json:"name" validate:"required"`
	Rules []repository.FilterRule `json:"rules"`
}

// CreateFilter creates a new saved filter
func (h *FilterHandler) CreateFilter(c echo.Context) error {
	user := middleware.GetUser(c)
	if user == nil {
		return api.Unauthorized(c, "Not authenticated")
	}

	var req CreateFilterRequest
	if err := c.Bind(&req); err != nil {
		return api.BadRequest(c, "Invalid request body")
	}

	if req.Name == "" {
		return api.ValidationError(c, "Name is required")
	}

	filter := &repository.SavedFilter{
		UserID: user.UserID,
		Name:   req.Name,
		Rules:  req.Rules,
	}

	if filter.Rules == nil {
		filter.Rules = []repository.FilterRule{}
	}

	if err := h.repo.Create(c.Request().Context(), filter); err != nil {
		return api.InternalError(c, "Failed to create filter")
	}

	return api.Created(c, filter)
}

// UpdateFilterRequest represents the update filter request body
type UpdateFilterRequest struct {
	Name  string                  `json:"name,omitempty"`
	Rules []repository.FilterRule `json:"rules,omitempty"`
}

// UpdateFilter updates a saved filter
func (h *FilterHandler) UpdateFilter(c echo.Context) error {
	user := middleware.GetUser(c)
	if user == nil {
		return api.Unauthorized(c, "Not authenticated")
	}

	filterID := c.Param("id")
	if filterID == "" {
		return api.BadRequest(c, "Filter ID is required")
	}

	var req UpdateFilterRequest
	if err := c.Bind(&req); err != nil {
		return api.BadRequest(c, "Invalid request body")
	}

	// Get existing filter
	filter, err := h.repo.GetByID(c.Request().Context(), user.UserID, filterID)
	if err != nil {
		return api.InternalError(c, "Failed to get filter")
	}
	if filter == nil {
		return api.NotFound(c, "Filter not found")
	}

	// Update fields
	if req.Name != "" {
		filter.Name = req.Name
	}
	if req.Rules != nil {
		filter.Rules = req.Rules
	}

	if err := h.repo.Update(c.Request().Context(), filter); err != nil {
		return api.InternalError(c, "Failed to update filter")
	}

	return api.Success(c, filter)
}

// DeleteFilter deletes a saved filter
func (h *FilterHandler) DeleteFilter(c echo.Context) error {
	user := middleware.GetUser(c)
	if user == nil {
		return api.Unauthorized(c, "Not authenticated")
	}

	filterID := c.Param("id")
	if filterID == "" {
		return api.BadRequest(c, "Filter ID is required")
	}

	err := h.repo.Delete(c.Request().Context(), user.UserID, filterID)
	if err != nil {
		if err == sql.ErrNoRows {
			return api.NotFound(c, "Filter not found")
		}
		return api.InternalError(c, "Failed to delete filter")
	}

	return api.NoContent(c)
}

// ReorderFiltersRequest represents the reorder filters request body
type ReorderFiltersRequest struct {
	FilterIDs []string `json:"filter_ids" validate:"required"`
}

// ReorderFilters updates the position of all filters
func (h *FilterHandler) ReorderFilters(c echo.Context) error {
	user := middleware.GetUser(c)
	if user == nil {
		return api.Unauthorized(c, "Not authenticated")
	}

	var req ReorderFiltersRequest
	if err := c.Bind(&req); err != nil {
		return api.BadRequest(c, "Invalid request body")
	}

	if len(req.FilterIDs) == 0 {
		return api.BadRequest(c, "Filter IDs are required")
	}

	if err := h.repo.Reorder(c.Request().Context(), user.UserID, req.FilterIDs); err != nil {
		return api.InternalError(c, "Failed to reorder filters")
	}

	// Return updated list
	filters, _ := h.repo.List(c.Request().Context(), user.UserID)
	return api.Success(c, filters)
}
