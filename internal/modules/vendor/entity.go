package vendor

import (
	"time"

	"github.com/google/uuid"
)

// Vendor merepresentasikan data supplier atau penyedia barang
type Vendor struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Name          string    `gorm:"type:varchar(150);not null" json:"name"`
	Category      string    `gorm:"type:varchar(100);index" json:"category"` // Kategori / jenis vendor (Hardware Komputer, Logistik, ATK, dll)
	ContactPerson string    `gorm:"type:varchar(100)" json:"contact_person"`
	Email         string    `gorm:"type:varchar(150);uniqueIndex;not null" json:"email"`
	Address       string    `gorm:"type:text" json:"address"`
	IsActive      bool      `gorm:"default:true" json:"is_active"`
	CreatedAt     time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (Vendor) TableName() string {
	return "vendors"
}
