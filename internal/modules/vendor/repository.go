package vendor

import (
	"errors"
	"strings"
	"time"

	"proceive/pkg/utils"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrVendorNotFound = errors.New("vendor tidak ditemukan")
)

type Repository interface {
	CreateWithUser(v *Vendor, userPasswordHash string, vendorRoleID uuid.UUID) error
	FindAll(p *utils.Pagination, isActive *bool, category string) ([]Vendor, int64, error)
	FindByID(id uuid.UUID) (*Vendor, error)
	FindByEmail(email string) (*Vendor, error)
	IsEmailUsedInUsers(email string, excludeVendorID *uuid.UUID) (bool, error)
	GetVendorRoleID() (uuid.UUID, error)
	UpdateWithUser(v *Vendor) error
	DeleteWithUser(id uuid.UUID) error
	CountPurchaseOrders(id uuid.UUID) (int64, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) CreateWithUser(v *Vendor, userPasswordHash string, vendorRoleID uuid.UUID) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(v).Error; err != nil {
			return err
		}

		userAccount := map[string]interface{}{
			"id":            uuid.New(),
			"name":          v.Name, // Menggunakan nama perusahaan dari vendor
			"email":         v.Email,
			"password_hash": userPasswordHash,
			"role_id":       vendorRoleID,
			"vendor_id":     v.ID,
			"created_at":    time.Now(),
			"updated_at":    time.Now(),
		}

		return tx.Table("users").Create(userAccount).Error
	})
}

func (r *repository) FindAll(p *utils.Pagination, isActive *bool, category string) ([]Vendor, int64, error) {
	var vendors []Vendor
	var total int64

	query := r.db.Model(&Vendor{})

	if isActive != nil {
		query = query.Where("is_active = ?", *isActive)
	}

	if category != "" {
		query = query.Where("LOWER(category) = LOWER(?)", strings.TrimSpace(category))
	}

	if p != nil && p.Search != "" {
		pattern := "%" + strings.ToLower(p.Search) + "%"
		query = query.Where("LOWER(name) LIKE ? OR LOWER(category) LIKE ? OR LOWER(email) LIKE ? OR LOWER(contact_person) LIKE ? OR LOWER(address) LIKE ?",
			pattern, pattern, pattern, pattern, pattern)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if p != nil {
		sortBy := "created_at"
		switch strings.ToLower(p.SortBy) {
		case "name", "category", "email", "contact_person", "is_active", "created_at", "updated_at":
			sortBy = strings.ToLower(p.SortBy)
		}

		order := "DESC"
		if strings.ToUpper(p.Order) == "ASC" {
			order = "ASC"
		}

		query = query.Order(sortBy + " " + order).Offset(p.Offset()).Limit(p.Limit)
	} else {
		query = query.Order("created_at DESC")
	}

	if err := query.Find(&vendors).Error; err != nil {
		return nil, 0, err
	}

	return vendors, total, nil
}

func (r *repository) FindByID(id uuid.UUID) (*Vendor, error) {
	var v Vendor
	if err := r.db.First(&v, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrVendorNotFound
		}
		return nil, err
	}
	return &v, nil
}

func (r *repository) FindByEmail(email string) (*Vendor, error) {
	var v Vendor
	if err := r.db.First(&v, "LOWER(email) = LOWER(?)", email).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrVendorNotFound
		}
		return nil, err
	}
	return &v, nil
}

func (r *repository) IsEmailUsedInUsers(email string, excludeVendorID *uuid.UUID) (bool, error) {
	var count int64
	query := r.db.Table("users").Where("LOWER(email) = LOWER(?)", strings.TrimSpace(email))
	if excludeVendorID != nil {
		query = query.Where("vendor_id IS NULL OR vendor_id != ?", *excludeVendorID)
	}
	if err := query.Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *repository) GetVendorRoleID() (uuid.UUID, error) {
	var res struct {
		ID uuid.UUID
	}
	if err := r.db.Table("roles").Where("code = ?", "VENDOR").Select("id").First(&res).Error; err != nil {
		return uuid.Nil, err
	}
	return res.ID, nil
}

func (r *repository) UpdateWithUser(v *Vendor) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(v).Error; err != nil {
			return err
		}

		// Sinkronisasi pembaruan nama perusahaan dan email ke akun pengguna vendor terkait
		userUpdates := map[string]interface{}{
			"name":       v.Name,
			"email":      v.Email,
			"updated_at": time.Now(),
		}

		return tx.Table("users").Where("vendor_id = ?", v.ID).Updates(userUpdates).Error
	})
}

func (r *repository) DeleteWithUser(id uuid.UUID) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 1. Hapus akun user yang terkait dengan vendor ini
		if err := tx.Table("users").Where("vendor_id = ?", id).Delete(nil).Error; err != nil {
			return err
		}

		// 2. Hapus data master vendor
		res := tx.Delete(&Vendor{}, "id = ?", id)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return ErrVendorNotFound
		}

		return nil
	})
}

func (r *repository) CountPurchaseOrders(id uuid.UUID) (int64, error) {
	var poCount int64
	if err := r.db.Table("purchase_orders").Where("vendor_id = ?", id).Count(&poCount).Error; err != nil {
		return 0, err
	}
	return poCount, nil
}
