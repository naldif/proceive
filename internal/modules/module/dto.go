package module

import (
	"time"

	"github.com/google/uuid"
)

// CreateModuleRequest payload untuk membuat modul baru
type CreateModuleRequest struct {
	Name        string `json:"name" binding:"required,min=2,max=100"`
	Slug        string `json:"slug" binding:"required,min=2,max=50"`
	Description string `json:"description" binding:"omitempty,max=255"`
}

// UpdateModuleRequest payload untuk memperbarui modul
type UpdateModuleRequest struct {
	Name        string `json:"name" binding:"omitempty,min=2,max=100"`
	Slug        string `json:"slug" binding:"omitempty,min=2,max=50"`
	Description string `json:"description" binding:"omitempty,max=255"`
}

// ModuleResponse representasi output data modul
type ModuleResponse struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func ToModuleResponse(m *Module) ModuleResponse {
	return ModuleResponse{
		ID:          m.ID,
		Name:        m.Name,
		Slug:        m.Slug,
		Description: m.Description,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}

func ToModuleResponseList(modules []Module) []ModuleResponse {
	res := make([]ModuleResponse, len(modules))
	for i, m := range modules {
		res[i] = ToModuleResponse(&m)
	}
	return res
}
