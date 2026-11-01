package role

import (
	"time"

	"proceive/internal/modules/permission"

	"github.com/google/uuid"
)

// CreateRoleRequest parameter pembuatan role baru.
type CreateRoleRequest struct {
	Code          string   `json:"code" binding:"required,min=2,max=50"`
	Name          string   `json:"name" binding:"required,min=2,max=100"`
	Description   string   `json:"description" binding:"omitempty,max=255"`
	PermissionIDs []string `json:"permission_ids" binding:"omitempty,dive,uuid"`
}

// UpdateRoleRequest parameter pembaruan data role.
type UpdateRoleRequest struct {
	Code        string `json:"code" binding:"omitempty,min=2,max=50"`
	Name        string `json:"name" binding:"omitempty,min=2,max=100"`
	Description string `json:"description" binding:"omitempty,max=255"`
}

// AssignPermissionsRequest parameter penugasan daftar permission ke role.
type AssignPermissionsRequest struct {
	PermissionIDs []string `json:"permission_ids" binding:"required,dive,uuid"`
}

// PermissionItemResponse rincian permission pada response role.
type PermissionItemResponse struct {
	ID          uuid.UUID `json:"id"`
	ModuleID    uuid.UUID `json:"module_id"`
	Name        string    `json:"name"`
	Code        string    `json:"code"`
	Description string    `json:"description"`
}

// RoleResponse data ringkas role.
type RoleResponse struct {
	ID          uuid.UUID `json:"id"`
	Code        string    `json:"code"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// RoleDetailResponse data detail role beserta daftar permissions.
type RoleDetailResponse struct {
	ID          uuid.UUID                `json:"id"`
	Code        string                   `json:"code"`
	Name        string                   `json:"name"`
	Description string                   `json:"description"`
	Permissions []PermissionItemResponse `json:"permissions"`
	CreatedAt   time.Time                `json:"created_at"`
	UpdatedAt   time.Time                `json:"updated_at"`
}

func ToPermissionItemResponse(p *permission.Permission) PermissionItemResponse {
	return PermissionItemResponse{
		ID:          p.ID,
		ModuleID:    p.ModuleID,
		Name:        p.Name,
		Code:        p.Code,
		Description: p.Description,
	}
}

func ToPermissionItemResponseList(perms []permission.Permission) []PermissionItemResponse {
	res := make([]PermissionItemResponse, len(perms))
	for i, p := range perms {
		res[i] = ToPermissionItemResponse(&p)
	}
	return res
}

func ToRoleResponse(r *Role) RoleResponse {
	return RoleResponse{
		ID:          r.ID,
		Code:        r.Code,
		Name:        r.Name,
		Description: r.Description,
		CreatedAt:   r.CreatedAt,
		UpdatedAt:   r.UpdatedAt,
	}
}

func ToRoleResponseList(roles []Role) []RoleResponse {
	res := make([]RoleResponse, len(roles))
	for i, r := range roles {
		res[i] = ToRoleResponse(&r)
	}
	return res
}

func ToRoleDetailResponse(r *Role) RoleDetailResponse {
	return RoleDetailResponse{
		ID:          r.ID,
		Code:        r.Code,
		Name:        r.Name,
		Description: r.Description,
		Permissions: ToPermissionItemResponseList(r.Permissions),
		CreatedAt:   r.CreatedAt,
		UpdatedAt:   r.UpdatedAt,
	}
}
