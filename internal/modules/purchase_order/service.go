package purchaseorder

import (
	"errors"
	"fmt"
	"strings"

	purchaserequisition "proceive/internal/modules/purchase_requisition"
	"proceive/internal/modules/user"
	"proceive/internal/modules/vendor"
	"proceive/pkg/response"
	"proceive/pkg/utils"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrPRNotFound           = errors.New("purchase requisition tidak ditemukan")
	ErrPRNotApproved        = errors.New("purchase requisition belum disetujui, hanya PR berstatus APPROVED yang dapat diterbitkan menjadi PO")
	ErrPOAlreadyExists      = errors.New("purchase order untuk pengajuan PR ini sudah pernah diterbitkan")
	ErrVendorNotFound       = errors.New("vendor yang dipilih tidak ditemukan")
	ErrVendorNotActive      = errors.New("vendor yang dipilih sedang tidak aktif")
	ErrUnauthorizedPOView   = errors.New("anda tidak memiliki hak akses untuk melihat purchase order ini")
	ErrUnauthorizedPOAction = errors.New("anda tidak memiliki hak akses untuk mengubah purchase order ini")
	ErrPONotDraft           = errors.New("hanya purchase order berstatus DRAFT yang dapat dihapus")
)

type Service interface {
	CreateFromPR(prID uuid.UUID, req *CreatePOFromPRRequest, currentUserID uuid.UUID) (*POResponse, error)
	GetAll(p *utils.Pagination, filter *POFilter, currentUserID uuid.UUID, userRole string) ([]POResponse, response.PaginationMeta, error)
	GetByID(id uuid.UUID, currentUserID uuid.UUID, userRole string) (*POResponse, error)
	UpdateStatus(id uuid.UUID, req *UpdatePOStatusRequest, currentUserID uuid.UUID, userRole string) (*POResponse, error)
	Delete(id uuid.UUID, currentUserID uuid.UUID, userRole string) error
	GeneratePDF(poNumber string) ([]byte, *PurchaseOrder, error)
}

type service struct {
	repo Repository
	db   *gorm.DB
}

func NewService(repo Repository, db *gorm.DB) Service {
	return &service{
		repo: repo,
		db:   db,
	}
}

// CreateFromPR membuat PO baru berdasarkan PR yang telah disetujui.
func (s *service) CreateFromPR(prID uuid.UUID, req *CreatePOFromPRRequest, currentUserID uuid.UUID) (*POResponse, error) {
	// Validasi status PR
	var pr purchaserequisition.PurchaseRequisition
	if err := s.db.Preload("Lines").Preload("Lines.Item").First(&pr, "id = ?", prID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPRNotFound
		}
		return nil, err
	}

	if pr.Status != purchaserequisition.StatusApproved {
		return nil, ErrPRNotApproved
	}

	// Cegah duplikasi PO dari PR yang sama
	existingPO, err := s.repo.FindByPRID(prID)
	if err != nil {
		return nil, err
	}
	if existingPO != nil {
		return nil, ErrPOAlreadyExists
	}

	// Validasi data vendor
	vendorUUID, err := uuid.Parse(req.VendorID)
	if err != nil {
		return nil, ErrVendorNotFound
	}

	var v vendor.Vendor
	if err := s.db.First(&v, "id = ?", vendorUUID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrVendorNotFound
		}
		return nil, err
	}
	if !v.IsActive {
		return nil, ErrVendorNotActive
	}

	poNumber, err := s.repo.GeneratePONumber()
	if err != nil {
		return nil, err
	}

	// Salin rincian item barang dari PR
	if len(pr.Lines) == 0 {
		return nil, errors.New("purchase requisition tidak memiliki rincian barang di pr_lines")
	}

	var poLines []POLine
	var totalAmount float64

	for _, prLine := range pr.Lines {
		unitPrice := prLine.EstimatedPrice
		subtotal := float64(prLine.Qty) * unitPrice
		totalAmount += subtotal

		poLines = append(poLines, POLine{
			ItemID:    prLine.ItemID,
			Qty:       prLine.Qty,
			UnitPrice: unitPrice,
		})
	}

	// 6. Siapkan entitas PurchaseOrder
	pdfURL := fmt.Sprintf("/api/v1/purchase-orders/documents/%s.pdf", poNumber)

	po := PurchaseOrder{
		PRID:         prID,
		VendorID:     vendorUUID,
		PurchasingID: currentUserID,
		PONumber:     poNumber,
		Status:       StatusSentToVendor, // Langsung diterbitkan & dikirim ke vendor
		TotalAmount:  totalAmount,
		PDFURL:       pdfURL,
		Lines:        poLines,
	}

	if err := s.repo.Create(&po); err != nil {
		return nil, err
	}

	// Ambil data lengkap dengan preloads
	created, err := s.repo.FindByID(po.ID)
	if err != nil {
		res := ToPOResponse(&po)
		return &res, nil
	}

	res := ToPOResponse(created)
	return &res, nil
}

// GetAll mengambil daftar Purchase Order dengan filter dan pagination.
func (s *service) GetAll(p *utils.Pagination, filter *POFilter, currentUserID uuid.UUID, userRole string) ([]POResponse, response.PaginationMeta, error) {
	if filter == nil {
		filter = &POFilter{}
	}

	normalizedRole := strings.ToUpper(strings.TrimSpace(userRole))
	if normalizedRole == "VENDOR" {
		var currentUser user.User
		if err := s.db.Select("id", "vendor_id").First(&currentUser, "id = ?", currentUserID).Error; err != nil {
			return nil, response.PaginationMeta{}, err
		}

		if currentUser.VendorID == nil {
			return []POResponse{}, response.PaginationMeta{}, nil
		}
		filter.VendorID = currentUser.VendorID
	}

	pos, total, err := s.repo.FindAll(p, filter)
	if err != nil {
		return nil, response.PaginationMeta{}, err
	}

	var meta response.PaginationMeta
	if p != nil {
		meta = p.Meta(total)
	}

	return ToPOResponseList(pos), meta, nil
}

// GetByID mengambil detail Purchase Order berdasarkan ID.
func (s *service) GetByID(id uuid.UUID, currentUserID uuid.UUID, userRole string) (*POResponse, error) {
	po, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	normalizedRole := strings.ToUpper(strings.TrimSpace(userRole))
	if normalizedRole == "VENDOR" {
		var currentUser user.User
		if err := s.db.Select("id", "vendor_id").First(&currentUser, "id = ?", currentUserID).Error; err != nil {
			return nil, err
		}

		if currentUser.VendorID == nil || *currentUser.VendorID != po.VendorID {
			return nil, ErrUnauthorizedPOView
		}
	}

	res := ToPOResponse(po)
	return &res, nil
}

// UpdateStatus memperbarui status Purchase Order.
func (s *service) UpdateStatus(id uuid.UUID, req *UpdatePOStatusRequest, currentUserID uuid.UUID, userRole string) (*POResponse, error) {
	po, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	targetStatus := strings.ToUpper(strings.TrimSpace(req.Status))
	po.Status = targetStatus

	if err := s.repo.Update(po); err != nil {
		return nil, err
	}

	updated, err := s.repo.FindByID(id)
	if err != nil {
		res := ToPOResponse(po)
		return &res, nil
	}

	res := ToPOResponse(updated)
	return &res, nil
}

// Delete menghapus Purchase Order yang masih berstatus draft.
func (s *service) Delete(id uuid.UUID, currentUserID uuid.UUID, userRole string) error {
	po, err := s.repo.FindByID(id)
	if err != nil {
		return err
	}

	if po.Status != StatusDraft {
		return ErrPONotDraft
	}

	return s.repo.Delete(id)
}

// GeneratePDF menghasilkan file PDF untuk nomor PO tertentu.
func (s *service) GeneratePDF(poNumber string) ([]byte, *PurchaseOrder, error) {
	po, err := s.repo.FindByPONumber(poNumber)
	if err != nil {
		return nil, nil, err
	}

	pdfBytes := GeneratePOPDF(po)
	return pdfBytes, po, nil
}
