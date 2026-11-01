package user

import (
	"time"

	"github.com/google/uuid"
)

// CreateUserRequest parameter pembuatan user baru.
type CreateUserRequest struct {
	Name       string  `json:"name" binding:"required,min=2,max=150"`
	Email      string  `json:"email" binding:"required,email,max=150"`
	Password   string  `json:"password" binding:"required,min=6,max=50"`
	RoleID     string  `json:"role_id" binding:"required,uuid"`
	DivisionID *string `json:"division_id" binding:"omitempty,uuid"`
	PositionID *string `json:"position_id" binding:"omitempty,uuid"`
	VendorID   *string `json:"vendor_id" binding:"omitempty,uuid"`
}

// UpdateUserRequest parameter pembaruan data user.
type UpdateUserRequest struct {
	Name       string  `json:"name" binding:"omitempty,min=2,max=150"`
	Email      string  `json:"email" binding:"omitempty,email,max=150"`
	Password   string  `json:"password" binding:"omitempty,min=6,max=50"`
	RoleID     string  `json:"role_id" binding:"omitempty,uuid"`
	DivisionID *string `json:"division_id" binding:"omitempty,uuid"`
	PositionID *string `json:"position_id" binding:"omitempty,uuid"`
	VendorID   *string `json:"vendor_id" binding:"omitempty,uuid"`
}

// UserResponse rincian data pengguna.
type UserResponse struct {
	ID         uuid.UUID  `json:"id"`
	Name       string     `json:"name"`
	Email      string     `json:"email"`
	RoleID     uuid.UUID  `json:"role_id"`
	Role       string     `json:"role,omitempty"`
	RoleCode   string     `json:"role_code,omitempty"`
	DivisionID *uuid.UUID `json:"division_id,omitempty"`
	Division   string     `json:"division,omitempty"`
	PositionID *uuid.UUID `json:"position_id,omitempty"`
	Position   string     `json:"position,omitempty"`
	VendorID   *uuid.UUID `json:"vendor_id,omitempty"`
	VendorName string     `json:"vendor_name,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

func ToUserResponse(u *User) UserResponse {
	roleName := ""
	roleCode := ""
	if u.Role != nil {
		roleName = u.Role.Name
		roleCode = u.Role.Code
	}

	divName := ""
	if u.Division != nil {
		divName = u.Division.Name
	}

	posName := ""
	if u.Position != nil {
		posName = u.Position.Name
	}

	vendorName := ""
	if u.Vendor != nil {
		vendorName = u.Vendor.Name
	}

	return UserResponse{
		ID:         u.ID,
		Name:       u.Name,
		Email:      u.Email,
		RoleID:     u.RoleID,
		Role:       roleName,
		RoleCode:   roleCode,
		DivisionID: u.DivisionID,
		Division:   divName,
		PositionID: u.PositionID,
		Position:   posName,
		VendorID:   u.VendorID,
		VendorName: vendorName,
		CreatedAt:  u.CreatedAt,
		UpdatedAt:  u.UpdatedAt,
	}
}

func ToUserResponseList(users []User) []UserResponse {
	res := make([]UserResponse, len(users))
	for i, u := range users {
		res[i] = ToUserResponse(&u)
	}
	return res
}
