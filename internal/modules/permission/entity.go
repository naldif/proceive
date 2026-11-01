package permission

import (
	"time"

	"proceive/internal/modules/module"

	"github.com/google/uuid"
)

// Permission merepresentasikan hak aksi spesifik pada suatu modul
type Permission struct {
	ID          uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ModuleID    uuid.UUID      `gorm:"type:uuid;not null;index" json:"module_id"`
	Name        string         `gorm:"type:varchar(100);not null" json:"name"`
	Code        string         `gorm:"type:varchar(100);uniqueIndex;not null" json:"code"`
	Description string         `gorm:"type:text" json:"description"`
	CreatedAt   time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	Module      *module.Module `gorm:"foreignKey:ModuleID;constraint:OnDelete:CASCADE" json:"module,omitempty"`
}

func (Permission) TableName() string {
	return "permissions"
}
