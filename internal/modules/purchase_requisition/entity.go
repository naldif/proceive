package purchaserequisition

import (
	"time"

	"proceive/internal/modules/item"
	"proceive/internal/modules/user"

	"github.com/google/uuid"
)

// Status Purchase Requisition
const (
	StatusDraft           = "DRAFT"
	StatusPendingApproval = "PENDING_APPROVAL"
	StatusApproved        = "APPROVED"
	StatusRejected        = "REJECTED"
)

// PurchaseRequisition merepresentasikan header pengajuan permintaan barang
type PurchaseRequisition struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	RequesterID uuid.UUID  `gorm:"type:uuid;not null;index" json:"requester_id"`
	ApproverID  *uuid.UUID `gorm:"type:uuid;index" json:"approver_id,omitempty"`
	PRNumber    string     `gorm:"type:varchar(50);uniqueIndex;not null" json:"pr_number"`
	Status      string     `gorm:"type:varchar(30);not null;default:'DRAFT';index" json:"status"`
	Notes         string     `gorm:"type:text" json:"notes"`
	ApprovalNotes string     `gorm:"type:text" json:"approval_notes,omitempty"`
	CreatedAt     time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time  `gorm:"autoUpdateTime" json:"updated_at"`

	Requester *user.User `gorm:"foreignKey:RequesterID;constraint:OnDelete:RESTRICT" json:"requester,omitempty"`
	Approver  *user.User `gorm:"foreignKey:ApproverID;constraint:OnDelete:SET NULL" json:"approver,omitempty"`
	Lines     []PRLine   `gorm:"foreignKey:PRID;constraint:OnDelete:CASCADE" json:"lines,omitempty"`
}

func (PurchaseRequisition) TableName() string {
	return "purchase_requisitions"
}

// PRLine merepresentasikan detail item barang pada Purchase Requisition
type PRLine struct {
	ID             uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	PRID           uuid.UUID  `gorm:"type:uuid;not null;index" json:"pr_id"`
	ItemID         uuid.UUID  `gorm:"type:uuid;not null;index" json:"item_id"`
	Qty            int        `gorm:"not null" json:"qty"`
	EstimatedPrice float64    `gorm:"type:decimal(15,2);default:0" json:"estimated_price"`
	CreatedAt      time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt      time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
	Item           *item.Item `gorm:"foreignKey:ItemID;constraint:OnDelete:RESTRICT" json:"item,omitempty"`
}

func (PRLine) TableName() string {
	return "pr_lines"
}
