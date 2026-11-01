package goodsreceipt

import (
	"time"

	purchaseorder "proceive/internal/modules/purchase_order"
	"proceive/internal/modules/user"

	"github.com/google/uuid"
)

// Status Goods Receipt
const (
	StatusWaitingSJ   = "WAITING_SJ"
	StatusSJUploaded  = "SJ_UPLOADED"
	StatusSJRevised   = "SJ_REVISED"
	StatusSJApproved  = "SJ_APPROVED"
	StatusGRCompleted = "GR_COMPLETED"
)

// GoodsReceipt merepresentasikan proses logistik penerimaan barang dari vendor
type GoodsReceipt struct {
	ID                uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	POID              uuid.UUID  `gorm:"type:uuid;not null;index" json:"po_id"`
	PurchasingID      *uuid.UUID `gorm:"type:uuid;index" json:"purchasing_id,omitempty"`
	PPICID            *uuid.UUID `gorm:"type:uuid;index" json:"ppic_id,omitempty"`
	SuratJalanNo      string     `gorm:"type:varchar(100)" json:"surat_jalan_no"`
	SuratJalanURL     string     `gorm:"type:varchar(255)" json:"surat_jalan_url"`
	GRDocumentURL     string     `gorm:"type:varchar(255)" json:"gr_document_url,omitempty"`
	VerificationNotes string     `gorm:"type:text" json:"verification_notes"`
	Notes             string     `gorm:"type:text" json:"notes,omitempty"`
	GRNumber          *string    `gorm:"type:varchar(50);uniqueIndex" json:"gr_number,omitempty"`
	Status            string     `gorm:"type:varchar(30);not null;default:'WAITING_SJ';index" json:"status"`
	CreatedAt         time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt         time.Time  `gorm:"autoUpdateTime" json:"updated_at"`

	PO         *purchaseorder.PurchaseOrder `gorm:"foreignKey:POID;constraint:OnDelete:RESTRICT" json:"po,omitempty"`
	Purchasing *user.User                   `gorm:"foreignKey:PurchasingID;constraint:OnDelete:SET NULL" json:"purchasing,omitempty"`
	PPIC       *user.User                   `gorm:"foreignKey:PPICID;constraint:OnDelete:SET NULL" json:"ppic,omitempty"`
}

func (GoodsReceipt) TableName() string {
	return "goods_receipts"
}
