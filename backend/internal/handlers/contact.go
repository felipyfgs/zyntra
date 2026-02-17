package handlers

import (
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/zyntra/backend/internal/api"
	"github.com/zyntra/backend/internal/domain"
	"github.com/zyntra/backend/internal/services"
)

// ContactHandler handler de contatos
type ContactHandler struct {
	service     *services.ContactService
	convService *services.ConversationService
}

// NewContactHandler cria novo handler
func NewContactHandler(service *services.ContactService, convService *services.ConversationService) *ContactHandler {
	return &ContactHandler{
		service:     service,
		convService: convService,
	}
}

// List lista contatos
func (h *ContactHandler) List(c echo.Context) error {
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	offset, _ := strconv.Atoi(c.QueryParam("offset"))
	search := c.QueryParam("search")

	if limit <= 0 {
		limit = 50
	}

	var contacts []*domain.Contact
	var err error

	if search != "" {
		contacts, err = h.service.Search(c.Request().Context(), search, limit)
	} else {
		contacts, err = h.service.List(c.Request().Context(), limit, offset)
	}

	if err != nil {
		return api.InternalError(c, err.Error())
	}

	return api.Success(c, contacts)
}

// Get retorna um contato por ID
func (h *ContactHandler) Get(c echo.Context) error {
	id := c.Param("id")
	contact, err := h.service.GetWithInboxes(c.Request().Context(), id)
	if err != nil {
		return api.NotFound(c, err.Error())
	}
	return api.Success(c, contact)
}

// CreateContactRequest request para criar contato
type CreateContactRequest struct {
	Name             string                 `json:"name"`
	Email            string                 `json:"email,omitempty"`
	PhoneNumber      string                 `json:"phone_number,omitempty"`
	CustomAttributes map[string]interface{} `json:"custom_attributes,omitempty"`
}

// Create cria um novo contato
func (h *ContactHandler) Create(c echo.Context) error {
	var req CreateContactRequest
	if err := c.Bind(&req); err != nil {
		return api.BadRequest(c, "Invalid request body")
	}

	contact, err := h.service.Create(c.Request().Context(), domain.CreateContactRequest{
		Name:             req.Name,
		Email:            req.Email,
		PhoneNumber:      req.PhoneNumber,
		CustomAttributes: req.CustomAttributes,
	})
	if err != nil {
		return api.InternalError(c, err.Error())
	}

	return api.Created(c, contact)
}

// UpdateContactRequest request para atualizar contato
type UpdateContactRequest struct {
	Name             *string                `json:"name,omitempty"`
	Email            *string                `json:"email,omitempty"`
	PhoneNumber      *string                `json:"phone_number,omitempty"`
	AvatarURL        *string                `json:"avatar_url,omitempty"`
	CustomAttributes map[string]interface{} `json:"custom_attributes,omitempty"`
}

// Update atualiza um contato
func (h *ContactHandler) Update(c echo.Context) error {
	id := c.Param("id")
	var req UpdateContactRequest
	if err := c.Bind(&req); err != nil {
		return api.BadRequest(c, "Invalid request body")
	}

	contact, err := h.service.Update(c.Request().Context(), id, domain.UpdateContactRequest{
		Name:             req.Name,
		Email:            req.Email,
		PhoneNumber:      req.PhoneNumber,
		AvatarURL:        req.AvatarURL,
		CustomAttributes: req.CustomAttributes,
	})
	if err != nil {
		return api.InternalError(c, err.Error())
	}

	return api.Success(c, contact)
}

// Delete remove um contato
func (h *ContactHandler) Delete(c echo.Context) error {
	id := c.Param("id")
	if err := h.service.Delete(c.Request().Context(), id); err != nil {
		return api.InternalError(c, err.Error())
	}
	return api.NoContent(c)
}

// GetConversations lista conversas de um contato
func (h *ContactHandler) GetConversations(c echo.Context) error {
	id := c.Param("id")
	
	filter := domain.ConversationFilter{
		ContactID: &id,
	}

	conversations, err := h.convService.List(c.Request().Context(), filter)
	if err != nil {
		return api.InternalError(c, err.Error())
	}

	return api.Success(c, conversations)
}
