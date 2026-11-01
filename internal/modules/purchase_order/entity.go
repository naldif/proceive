package purchaseorder

import (
	"time"

	"proceive/internal/modules/item"
	purchaserequisition "proceive/internal/modules/purchase_requisition"
	"proceive/internal/modules/user"
	"proceive/internal/modules/vendor"

	"github.com/google/uuid"
)

// Status Purchase Order
const (
	StatusDraft                 = "DRAFT"
	StatusSentToVendor          = "SENT_TO_VENDOR"
	StatusSJPendingVerification = "SJ_PENDING_VERIFICATION"
	StatusSJRevised             = "SJ_REVISED"
	StatusReadyForGR            = "READY_FOR_GR"
	StatusGRCompleted           = "GR_COMPLETED"
)

// PurchaseOrder merepresentasikan header pesanan pembelian ke vendor
type PurchaseOrder struct {
	ID           uuid.UUID                                `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	PRID         uuid.UUID                                `gorm:"type:uuid;not null;index" json:"pr_id"`
	VendorID     uuid.UUID                                `gorm:"type:uuid;not null;index" json:"vendor_id"`
	PurchasingID uuid.UUID                                `gorm:"type:uuid;not null;index" json:"purchasing_id"`
	PONumber     string                                   `gorm:"type:varchar(50);uniqueIndex;not null" json:"po_number"`
	Status       string                                   `gorm:"type:varchar(30);not null;default:'DRAFT';index" json:"status"`
	TotalAmount  float64                                  `gorm:"type:decimal(15,2);default:0" json:"total_amount"`
	PDFURL       string                                   `gorm:"type:varchar(255)" json:"pdf_url"`
	CreatedAt    time.Time                                `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time                                `gorm:"autoUpdateTime" json:"updated_at"`

	PR         *purchaserequisition.PurchaseRequisition `gorm:"foreignKey:PRID;constraint:OnDelete:RESTRICT" json:"pr,omitempty"`
	Vendor     *vendor.Vendor                           `gorm:"foreignKey:VendorID;constraint:OnDelete:RESTRICT" json:"vendor,omitempty"`
	Purchasing *user.User                               `gorm:"foreignKey:PurchasingID;constraint:OnDelete:RESTRICT" json:"purchasing,omitempty"`
	Lines      []POLine                                 `gorm:"foreignKey:POID;constraint:OnDelete:CASCADE" json:"lines,omitempty"`
}

func (PurchaseOrder) TableName() string {
	return "purchase_orders"
}

// POLine merepresentasikan item barang pada Purchase Order
type POLine struct {
	ID        uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	POID      uuid.UUID  `gorm:"type:uuid;not null;index" json:"po_id"`
	ItemID    uuid.UUID  `gorm:"type:uuid;not null;index" json:"item_id"`
	Qty       int        `gorm:"not null" json:"qty"`
	UnitPrice float64    `gorm:"type:decimal(15,2);default:0" json:"unit_price"`
	CreatedAt time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
	Item      *item.Item `gorm:"foreignKey:ItemID;constraint:OnDelete:RESTRICT" json:"item,omitempty"`
}

func (POLine) TableName() string {
	return "po_lines"
}
