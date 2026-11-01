package goodsreceipt

import (
	"time"

	"github.com/google/uuid"
)

// UploadSuratJalanRequest parameter nomor surat jalan dari vendor.
type UploadSuratJalanRequest struct {
	SuratJalanNo string `form:"surat_jalan_no" binding:"required,max=100"`
}

// VerifySuratJalanRequest parameter verifikasi surat jalan oleh purchasing.
type VerifySuratJalanRequest struct {
	Status            string `json:"status" binding:"required,oneof=SJ_APPROVED SJ_REVISED"`
	VerificationNotes string `json:"verification_notes" binding:"omitempty,max=1000"`
}

// CompleteGRRequest parameter catatan penerimaan fisik barang oleh PPIC.
type CompleteGRRequest struct {
	Notes string `json:"notes" form:"notes" binding:"omitempty,max=1000"`
}

// GoodsReceiptResponse rincian data Goods Receipt.
type GoodsReceiptResponse struct {
	ID                uuid.UUID  `json:"id"`
	POID              uuid.UUID  `json:"po_id"`
	PONumber          string     `json:"po_number,omitempty"`
	PRNumber          string     `json:"pr_number,omitempty"`
	VendorID          *uuid.UUID `json:"vendor_id,omitempty"`
	VendorName        string     `json:"vendor_name,omitempty"`
	SuratJalanNo      string     `json:"surat_jalan_no"`
	SuratJalanURL     string     `json:"surat_jalan_url"`
	GRDocumentURL     string     `json:"gr_document_url,omitempty"`
	VerificationNotes string     `json:"verification_notes,omitempty"`
	Notes             string     `json:"notes,omitempty"`
	GRNumber          *string    `json:"gr_number,omitempty"`
	Status            string     `json:"status"`
	PurchasingID      *uuid.UUID `json:"purchasing_id,omitempty"`
	PurchasingName    string     `json:"purchasing_name,omitempty"`
	PPICID            *uuid.UUID `json:"ppic_id,omitempty"`
	PPICName          string     `json:"ppic_name,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

func ToGoodsReceiptResponse(gr *GoodsReceipt) GoodsReceiptResponse {
	poNumber := ""
	prNumber := ""
	var vendorID *uuid.UUID
	vendorName := ""

	if gr.PO != nil {
		poNumber = gr.PO.PONumber
		vendorID = &gr.PO.VendorID
		if gr.PO.Vendor != nil {
			vendorName = gr.PO.Vendor.Name
		}
		if gr.PO.PR != nil {
			prNumber = gr.PO.PR.PRNumber
		}
	}

	purchName := ""
	if gr.Purchasing != nil {
		purchName = gr.Purchasing.Name
	}

	ppicName := ""
	if gr.PPIC != nil {
		ppicName = gr.PPIC.Name
	}

	return GoodsReceiptResponse{
		ID:                gr.ID,
		POID:              gr.POID,
		PONumber:          poNumber,
		PRNumber:          prNumber,
		VendorID:          vendorID,
		VendorName:        vendorName,
		SuratJalanNo:      gr.SuratJalanNo,
		SuratJalanURL:     gr.SuratJalanURL,
		GRDocumentURL:     gr.GRDocumentURL,
		VerificationNotes: gr.VerificationNotes,
		Notes:             gr.Notes,
		GRNumber:          gr.GRNumber,
		Status:            gr.Status,
		PurchasingID:      gr.PurchasingID,
		PurchasingName:    purchName,
		PPICID:            gr.PPICID,
		PPICName:          ppicName,
		CreatedAt:         gr.CreatedAt,
		UpdatedAt:         gr.UpdatedAt,
	}
}

func ToGoodsReceiptResponseList(grs []GoodsReceipt) []GoodsReceiptResponse {
	res := make([]GoodsReceiptResponse, len(grs))
	for i, gr := range grs {
		res[i] = ToGoodsReceiptResponse(&gr)
	}
	return res
}
