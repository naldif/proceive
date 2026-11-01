package goodsreceipt

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"proceive/pkg/utils"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrGRNotFound = errors.New("goods receipt tidak ditemukan")
)

type GRFilter struct {
	Status   string
	VendorID *uuid.UUID
}

type Repository interface {
	Create(gr *GoodsReceipt) error
	FindAll(p *utils.Pagination, filter *GRFilter) ([]GoodsReceipt, int64, error)
	FindByID(id uuid.UUID) (*GoodsReceipt, error)
	FindByPOID(poID uuid.UUID) (*GoodsReceipt, error)
	Update(gr *GoodsReceipt) error
	GenerateGRNumber() (string, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) GenerateGRNumber() (string, error) {
	prefix := fmt.Sprintf("GR-%s-", time.Now().Format("200601"))
	var count int64
	if err := r.db.Model(&GoodsReceipt{}).Where("gr_number LIKE ?", prefix+"%").Count(&count).Error; err != nil {
		return "", err
	}

	seq := count + 1
	for {
		candidate := fmt.Sprintf("%s%04d", prefix, seq)
		var exists int64
		if err := r.db.Model(&GoodsReceipt{}).Where("gr_number = ?", candidate).Count(&exists).Error; err != nil {
			return "", err
		}
		if exists == 0 {
			return candidate, nil
		}
		seq++
	}
}

func (r *repository) Create(gr *GoodsReceipt) error {
	return r.db.Create(gr).Error
}

func (r *repository) FindAll(p *utils.Pagination, filter *GRFilter) ([]GoodsReceipt, int64, error) {
	var grs []GoodsReceipt
	var total int64

	query := r.db.Model(&GoodsReceipt{}).
		Joins("LEFT JOIN purchase_orders ON purchase_orders.id = goods_receipts.po_id").
		Joins("LEFT JOIN vendors ON vendors.id = purchase_orders.vendor_id")

	if filter != nil {
		if filter.Status != "" {
			query = query.Where("goods_receipts.status = ?", strings.ToUpper(strings.TrimSpace(filter.Status)))
		}
		if filter.VendorID != nil {
			query = query.Where("purchase_orders.vendor_id = ?", *filter.VendorID)
		}
	}

	if p != nil && p.Search != "" {
		pattern := "%" + strings.ToLower(p.Search) + "%"
		query = query.Where("LOWER(goods_receipts.surat_jalan_no) LIKE ? OR LOWER(goods_receipts.gr_number) LIKE ? OR LOWER(purchase_orders.po_number) LIKE ? OR LOWER(vendors.name) LIKE ?",
			pattern, pattern, pattern, pattern)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if p != nil {
		sortBy := "goods_receipts.created_at"
		switch strings.ToLower(p.SortBy) {
		case "surat_jalan_no", "gr_number", "status", "created_at", "updated_at":
			sortBy = "goods_receipts." + strings.ToLower(p.SortBy)
		}

		order := "DESC"
		if strings.ToUpper(p.Order) == "ASC" {
			order = "ASC"
		}

		query = query.Order(sortBy + " " + order).Offset(p.Offset()).Limit(p.Limit)
	} else {
		query = query.Order("goods_receipts.created_at DESC")
	}

	err := query.
		Preload("PO").
		Preload("PO.Vendor").
		Preload("PO.PR").
		Preload("Purchasing").
		Preload("PPIC").
		Find(&grs).Error

	if err != nil {
		return nil, 0, err
	}

	return grs, total, nil
}

func (r *repository) FindByID(id uuid.UUID) (*GoodsReceipt, error) {
	var gr GoodsReceipt
	err := r.db.
		Preload("PO").
		Preload("PO.Vendor").
		Preload("PO.Purchasing").
		Preload("PO.PR").
		Preload("PO.Lines").
		Preload("PO.Lines.Item").
		Preload("Purchasing").
		Preload("PPIC").
		First(&gr, "id = ?", id).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrGRNotFound
		}
		return nil, err
	}

	return &gr, nil
}

func (r *repository) FindByPOID(poID uuid.UUID) (*GoodsReceipt, error) {
	var gr GoodsReceipt
	err := r.db.
		Preload("PO").
		Preload("PO.Vendor").
		Preload("PO.PR").
		Preload("Purchasing").
		Preload("PPIC").
		First(&gr, "po_id = ?", poID).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &gr, nil
}

func (r *repository) Update(gr *GoodsReceipt) error {
	return r.db.Save(gr).Error
}
