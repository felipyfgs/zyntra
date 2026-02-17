package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func GetContacts(c echo.Context) error {
	// TODO: Implement actual contact listing
	contacts := []map[string]interface{}{
		{
			"id":         "1",
			"name":       "John Doe",
			"phone":      "+5511999999999",
			"avatar_url": "",
			"created_at": "2024-02-16T10:00:00Z",
		},
	}

	return c.JSON(http.StatusOK, contacts)
}

type CreateContactRequest struct {
	ConnectionID string `json:"connection_id" validate:"required"`
	Phone        string `json:"phone" validate:"required"`
	Name         string `json:"name" validate:"required"`
}

func CreateContact(c echo.Context) error {
	var req CreateContactRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
	}

	// TODO: Implement actual contact creation
	contact := map[string]interface{}{
		"id":            "2",
		"connection_id": req.ConnectionID,
		"name":          req.Name,
		"phone":         req.Phone,
		"created_at":    "2024-02-16T10:00:00Z",
	}

	return c.JSON(http.StatusCreated, contact)
}

type UpdateContactRequest struct {
	Name string `json:"name" validate:"required"`
}

func UpdateContact(c echo.Context) error {
	id := c.Param("id")

	var req UpdateContactRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
	}

	// TODO: Implement actual contact update
	contact := map[string]interface{}{
		"id":         id,
		"name":       req.Name,
		"updated_at": "2024-02-16T11:00:00Z",
	}

	return c.JSON(http.StatusOK, contact)
}

func DeleteContact(c echo.Context) error {
	// id := c.Param("id")

	// TODO: Implement actual contact deletion

	return c.NoContent(http.StatusNoContent)
}
