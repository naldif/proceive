package user

import (
	"time"

	"proceive/internal/modules/division"
	"proceive/internal/modules/position"
	"proceive/internal/modules/role"
	"proceive/internal/modules/vendor"

	"github.com/google/uuid"
)

// User merepresentasikan data pengguna sistem (Pegawai internal & Akun Vendor)
type User struct {
	ID           uuid.UUID          `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	DivisionID   *uuid.UUID         `gorm:"type:uuid;index" json:"division_id,omitempty"`
	PositionID   *uuid.UUID         `gorm:"type:uuid;index" json:"position_id,omitempty"`
	RoleID       uuid.UUID          `gorm:"type:uuid;not null;index" json:"role_id"`
	VendorID     *uuid.UUID         `gorm:"type:uuid;index" json:"vendor_id,omitempty"`
	Name         string             `gorm:"type:varchar(150);not null" json:"name"`
	Email        string             `gorm:"type:varchar(150);uniqueIndex;not null" json:"email"`
	PasswordHash string             `gorm:"type:varchar(255);not null" json:"-"`
	CreatedAt    time.Time          `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time          `gorm:"autoUpdateTime" json:"updated_at"`
	Division     *division.Division `gorm:"foreignKey:DivisionID;constraint:OnDelete:SET NULL" json:"division,omitempty"`
	Position     *position.Position `gorm:"foreignKey:PositionID;constraint:OnDelete:SET NULL" json:"position,omitempty"`
	Role         *role.Role         `gorm:"foreignKey:RoleID;constraint:OnDelete:RESTRICT" json:"role,omitempty"`
	Vendor       *vendor.Vendor     `gorm:"foreignKey:VendorID;constraint:OnDelete:SET NULL" json:"vendor,omitempty"`
}

func (User) TableName() string {
	return "users"
}
