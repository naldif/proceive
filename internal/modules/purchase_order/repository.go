package purchaseorder

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
	ErrPONotFound = errors.New("purchase order tidak ditemukan")
)

type POFilter struct {
	Status       string
	VendorID     *uuid.UUID
	PurchasingID *uuid.UUID
}

type Repository interface {
	Create(po *PurchaseOrder) error
	FindAll(p *utils.Pagination, filter *POFilter) ([]PurchaseOrder, int64, error)
	FindByID(id uuid.UUID) (*PurchaseOrder, error)
	FindByPRID(prID uuid.UUID) (*PurchaseOrder, error)
	FindByPONumber(poNumber string) (*PurchaseOrder, error)
	Update(po *PurchaseOrder) error
	Delete(id uuid.UUID) error
	GeneratePONumber() (string, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) GeneratePONumber() (string, error) {
	prefix := fmt.Sprintf("PO-%s-", time.Now().Format("200601"))
	var count int64
	if err := r.db.Model(&PurchaseOrder{}).Where("po_number LIKE ?", prefix+"%").Count(&count).Error; err != nil {
		return "", err
	}

	seq := count + 1
	for {
		candidate := fmt.Sprintf("%s%04d", prefix, seq)
		var exists int64
		if err := r.db.Model(&PurchaseOrder{}).Where("po_number = ?", candidate).Count(&exists).Error; err != nil {
			return "", err
		}
		if exists == 0 {
			return candidate, nil
		}
		seq++
	}
}

func (r *repository) Create(po *PurchaseOrder) error {
	return r.db.Create(po).Error
}

func (r *repository) FindAll(p *utils.Pagination, filter *POFilter) ([]PurchaseOrder, int64, error) {
	var pos []PurchaseOrder
	var total int64

	query := r.db.Model(&PurchaseOrder{}).
		Joins("LEFT JOIN vendors ON vendors.id = purchase_orders.vendor_id").
		Joins("LEFT JOIN users ON users.id = purchase_orders.purchasing_id")

	if filter != nil {
		if filter.Status != "" {
			query = query.Where("purchase_orders.status = ?", strings.ToUpper(strings.TrimSpace(filter.Status)))
		}
		if filter.VendorID != nil {
			query = query.Where("purchase_orders.vendor_id = ?", *filter.VendorID)
		}
		if filter.PurchasingID != nil {
			query = query.Where("purchase_orders.purchasing_id = ?", *filter.PurchasingID)
		}
	}

	if p != nil && p.Search != "" {
		pattern := "%" + strings.ToLower(p.Search) + "%"
		query = query.Where("LOWER(purchase_orders.po_number) LIKE ? OR LOWER(vendors.name) LIKE ? OR LOWER(users.name) LIKE ?",
			pattern, pattern, pattern)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if p != nil {
		sortBy := "purchase_orders.created_at"
		switch strings.ToLower(p.SortBy) {
		case "po_number", "status", "total_amount", "created_at", "updated_at":
			sortBy = "purchase_orders." + strings.ToLower(p.SortBy)
		}

		order := "DESC"
		if strings.ToUpper(p.Order) == "ASC" {
			order = "ASC"
		}

		query = query.Order(sortBy + " " + order).Offset(p.Offset()).Limit(p.Limit)
	} else {
		query = query.Order("purchase_orders.created_at DESC")
	}

	err := query.
		Preload("PR").
		Preload("Vendor").
		Preload("Purchasing").
		Preload("Lines").
		Preload("Lines.Item").
		Find(&pos).Error

	if err != nil {
		return nil, 0, err
	}

	return pos, total, nil
}

func (r *repository) FindByID(id uuid.UUID) (*PurchaseOrder, error) {
	var po PurchaseOrder
	err := r.db.
		Preload("PR").
		Preload("PR.Requester").
		Preload("PR.Requester.Division").
		Preload("PR.Approver").
		Preload("Vendor").
		Preload("Purchasing").
		Preload("Lines").
		Preload("Lines.Item").
		First(&po, "id = ?", id).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPONotFound
		}
		return nil, err
	}

	return &po, nil
}

func (r *repository) FindByPRID(prID uuid.UUID) (*PurchaseOrder, error) {
	var po PurchaseOrder
	err := r.db.First(&po, "pr_id = ?", prID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &po, nil
}

func (r *repository) FindByPONumber(poNumber string) (*PurchaseOrder, error) {
	var po PurchaseOrder
	err := r.db.
		Preload("PR").
		Preload("PR.Requester").
		Preload("PR.Requester.Division").
		Preload("PR.Approver").
		Preload("Vendor").
		Preload("Purchasing").
		Preload("Lines").
		Preload("Lines.Item").
		First(&po, "po_number = ?", poNumber).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPONotFound
		}
		return nil, err
	}

	return &po, nil
}

func (r *repository) Update(po *PurchaseOrder) error {
	return r.db.Save(po).Error
}

func (r *repository) Delete(id uuid.UUID) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("po_id = ?", id).Delete(&POLine{}).Error; err != nil {
			return err
		}
		result := tx.Delete(&PurchaseOrder{}, "id = ?", id)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return ErrPONotFound
		}
		return nil
	})
}
