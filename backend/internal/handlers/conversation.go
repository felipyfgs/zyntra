package handlers

import (
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/zyntra/backend/internal/api"
	"github.com/zyntra/backend/internal/domain"
	"github.com/zyntra/backend/internal/services"
)

// ConversationHandler handler de conversas
type ConversationHandler struct {
	service *services.ConversationService
}

// NewConversationHandler cria novo handler
func NewConversationHandler(service *services.ConversationService) *ConversationHandler {
	return &ConversationHandler{service: service}
}

// List lista conversas com filtros
func (h *ConversationHandler) List(c echo.Context) error {
	filter := domain.ConversationFilter{}

	if inboxID := c.QueryParam("inbox_id"); inboxID != "" {
		filter.InboxID = &inboxID
	}
	if status := c.QueryParam("status"); status != "" {
		s := domain.ConversationStatus(status)
		filter.Status = &s
	}
	if assigneeID := c.QueryParam("assignee_id"); assigneeID != "" {
		filter.AssigneeID = &assigneeID
	}
	if favorite := c.QueryParam("favorite"); favorite == "true" {
		f := true
		filter.IsFavorite = &f
	}
	if archived := c.QueryParam("archived"); archived == "true" {
		a := true
		filter.IsArchived = &a
	}
	if limit, _ := strconv.Atoi(c.QueryParam("limit")); limit > 0 {
		filter.Limit = limit
	}
	if offset, _ := strconv.Atoi(c.QueryParam("offset")); offset > 0 {
		filter.Offset = offset
	}

	conversations, err := h.service.ListWithDetails(c.Request().Context(), filter)
	if err != nil {
		return api.InternalError(c, err.Error())
	}

	return api.Success(c, conversations)
}

// Get retorna uma conversa por ID
func (h *ConversationHandler) Get(c echo.Context) error {
	id := c.Param("id")
	conv, err := h.service.GetWithDetails(c.Request().Context(), id)
	if err != nil {
		return api.NotFound(c, err.Error())
	}
	return api.Success(c, conv)
}

// UpdateConversationRequest request para atualizar conversa
type UpdateConversationRequest struct {
	Status     *string `json:"status,omitempty"`
	Priority   *string `json:"priority,omitempty"`
	AssigneeID *string `json:"assignee_id,omitempty"`
}

// Update atualiza uma conversa
func (h *ConversationHandler) Update(c echo.Context) error {
	id := c.Param("id")
	var req UpdateConversationRequest
	if err := c.Bind(&req); err != nil {
		return api.BadRequest(c, "Invalid request body")
	}

	updateReq := domain.UpdateConversationRequest{}
	if req.Status != nil {
		s := domain.ConversationStatus(*req.Status)
		updateReq.Status = &s
	}
	if req.Priority != nil {
		p := domain.ConversationPriority(*req.Priority)
		updateReq.Priority = &p
	}
	if req.AssigneeID != nil {
		updateReq.AssigneeID = req.AssigneeID
	}

	conv, err := h.service.Update(c.Request().Context(), id, updateReq)
	if err != nil {
		return api.InternalError(c, err.Error())
	}

	return api.Success(c, conv)
}

// Delete remove uma conversa
func (h *ConversationHandler) Delete(c echo.Context) error {
	id := c.Param("id")
	if err := h.service.Delete(c.Request().Context(), id); err != nil {
		return api.InternalError(c, err.Error())
	}
	return api.NoContent(c)
}

// MarkAsRead marca conversa como lida
func (h *ConversationHandler) MarkAsRead(c echo.Context) error {
	id := c.Param("id")
	if err := h.service.MarkAsRead(c.Request().Context(), id); err != nil {
		return api.InternalError(c, err.Error())
	}
	return api.Success(c, map[string]string{"message": "Marked as read"})
}

// AssignRequest request para atribuir conversa
type AssignRequest struct {
	AssigneeID string `json:"assignee_id"`
}

// Assign atribui conversa a um agente
func (h *ConversationHandler) Assign(c echo.Context) error {
	id := c.Param("id")
	var req AssignRequest
	if err := c.Bind(&req); err != nil {
		return api.BadRequest(c, "Invalid request body")
	}

	if req.AssigneeID == "" {
		if err := h.service.Unassign(c.Request().Context(), id); err != nil {
			return api.InternalError(c, err.Error())
		}
	} else {
		if err := h.service.Assign(c.Request().Context(), id, req.AssigneeID); err != nil {
			return api.InternalError(c, err.Error())
		}
	}

	return api.Success(c, map[string]string{"message": "Assignment updated"})
}

// ToggleFavorite alterna favorito
func (h *ConversationHandler) ToggleFavorite(c echo.Context) error {
	id := c.Param("id")
	conv, err := h.service.GetByID(c.Request().Context(), id)
	if err != nil {
		return api.NotFound(c, err.Error())
	}

	if err := h.service.SetFavorite(c.Request().Context(), id, !conv.IsFavorite); err != nil {
		return api.InternalError(c, err.Error())
	}

	return api.Success(c, map[string]bool{"is_favorite": !conv.IsFavorite})
}

// ToggleArchive alterna arquivado
func (h *ConversationHandler) ToggleArchive(c echo.Context) error {
	id := c.Param("id")
	conv, err := h.service.GetByID(c.Request().Context(), id)
	if err != nil {
		return api.NotFound(c, err.Error())
	}

	if err := h.service.SetArchived(c.Request().Context(), id, !conv.IsArchived); err != nil {
		return api.InternalError(c, err.Error())
	}

	return api.Success(c, map[string]bool{"is_archived": !conv.IsArchived})
}
