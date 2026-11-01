package division

import (
	"time"

	"github.com/google/uuid"
)

// CreateDivisionRequest payload untuk membuat data master divisi baru
type CreateDivisionRequest struct {
	Code string `json:"code" binding:"required,min=2,max=50"`
	Name string `json:"name" binding:"required,min=2,max=100"`
}

// UpdateDivisionRequest payload untuk memperbarui data divisi
type UpdateDivisionRequest struct {
	Code string `json:"code" binding:"omitempty,min=2,max=50"`
	Name string `json:"name" binding:"omitempty,min=2,max=100"`
}

// DivisionResponse representasi output data master divisi
type DivisionResponse struct {
	ID        uuid.UUID `json:"id"`
	Code      string    `json:"code"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func ToDivisionResponse(d *Division) DivisionResponse {
	return DivisionResponse{
		ID:        d.ID,
		Code:      d.Code,
		Name:      d.Name,
		CreatedAt: d.CreatedAt,
		UpdatedAt: d.UpdatedAt,
	}
}

func ToDivisionResponseList(divisions []Division) []DivisionResponse {
	res := make([]DivisionResponse, len(divisions))
	for i, d := range divisions {
		res[i] = ToDivisionResponse(&d)
	}
	return res
}
