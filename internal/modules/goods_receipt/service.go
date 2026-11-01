package goodsreceipt

import (
	"context"
	"errors"
	"fmt"
	"mime/multipart"
	"path/filepath"
	"strings"

	"proceive/internal/config"
	purchaseorder "proceive/internal/modules/purchase_order"
	"proceive/internal/modules/user"
	"proceive/pkg/response"
	"proceive/pkg/utils"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrPONotFound           = errors.New("purchase order tidak ditemukan")
	ErrPOInvalidStatusForSJ = errors.New("surat jalan hanya dapat diunggah untuk PO yang berstatus SENT_TO_VENDOR atau SJ_REVISED")
	ErrUnauthorizedVendor   = errors.New("anda hanya memiliki akses untuk mengunggah surat jalan pesanan milik perusahaan anda")
	ErrUnauthorizedGRView   = errors.New("anda tidak memiliki hak akses untuk melihat data goods receipt ini")
	ErrGRNotUploaded        = errors.New("hanya surat jalan yang sudah diunggah (SJ_UPLOADED) yang dapat diverifikasi")
	ErrGRNotApproved        = errors.New("surat jalan belum disetujui (SJ_APPROVED) oleh purchasing, belum dapat menyelesaikan penerimaan fisik barang")
	ErrGRAlreadyCompleted   = errors.New("proses penerimaan barang (goods receipt) sudah selesai")
)

type Service interface {
	UploadSuratJalan(ctx context.Context, poID uuid.UUID, req *UploadSuratJalanRequest, fileHeader *multipart.FileHeader, currentUserID uuid.UUID, userRole string) (*GoodsReceiptResponse, error)
	VerifySuratJalan(id uuid.UUID, req *VerifySuratJalanRequest, currentUserID uuid.UUID, userRole string) (*GoodsReceiptResponse, error)
	CompleteGR(ctx context.Context, id uuid.UUID, req *CompleteGRRequest, fileHeader *multipart.FileHeader, currentUserID uuid.UUID, userRole string) (*GoodsReceiptResponse, error)
	GetAll(p *utils.Pagination, filter *GRFilter, currentUserID uuid.UUID, userRole string) ([]GoodsReceiptResponse, response.PaginationMeta, error)
	GetByID(id uuid.UUID, currentUserID uuid.UUID, userRole string) (*GoodsReceiptResponse, error)
}

type service struct {
	repo Repository
	db   *gorm.DB
	cfg  *config.Config
}

func NewService(repo Repository, db *gorm.DB, cfg *config.Config) Service {
	return &service{
		repo: repo,
		db:   db,
		cfg:  cfg,
	}
}

// UploadSuratJalan memproses upload surat jalan dari vendor untuk PO yang sedang berjalan.
func (s *service) UploadSuratJalan(ctx context.Context, poID uuid.UUID, req *UploadSuratJalanRequest, fileHeader *multipart.FileHeader, currentUserID uuid.UUID, userRole string) (*GoodsReceiptResponse, error) {
	var po purchaseorder.PurchaseOrder
	if err := s.db.First(&po, "id = ?", poID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPONotFound
		}
		return nil, err
	}

	if po.Status != purchaseorder.StatusSentToVendor && po.Status != purchaseorder.StatusSJRevised && po.Status != purchaseorder.StatusSJPendingVerification {
		return nil, ErrPOInvalidStatusForSJ
	}

	normalizedRole := strings.ToUpper(strings.TrimSpace(userRole))
	if normalizedRole == "VENDOR" {
		var currentUser user.User
		if err := s.db.Select("id", "vendor_id").First(&currentUser, "id = ?", currentUserID).Error; err != nil {
			return nil, err
		}

		if currentUser.VendorID == nil || *currentUser.VendorID != po.VendorID {
			return nil, ErrUnauthorizedVendor
		}
	}

	if err := utils.ValidateDocumentFile(fileHeader); err != nil {
		return nil, err
	}

	src, err := fileHeader.Open()
	if err != nil {
		return nil, err
	}
	defer src.Close()

	var fileURL string
	if localPath, localErr := utils.SaveLocalFile(fileHeader, "storage", "proceive", "surat_jalan"); localErr == nil {
		fileURL = "/storage/" + filepath.ToSlash(localPath)
	}

	cloudURL, err := utils.UploadToCloudinary(ctx, src, fileHeader.Filename, s.cfg.Cloudinary.CloudName, s.cfg.Cloudinary.APIKey, s.cfg.Cloudinary.APISecret, "proceive", "surat_jalan")
	if err == nil && cloudURL != "" {
		fileURL = cloudURL
	}

	for strings.Contains(fileURL, ".pdf.pdf") {
		fileURL = strings.ReplaceAll(fileURL, ".pdf.pdf", ".pdf")
	}

	if fileURL == "" {
		return nil, fmt.Errorf("gagal mengunggah dokumen surat jalan: %w", err)
	}

	var res *GoodsReceipt
	err = s.db.Transaction(func(tx *gorm.DB) error {
		existingGR, err := s.repo.FindByPOID(poID)
		if err != nil {
			return err
		}

		if existingGR != nil {
			existingGR.SuratJalanNo = strings.TrimSpace(req.SuratJalanNo)
			existingGR.SuratJalanURL = fileURL
			existingGR.Status = StatusSJUploaded
			existingGR.VerificationNotes = ""

			if err := tx.Save(existingGR).Error; err != nil {
				return err
			}
			res = existingGR
		} else {
			newGR := GoodsReceipt{
				POID:          poID,
				SuratJalanNo:  strings.TrimSpace(req.SuratJalanNo),
				SuratJalanURL: fileURL,
				Status:        StatusSJUploaded,
			}
			if err := tx.Create(&newGR).Error; err != nil {
				return err
			}
			res = &newGR
		}

		if err := tx.Model(&purchaseorder.PurchaseOrder{}).Where("id = ?", poID).Update("status", purchaseorder.StatusSJPendingVerification).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	created, err := s.repo.FindByID(res.ID)
	if err != nil {
		out := ToGoodsReceiptResponse(res)
		return &out, nil
	}

	out := ToGoodsReceiptResponse(created)
	return &out, nil
}

// VerifySuratJalan memproses persetujuan atau permintaan revisi surat jalan oleh purchasing.
func (s *service) VerifySuratJalan(id uuid.UUID, req *VerifySuratJalanRequest, currentUserID uuid.UUID, userRole string) (*GoodsReceiptResponse, error) {
	gr, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	if gr.Status != StatusSJUploaded {
		return nil, ErrGRNotUploaded
	}

	targetStatus := strings.ToUpper(strings.TrimSpace(req.Status))
	if targetStatus != StatusSJApproved && targetStatus != StatusSJRevised {
		targetStatus = StatusSJApproved
	}

	poTargetStatus := purchaseorder.StatusReadyForGR
	if targetStatus == StatusSJRevised {
		poTargetStatus = purchaseorder.StatusSJRevised
	}

	err = s.db.Transaction(func(tx *gorm.DB) error {
		gr.Status = targetStatus
		gr.PurchasingID = &currentUserID
		gr.VerificationNotes = strings.TrimSpace(req.VerificationNotes)

		if err := tx.Save(gr).Error; err != nil {
			return err
		}

		if err := tx.Model(&purchaseorder.PurchaseOrder{}).Where("id = ?", gr.POID).Update("status", poTargetStatus).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	updated, err := s.repo.FindByID(id)
	if err != nil {
		out := ToGoodsReceiptResponse(gr)
		return &out, nil
	}

	out := ToGoodsReceiptResponse(updated)
	return &out, nil
}

// CompleteGR menyelesaikan inspeksi penerimaan fisik barang dan menerbitkan nomor dokumen GR.
func (s *service) CompleteGR(ctx context.Context, id uuid.UUID, req *CompleteGRRequest, fileHeader *multipart.FileHeader, currentUserID uuid.UUID, userRole string) (*GoodsReceiptResponse, error) {
	gr, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	if gr.Status == StatusGRCompleted {
		return nil, ErrGRAlreadyCompleted
	}

	if gr.Status != StatusSJApproved {
		return nil, ErrGRNotApproved
	}

	var grFileURL string
	if fileHeader != nil {
		if err := utils.ValidateDocumentFile(fileHeader); err != nil {
			return nil, err
		}

		if localPath, localErr := utils.SaveLocalFile(fileHeader, "storage", "proceive", "goods_receipt"); localErr == nil {
			grFileURL = "/storage/" + filepath.ToSlash(localPath)
		}

		src, err := fileHeader.Open()
		if err == nil {
			defer src.Close()
			cloudURL, err := utils.UploadToCloudinary(ctx, src, fileHeader.Filename, s.cfg.Cloudinary.CloudName, s.cfg.Cloudinary.APIKey, s.cfg.Cloudinary.APISecret, "proceive", "goods_receipt")
			if err == nil && cloudURL != "" {
				grFileURL = cloudURL
			}
		}

		for strings.Contains(grFileURL, ".pdf.pdf") {
			grFileURL = strings.ReplaceAll(grFileURL, ".pdf.pdf", ".pdf")
		}
	}

	grNumber, err := s.repo.GenerateGRNumber()
	if err != nil {
		return nil, err
	}

	err = s.db.Transaction(func(tx *gorm.DB) error {
		gr.GRNumber = &grNumber
		gr.PPICID = &currentUserID
		gr.Status = StatusGRCompleted
		gr.Notes = strings.TrimSpace(req.Notes)
		if grFileURL != "" {
			gr.GRDocumentURL = grFileURL
		}

		if err := tx.Save(gr).Error; err != nil {
			return err
		}

		if err := tx.Model(&purchaseorder.PurchaseOrder{}).Where("id = ?", gr.POID).Update("status", purchaseorder.StatusGRCompleted).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	updated, err := s.repo.FindByID(id)
	if err != nil {
		out := ToGoodsReceiptResponse(gr)
		return &out, nil
	}

	out := ToGoodsReceiptResponse(updated)
	return &out, nil
}

// GetAll mengambil daftar Goods Receipt berdasarkan filter status dan hak akses role.
func (s *service) GetAll(p *utils.Pagination, filter *GRFilter, currentUserID uuid.UUID, userRole string) ([]GoodsReceiptResponse, response.PaginationMeta, error) {
	if filter == nil {
		filter = &GRFilter{}
	}

	normalizedRole := strings.ToUpper(strings.TrimSpace(userRole))
	if normalizedRole == "VENDOR" {
		var currentUser user.User
		if err := s.db.Select("id", "vendor_id").First(&currentUser, "id = ?", currentUserID).Error; err != nil {
			return nil, response.PaginationMeta{}, err
		}

		if currentUser.VendorID == nil {
			return []GoodsReceiptResponse{}, response.PaginationMeta{}, nil
		}
		filter.VendorID = currentUser.VendorID
	}

	grs, total, err := s.repo.FindAll(p, filter)
	if err != nil {
		return nil, response.PaginationMeta{}, err
	}

	var meta response.PaginationMeta
	if p != nil {
		meta = p.Meta(total)
	}

	return ToGoodsReceiptResponseList(grs), meta, nil
}

// GetByID mengambil detail Goods Receipt berdasarkan ID dengan pemeriksaan hak akses.
func (s *service) GetByID(id uuid.UUID, currentUserID uuid.UUID, userRole string) (*GoodsReceiptResponse, error) {
	gr, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	normalizedRole := strings.ToUpper(strings.TrimSpace(userRole))
	if normalizedRole == "VENDOR" {
		var currentUser user.User
		if err := s.db.Select("id", "vendor_id").First(&currentUser, "id = ?", currentUserID).Error; err != nil {
			return nil, err
		}

		if currentUser.VendorID == nil || gr.PO == nil || *currentUser.VendorID != gr.PO.VendorID {
			return nil, ErrUnauthorizedGRView
		}
	}

	res := ToGoodsReceiptResponse(gr)
	return &res, nil
}
