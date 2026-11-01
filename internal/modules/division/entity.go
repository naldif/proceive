package division

import (
	"time"

	"github.com/google/uuid"
)

// Division merepresentasikan master divisi perusahaan (IT, HRD, Produksi, dll)
type Division struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Code      string    `gorm:"type:varchar(50);uniqueIndex;not null" json:"code"`
	Name      string    `gorm:"type:varchar(100);not null" json:"name"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (Division) TableName() string {
	return "divisions"
}
