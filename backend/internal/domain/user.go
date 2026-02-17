package domain

import (
	"time"
)

// UserRole role do usuario
type UserRole string

const (
	UserRoleAdmin UserRole = "admin"
	UserRoleAgent UserRole = "agent"
)

// User usuario/agente
type User struct {
	ID           string    `json:"id" db:"id"`
	Name         string    `json:"name" db:"name"`
	Email        string    `json:"email" db:"email"`
	PasswordHash string    `json:"-" db:"password_hash"`
	Role         UserRole  `json:"role" db:"role"`
	AvatarURL    string    `json:"avatar_url,omitempty" db:"avatar_url"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// UserPublic dados publicos do usuario
type UserPublic struct {
	ID        string   `json:"id"`
	Name      string   `json:"name"`
	Email     string   `json:"email"`
	Role      UserRole `json:"role"`
	AvatarURL string   `json:"avatar_url,omitempty"`
}

// ToPublic converte para dados publicos
func (u *User) ToPublic() UserPublic {
	return UserPublic{
		ID:        u.ID,
		Name:      u.Name,
		Email:     u.Email,
		Role:      u.Role,
		AvatarURL: u.AvatarURL,
	}
}
