package handlers

import (
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/zyntra/backend/internal/api"
	"github.com/zyntra/backend/internal/domain"
	"github.com/zyntra/backend/internal/repository"
)

// LabelHandler handler de labels
type LabelHandler struct {
	repo *repository.LabelRepository
}

// NewLabelHandler cria novo handler
func NewLabelHandler(repo *repository.LabelRepository) *LabelHandler {
	return &LabelHandler{repo: repo}
}

// List lista todos os labels
func (h *LabelHandler) List(c echo.Context) error {
	labels, err := h.repo.GetAll(c.Request().Context())
	if err != nil {
		return api.InternalError(c, err.Error())
	}
	return api.Success(c, labels)
}

// CreateLabelRequest request para criar label
type CreateLabelRequest struct {
	Title       string `json:"title" validate:"required"`
	Color       string `json:"color,omitempty"`
	Description string `json:"description,omitempty"`
}

// Create cria um novo label
func (h *LabelHandler) Create(c echo.Context) error {
	var req CreateLabelRequest
	if err := c.Bind(&req); err != nil {
		return api.BadRequest(c, "Invalid request body")
	}

	if req.Title == "" {
		return api.ValidationError(c, "Title is required")
	}

	color := req.Color
	if color == "" {
		color = "#1f93ff"
	}

	label := &domain.Label{
		ID:          uuid.New().String(),
		Title:       req.Title,
		Color:       color,
		Description: req.Description,
		CreatedAt:   time.Now(),
	}

	if err := h.repo.Create(c.Request().Context(), label); err != nil {
		return api.InternalError(c, err.Error())
	}

	return api.Created(c, label)
}

// Delete remove um label
func (h *LabelHandler) Delete(c echo.Context) error {
	id := c.Param("id")
	if err := h.repo.Delete(c.Request().Context(), id); err != nil {
		return api.InternalError(c, err.Error())
	}
	return api.NoContent(c)
}
