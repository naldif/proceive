package purchaserequisition

import (
	"errors"
	"strings"

	"proceive/internal/modules/item"
	"proceive/internal/modules/user"
	"proceive/pkg/response"
	"proceive/pkg/utils"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrItemNotFound         = errors.New("barang yang dipilih tidak ditemukan dalam katalog")
	ErrCannotApproveOwnPR   = errors.New("pengguna tidak dapat menyetujui atau menolak pengajuan PR miliknya sendiri")
	ErrDifferentDivision    = errors.New("hanya atasan dari divisi yang sama yang dapat menyetujui atau menolak pengajuan PR ini")
	ErrPRNotPending         = errors.New("hanya pengajuan PR berstatus PENDING_APPROVAL yang dapat disetujui atau ditolak")
	ErrPRNotDraft           = errors.New("hanya pengajuan PR berstatus DRAFT yang dapat diperbarui atau dihapus")
	ErrUnauthorizedPRAction = errors.New("anda tidak memiliki hak akses untuk mengubah pengajuan PR ini")
	ErrUnauthorizedPRView   = errors.New("anda tidak memiliki hak akses untuk melihat pengajuan PR ini")
)

type Service interface {
	Create(req *CreatePRRequest, currentUserID uuid.UUID) (*PRResponse, error)
	GetAll(p *utils.Pagination, filter *PRFilter, currentUserID uuid.UUID, userRole string) ([]PRResponse, response.PaginationMeta, error)
	GetByID(id uuid.UUID, currentUserID uuid.UUID, userRole string) (*PRResponse, error)
	Update(id uuid.UUID, req *UpdatePRRequest, currentUserID uuid.UUID, userRole string) (*PRResponse, error)
	Approve(id uuid.UUID, req *ApprovePRRequest, currentUserID uuid.UUID, userRole string) (*PRResponse, error)
	Delete(id uuid.UUID, currentUserID uuid.UUID, userRole string) error
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

// Create membuat Purchase Requisition baru beserta line items.
func (s *service) Create(req *CreatePRRequest, currentUserID uuid.UUID) (*PRResponse, error) {
	prNumber, err := s.repo.GeneratePRNumber()
	if err != nil {
		return nil, err
	}

	status := StatusPendingApproval
	if strings.ToUpper(strings.TrimSpace(req.Status)) == StatusDraft {
		status = StatusDraft
	}

	var lines []PRLine
	for _, itm := range req.Items {
		itemUUID, err := uuid.Parse(itm.ItemID)
		if err != nil {
			return nil, ErrItemNotFound
		}

		var itemObj item.Item
		if err := s.db.First(&itemObj, "id = ?", itemUUID).Error; err != nil {
			return nil, ErrItemNotFound
		}

		estPrice := itemObj.PriceEstimate
		if itm.EstimatedPrice != nil && *itm.EstimatedPrice > 0 {
			estPrice = *itm.EstimatedPrice
		}

		lines = append(lines, PRLine{
			ItemID:         itemUUID,
			Qty:            itm.Qty,
			EstimatedPrice: estPrice,
		})
	}

	pr := PurchaseRequisition{
		RequesterID: currentUserID,
		PRNumber:    prNumber,
		Status:      status,
		Notes:       strings.TrimSpace(req.Notes),
		Lines:       lines,
	}

	if err := s.repo.Create(&pr); err != nil {
		return nil, err
	}

	created, err := s.repo.FindByID(pr.ID)
	if err != nil {
		res := ToPRResponse(&pr)
		return &res, nil
	}

	res := ToPRResponse(created)
	return &res, nil
}

// GetAll mengambil daftar Purchase Requisition berdasarkan hak akses role pengguna.
func (s *service) GetAll(p *utils.Pagination, filter *PRFilter, currentUserID uuid.UUID, userRole string) ([]PRResponse, response.PaginationMeta, error) {
	if filter == nil {
		filter = &PRFilter{}
	}

	var currentUser user.User
	_ = s.db.Select("id", "division_id").First(&currentUser, "id = ?", currentUserID).Error

	normalizedRole := strings.ToUpper(strings.TrimSpace(userRole))
	if normalizedRole == "ADMIN" || normalizedRole == "PURCHASING" {
		// Akses seluruh data PR
	} else if normalizedRole == "APPROVER" {
		if currentUser.DivisionID != nil {
			filter.DivisionID = currentUser.DivisionID
		} else {
			filter.RequesterID = &currentUserID
		}
	} else {
		filter.RequesterID = &currentUserID
	}

	prs, total, err := s.repo.FindAll(p, filter)
	if err != nil {
		return nil, response.PaginationMeta{}, err
	}

	var meta response.PaginationMeta
	if p != nil {
		meta = p.Meta(total)
	}

	return ToPRResponseList(prs), meta, nil
}

// GetByID mengambil detail Purchase Requisition berdasarkan ID.
func (s *service) GetByID(id uuid.UUID, currentUserID uuid.UUID, userRole string) (*PRResponse, error) {
	pr, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	var currentUser user.User
	_ = s.db.Select("id", "division_id").First(&currentUser, "id = ?", currentUserID).Error

	normalizedRole := strings.ToUpper(strings.TrimSpace(userRole))
	if normalizedRole != "ADMIN" && normalizedRole != "PURCHASING" {
		if normalizedRole == "APPROVER" {
			if pr.Requester != nil && pr.Requester.DivisionID != nil && currentUser.DivisionID != nil {
				if *pr.Requester.DivisionID != *currentUser.DivisionID && pr.RequesterID != currentUserID {
					return nil, ErrUnauthorizedPRView
				}
			}
		} else {
			if pr.RequesterID != currentUserID {
				return nil, ErrUnauthorizedPRView
			}
		}
	}

	res := ToPRResponse(pr)
	return &res, nil
}

// Update memperbarui data Purchase Requisition yang masih berstatus draft.
func (s *service) Update(id uuid.UUID, req *UpdatePRRequest, currentUserID uuid.UUID, userRole string) (*PRResponse, error) {
	pr, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	if pr.Status != StatusDraft {
		return nil, ErrPRNotDraft
	}

	normalizedRole := strings.ToUpper(strings.TrimSpace(userRole))
	if pr.RequesterID != currentUserID && normalizedRole != "ADMIN" {
		return nil, ErrUnauthorizedPRAction
	}

	pr.Notes = strings.TrimSpace(req.Notes)

	if len(req.Items) > 0 {
		var newLines []PRLine
		for _, itm := range req.Items {
			itemUUID, err := uuid.Parse(itm.ItemID)
			if err != nil {
				return nil, ErrItemNotFound
			}

			var itemObj item.Item
			if err := s.db.First(&itemObj, "id = ?", itemUUID).Error; err != nil {
				return nil, ErrItemNotFound
			}

			estPrice := itemObj.PriceEstimate
			if itm.EstimatedPrice != nil && *itm.EstimatedPrice > 0 {
				estPrice = *itm.EstimatedPrice
			}

			newLines = append(newLines, PRLine{
				PRID:           id,
				ItemID:         itemUUID,
				Qty:            itm.Qty,
				EstimatedPrice: estPrice,
			})
		}

		if err := s.repo.ReplaceLines(id, newLines); err != nil {
			return nil, err
		}
	}

	if err := s.repo.Update(pr); err != nil {
		return nil, err
	}

	updated, err := s.repo.FindByID(id)
	if err != nil {
		res := ToPRResponse(pr)
		return &res, nil
	}

	res := ToPRResponse(updated)
	return &res, nil
}

// Approve memproses persetujuan atau penolakan pengajuan PR oleh approver.
func (s *service) Approve(id uuid.UUID, req *ApprovePRRequest, currentUserID uuid.UUID, userRole string) (*PRResponse, error) {
	pr, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	if pr.Status != StatusPendingApproval {
		return nil, ErrPRNotPending
	}

	// Tidak boleh menyetujui pengajuan milik sendiri
	if pr.RequesterID == currentUserID {
		return nil, ErrCannotApproveOwnPR
	}

	// Approver harus satu divisi dengan requester (kecuali ADMIN)
	normalizedRole := strings.ToUpper(strings.TrimSpace(userRole))
	if normalizedRole != "ADMIN" {
		var approverUser user.User
		if err := s.db.First(&approverUser, "id = ?", currentUserID).Error; err != nil {
			return nil, err
		}

		var reqUser user.User
		if err := s.db.First(&reqUser, "id = ?", pr.RequesterID).Error; err != nil {
			return nil, err
		}

		if reqUser.DivisionID == nil || approverUser.DivisionID == nil || *reqUser.DivisionID != *approverUser.DivisionID {
			return nil, ErrDifferentDivision
		}
	}

	targetStatus := strings.ToUpper(strings.TrimSpace(req.Status))
	if targetStatus != StatusApproved && targetStatus != StatusRejected {
		targetStatus = StatusApproved
	}

	pr.Status = targetStatus
	pr.ApproverID = &currentUserID
	pr.ApprovalNotes = strings.TrimSpace(req.Notes)

	if err := s.repo.Update(pr); err != nil {
		return nil, err
	}

	updated, err := s.repo.FindByID(id)
	if err != nil {
		res := ToPRResponse(pr)
		return &res, nil
	}

	res := ToPRResponse(updated)
	return &res, nil
}

// Delete menghapus pengajuan Purchase Requisition yang masih berstatus draft.
func (s *service) Delete(id uuid.UUID, currentUserID uuid.UUID, userRole string) error {
	pr, err := s.repo.FindByID(id)
	if err != nil {
		return err
	}

	if pr.Status != StatusDraft {
		return ErrPRNotDraft
	}

	normalizedRole := strings.ToUpper(strings.TrimSpace(userRole))
	if pr.RequesterID != currentUserID && normalizedRole != "ADMIN" {
		return ErrUnauthorizedPRAction
	}

	return s.repo.Delete(id)
}
