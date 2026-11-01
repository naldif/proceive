package auth

import (
	"time"

	"github.com/google/uuid"
)

// LoginRequest parameter autentikasi email dan password.
type LoginRequest struct {
	Email    string `json:"email" form:"email" binding:"required,email"`
	Password string `json:"password" form:"password" binding:"required"`
}

// UserInfoResponse ringkasan profil user pada response login.
type UserInfoResponse struct {
	ID         uuid.UUID  `json:"id"`
	Name       string     `json:"name"`
	Email      string     `json:"email"`
	Role       string     `json:"role"`
	RoleID     uuid.UUID  `json:"role_id"`
	DivisionID *uuid.UUID `json:"division_id,omitempty"`
	Division   string     `json:"division,omitempty"`
	PositionID *uuid.UUID `json:"position_id,omitempty"`
	Position   string     `json:"position,omitempty"`
	VendorID   *uuid.UUID `json:"vendor_id,omitempty"`
}

// LoginResponse response login beserta token JWT.
type LoginResponse struct {
	Token string           `json:"token"`
	User  UserInfoResponse `json:"user"`
}

// ProfileResponse data profil lengkap user yang sedang aktif (/auth/me).
type ProfileResponse struct {
	ID         uuid.UUID  `json:"id"`
	Name       string     `json:"name"`
	Email      string     `json:"email"`
	RoleID     uuid.UUID  `json:"role_id"`
	Role       string     `json:"role"`
	RoleCode   string     `json:"role_code"`
	DivisionID *uuid.UUID `json:"division_id,omitempty"`
	Division   string     `json:"division,omitempty"`
	PositionID *uuid.UUID `json:"position_id,omitempty"`
	Position   string     `json:"position,omitempty"`
	VendorID   *uuid.UUID `json:"vendor_id,omitempty"`
	VendorName string     `json:"vendor_name,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}
