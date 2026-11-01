package permission

import (
	"time"

	"github.com/google/uuid"
)

// CreatePermissionRequest payload untuk membuat permission baru
type CreatePermissionRequest struct {
	ModuleID    string `json:"module_id" binding:"required,uuid"`
	Name        string `json:"name" binding:"required,min=2,max=100"`
	Code        string `json:"code" binding:"required,min=2,max=100"`
	Description string `json:"description" binding:"omitempty,max=255"`
}

// UpdatePermissionRequest payload untuk memperbarui permission
type UpdatePermissionRequest struct {
	ModuleID    string `json:"module_id" binding:"omitempty,uuid"`
	Name        string `json:"name" binding:"omitempty,min=2,max=100"`
	Code        string `json:"code" binding:"omitempty,min=2,max=100"`
	Description string `json:"description" binding:"omitempty,max=255"`
}

// PermissionResponse representasi output data permission
type PermissionResponse struct {
	ID          uuid.UUID `json:"id"`
	ModuleID    uuid.UUID `json:"module_id"`
	ModuleName  string    `json:"module_name,omitempty"`
	ModuleSlug  string    `json:"module_slug,omitempty"`
	Name        string    `json:"name"`
	Code        string    `json:"code"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func ToPermissionResponse(p *Permission) PermissionResponse {
	modName := ""
	modSlug := ""
	if p.Module != nil {
		modName = p.Module.Name
		modSlug = p.Module.Slug
	}
	return PermissionResponse{
		ID:          p.ID,
		ModuleID:    p.ModuleID,
		ModuleName:  modName,
		ModuleSlug:  modSlug,
		Name:        p.Name,
		Code:        p.Code,
		Description: p.Description,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
}

func ToPermissionResponseList(permissions []Permission) []PermissionResponse {
	res := make([]PermissionResponse, len(permissions))
	for i, p := range permissions {
		res[i] = ToPermissionResponse(&p)
	}
	return res
}
