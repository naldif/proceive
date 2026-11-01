package item

import (
	"time"

	"github.com/google/uuid"
)

// Item merepresentasikan katalog master barang yang dapat dipesan
type Item struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Code          string    `gorm:"type:varchar(50);uniqueIndex;not null" json:"code"`
	Name          string    `gorm:"type:varchar(150);not null" json:"name"`
	Category      string    `gorm:"type:varchar(100);not null;index" json:"category"`
	UOM           string    `gorm:"type:varchar(20);not null" json:"uom"` // Unit of Measure: PCS, BOX, UNIT, KG, RIM
	PriceEstimate float64   `gorm:"type:decimal(15,2);default:0" json:"price_estimate"`
	CreatedAt     time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (Item) TableName() string {
	return "items"
}
