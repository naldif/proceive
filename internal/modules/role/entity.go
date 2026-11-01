package role

import (
	"time"

	"proceive/internal/modules/permission"

	"github.com/google/uuid"
)

// Role merepresentasikan peran pengguna dalam sistem RBAC
type Role struct {
	ID          uuid.UUID               `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Code        string                  `gorm:"type:varchar(50);uniqueIndex;not null" json:"code"`
	Name        string                  `gorm:"type:varchar(100);not null" json:"name"`
	Description string                  `gorm:"type:text" json:"description"`
	CreatedAt   time.Time               `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time               `gorm:"autoUpdateTime" json:"updated_at"`
	Permissions []permission.Permission `gorm:"many2many:role_has_permissions;constraint:OnDelete:CASCADE" json:"permissions,omitempty"`
}

func (Role) TableName() string {
	return "roles"
}

// RoleHasPermission merepresentasikan tabel pivot antara roles dan permissions
type RoleHasPermission struct {
	RoleID       uuid.UUID `gorm:"type:uuid;primaryKey;index" json:"role_id"`
	PermissionID uuid.UUID `gorm:"type:uuid;primaryKey;index" json:"permission_id"`
}

func (RoleHasPermission) TableName() string {
	return "role_has_permissions"
}
