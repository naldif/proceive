package purchaseorder

import (
	"time"

	"github.com/google/uuid"
)

// CreatePOLineRequest item pesanan dalam PO.
type CreatePOLineRequest struct {
	ItemID    string  `json:"item_id" binding:"required,uuid"`
	Qty       int     `json:"qty" binding:"required,min=1"`
	UnitPrice float64 `json:"unit_price" binding:"required,gte=0"`
}

// CreatePOFromPRRequest parameter pembuatan PO dari PR yang disetujui.
type CreatePOFromPRRequest struct {
	VendorID string `json:"vendor_id" binding:"required,uuid"`
}

// UpdatePOStatusRequest parameter pembaruan status PO.
type UpdatePOStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=DRAFT SENT_TO_VENDOR SJ_PENDING_VERIFICATION SJ_REVISED READY_FOR_GR GR_COMPLETED"`
}

// POLineResponse data rincian item pesanan PO.
type POLineResponse struct {
	ID        uuid.UUID `json:"id"`
	ItemID    uuid.UUID `json:"item_id"`
	ItemCode  string    `json:"item_code,omitempty"`
	ItemName  string    `json:"item_name,omitempty"`
	Category  string    `json:"category,omitempty"`
	UOM       string    `json:"uom,omitempty"`
	Qty       int       `json:"qty"`
	UnitPrice float64   `json:"unit_price"`
	Subtotal  float64   `json:"subtotal"`
}

// POResponse data Purchase Order lengkap.
type POResponse struct {
	ID              uuid.UUID        `json:"id"`
	PRID            uuid.UUID        `json:"pr_id"`
	PRNumber        string           `json:"pr_number,omitempty"`
	PONumber        string           `json:"po_number"`
	Status          string           `json:"status"`
	TotalAmount     float64          `json:"total_amount"`
	PDFURL          string           `json:"pdf_url,omitempty"`
	VendorID        uuid.UUID        `json:"vendor_id"`
	VendorName      string           `json:"vendor_name,omitempty"`
	VendorEmail     string           `json:"vendor_email,omitempty"`
	VendorPhone     string           `json:"vendor_phone,omitempty"`
	PurchasingID    uuid.UUID        `json:"purchasing_id"`
	PurchasingName  string           `json:"purchasing_name,omitempty"`
	PurchasingEmail string           `json:"purchasing_email,omitempty"`
	TotalItems      int              `json:"total_items"`
	Lines           []POLineResponse `json:"lines,omitempty"`
	CreatedAt       time.Time        `json:"created_at"`
	UpdatedAt       time.Time        `json:"updated_at"`
}

func ToPOLineResponse(l *POLine) POLineResponse {
	itemCode := ""
	itemName := ""
	category := ""
	uom := ""

	if l.Item != nil {
		itemCode = l.Item.Code
		itemName = l.Item.Name
		category = l.Item.Category
		uom = l.Item.UOM
	}

	subtotal := float64(l.Qty) * l.UnitPrice

	return POLineResponse{
		ID:        l.ID,
		ItemID:    l.ItemID,
		ItemCode:  itemCode,
		ItemName:  itemName,
		Category:  category,
		UOM:       uom,
		Qty:       l.Qty,
		UnitPrice: l.UnitPrice,
		Subtotal:  subtotal,
	}
}

func ToPOResponse(po *PurchaseOrder) POResponse {
	prNumber := ""
	if po.PR != nil {
		prNumber = po.PR.PRNumber
	}

	vName := ""
	vEmail := ""
	vPhone := ""
	if po.Vendor != nil {
		vName = po.Vendor.Name
		vEmail = po.Vendor.Email
		vPhone = po.Vendor.ContactPerson
	}

	purchName := ""
	purchEmail := ""
	if po.Purchasing != nil {
		purchName = po.Purchasing.Name
		purchEmail = po.Purchasing.Email
	}

	var linesRes []POLineResponse
	totalItems := 0
	if len(po.Lines) > 0 {
		linesRes = make([]POLineResponse, len(po.Lines))
		for i, line := range po.Lines {
			linesRes[i] = ToPOLineResponse(&line)
			totalItems += line.Qty
		}
	}

	return POResponse{
		ID:              po.ID,
		PRID:            po.PRID,
		PRNumber:        prNumber,
		PONumber:        po.PONumber,
		Status:          po.Status,
		TotalAmount:     po.TotalAmount,
		PDFURL:          po.PDFURL,
		VendorID:        po.VendorID,
		VendorName:      vName,
		VendorEmail:     vEmail,
		VendorPhone:     vPhone,
		PurchasingID:    po.PurchasingID,
		PurchasingName:  purchName,
		PurchasingEmail: purchEmail,
		TotalItems:      totalItems,
		Lines:           linesRes,
		CreatedAt:       po.CreatedAt,
		UpdatedAt:       po.UpdatedAt,
	}
}

func ToPOResponseList(pos []PurchaseOrder) []POResponse {
	res := make([]POResponse, len(pos))
	for i, po := range pos {
		res[i] = ToPOResponse(&po)
	}
	return res
}
