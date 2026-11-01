package purchaserequisition

import (
	"time"

	"github.com/google/uuid"
)

// CreatePRLineRequest item barang yang diminta dalam PR.
type CreatePRLineRequest struct {
	ItemID         string   `json:"item_id" binding:"required,uuid"`
	Qty            int      `json:"qty" binding:"required,min=1"`
	EstimatedPrice *float64 `json:"estimated_price" binding:"omitempty,min=0"`
}

// CreatePRRequest parameter pembuatan pengajuan PR baru.
type CreatePRRequest struct {
	Notes  string                `json:"notes" binding:"omitempty,max=1000"`
	Status string                `json:"status" binding:"omitempty,oneof=DRAFT PENDING_APPROVAL"`
	Items  []CreatePRLineRequest `json:"items" binding:"required,min=1,dive"`
}

// UpdatePRRequest parameter pembaruan PR draft.
type UpdatePRRequest struct {
	Notes string                `json:"notes" binding:"omitempty,max=1000"`
	Items []CreatePRLineRequest `json:"items" binding:"omitempty,min=1,dive"`
}

// ApprovePRRequest parameter persetujuan atau penolakan PR oleh approver.
type ApprovePRRequest struct {
	Status string `json:"status" binding:"omitempty,oneof=APPROVED REJECTED"`
	Notes  string `json:"notes" binding:"omitempty,max=1000"`
}

// PRLineResponse rincian item dalam Purchase Requisition.
type PRLineResponse struct {
	ID             uuid.UUID `json:"id"`
	ItemID         uuid.UUID `json:"item_id"`
	ItemCode       string    `json:"item_code,omitempty"`
	ItemName       string    `json:"item_name,omitempty"`
	Category       string    `json:"category,omitempty"`
	UOM            string    `json:"uom,omitempty"`
	Qty            int       `json:"qty"`
	EstimatedPrice float64   `json:"estimated_price"`
	Subtotal       float64   `json:"subtotal"`
}

// PRResponse rincian Purchase Requisition lengkap.
type PRResponse struct {
	ID                  uuid.UUID        `json:"id"`
	PRNumber            string           `json:"pr_number"`
	Status              string           `json:"status"`
	Notes               string           `json:"notes"`
	ApprovalNotes       string           `json:"approval_notes,omitempty"`
	RequesterID         uuid.UUID        `json:"requester_id"`
	RequesterName       string           `json:"requester_name,omitempty"`
	RequesterEmail      string           `json:"requester_email,omitempty"`
	DivisionID          *uuid.UUID       `json:"division_id,omitempty"`
	DivisionName        string           `json:"division_name,omitempty"`
	ApproverID          *uuid.UUID       `json:"approver_id,omitempty"`
	ApproverName        string           `json:"approver_name,omitempty"`
	TotalEstimatedPrice float64          `json:"total_estimated_price"`
	TotalItems          int              `json:"total_items"`
	Lines               []PRLineResponse `json:"lines,omitempty"`
	CreatedAt           time.Time        `json:"created_at"`
	UpdatedAt           time.Time        `json:"updated_at"`
}

func ToPRLineResponse(l *PRLine) PRLineResponse {
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

	subtotal := float64(l.Qty) * l.EstimatedPrice

	return PRLineResponse{
		ID:             l.ID,
		ItemID:         l.ItemID,
		ItemCode:       itemCode,
		ItemName:       itemName,
		Category:       category,
		UOM:            uom,
		Qty:            l.Qty,
		EstimatedPrice: l.EstimatedPrice,
		Subtotal:       subtotal,
	}
}

func ToPRResponse(pr *PurchaseRequisition) PRResponse {
	reqName := ""
	reqEmail := ""
	var divID *uuid.UUID
	divName := ""

	if pr.Requester != nil {
		reqName = pr.Requester.Name
		reqEmail = pr.Requester.Email
		divID = pr.Requester.DivisionID
		if pr.Requester.Division != nil {
			divName = pr.Requester.Division.Name
		}
	}

	approverName := ""
	if pr.Approver != nil {
		approverName = pr.Approver.Name
	}

	var linesRes []PRLineResponse
	var totalEstPrice float64
	totalItems := 0

	if len(pr.Lines) > 0 {
		linesRes = make([]PRLineResponse, len(pr.Lines))
		for i, line := range pr.Lines {
			linesRes[i] = ToPRLineResponse(&line)
			totalEstPrice += linesRes[i].Subtotal
			totalItems += line.Qty
		}
	}

	return PRResponse{
		ID:                  pr.ID,
		PRNumber:            pr.PRNumber,
		Status:              pr.Status,
		Notes:               pr.Notes,
		ApprovalNotes:       pr.ApprovalNotes,
		RequesterID:         pr.RequesterID,
		RequesterName:       reqName,
		RequesterEmail:      reqEmail,
		DivisionID:          divID,
		DivisionName:        divName,
		ApproverID:          pr.ApproverID,
		ApproverName:        approverName,
		TotalEstimatedPrice: totalEstPrice,
		TotalItems:          totalItems,
		Lines:               linesRes,
		CreatedAt:           pr.CreatedAt,
		UpdatedAt:           pr.UpdatedAt,
	}
}

func ToPRResponseList(prs []PurchaseRequisition) []PRResponse {
	res := make([]PRResponse, len(prs))
	for i, pr := range prs {
		res[i] = ToPRResponse(&pr)
	}
	return res
}
