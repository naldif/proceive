package position

import (
	"time"

	"github.com/google/uuid"
)

// CreatePositionRequest payload untuk membuat master jabatan baru
type CreatePositionRequest struct {
	Code string `json:"code" binding:"required,min=2,max=50"`
	Name string `json:"name" binding:"required,min=2,max=100"`
}

// UpdatePositionRequest payload untuk memperbarui data jabatan
type UpdatePositionRequest struct {
	Code string `json:"code" binding:"omitempty,min=2,max=50"`
	Name string `json:"name" binding:"omitempty,min=2,max=100"`
}

// PositionResponse representasi output data master jabatan
type PositionResponse struct {
	ID        uuid.UUID `json:"id"`
	Code      string    `json:"code"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func ToPositionResponse(p *Position) PositionResponse {
	return PositionResponse{
		ID:        p.ID,
		Code:      p.Code,
		Name:      p.Name,
		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
	}
}

func ToPositionResponseList(positions []Position) []PositionResponse {
	res := make([]PositionResponse, len(positions))
	for i, p := range positions {
		res[i] = ToPositionResponse(&p)
	}
	return res
}
