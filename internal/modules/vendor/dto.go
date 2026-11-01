package vendor

import (
	"time"

	"github.com/google/uuid"
)

// CreateVendorRequest parameter pembuatan master vendor baru.
type CreateVendorRequest struct {
	Name          string `json:"name" binding:"required,min=2,max=150"`
	Category      string `json:"category" binding:"required,min=2,max=100"`
	ContactPerson string `json:"contact_person" binding:"omitempty,max=100"`
	Email         string `json:"email" binding:"required,email,max=150"`
	Password      string `json:"password" binding:"omitempty,min=6,max=50"`
	Address       string `json:"address" binding:"omitempty,max=500"`
	IsActive      *bool  `json:"is_active" binding:"omitempty"`
}

// UpdateVendorRequest parameter pembaruan data vendor.
type UpdateVendorRequest struct {
	Name          string `json:"name" binding:"omitempty,min=2,max=150"`
	Category      string `json:"category" binding:"omitempty,min=2,max=100"`
	ContactPerson string `json:"contact_person" binding:"omitempty,max=100"`
	Email         string `json:"email" binding:"omitempty,email,max=150"`
	Address       string `json:"address" binding:"omitempty,max=500"`
	IsActive      *bool  `json:"is_active" binding:"omitempty"`
}

// VendorResponse rincian data rekanan vendor.
type VendorResponse struct {
	ID            uuid.UUID `json:"id"`
	Name          string    `json:"name"`
	Category      string    `json:"category"`
	ContactPerson string    `json:"contact_person"`
	Email         string    `json:"email"`
	Address       string    `json:"address"`
	IsActive      bool      `json:"is_active"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func ToVendorResponse(v *Vendor) VendorResponse {
	return VendorResponse{
		ID:            v.ID,
		Name:          v.Name,
		Category:      v.Category,
		ContactPerson: v.ContactPerson,
		Email:         v.Email,
		Address:       v.Address,
		IsActive:      v.IsActive,
		CreatedAt:     v.CreatedAt,
		UpdatedAt:     v.UpdatedAt,
	}
}

func ToVendorResponseList(vendors []Vendor) []VendorResponse {
	res := make([]VendorResponse, len(vendors))
	for i, v := range vendors {
		res[i] = ToVendorResponse(&v)
	}
	return res
}
