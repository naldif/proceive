package user

import (
	"errors"
	"strings"

	"proceive/pkg/utils"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrUserNotFound = errors.New("pengguna tidak ditemukan")
)

type UserFilter struct {
	RoleID     *uuid.UUID
	DivisionID *uuid.UUID
	PositionID *uuid.UUID
	VendorID   *uuid.UUID
}

type Repository interface {
	Create(u *User) error
	FindAll(p *utils.Pagination, filter *UserFilter) ([]User, int64, error)
	FindByID(id uuid.UUID) (*User, error)
	FindByEmail(email string) (*User, error)
	Update(u *User) error
	Delete(id uuid.UUID) error
	CountTransactions(id uuid.UUID) (prCount int64, poCount int64, err error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(u *User) error {
	return r.db.Create(u).Error
}

func (r *repository) FindAll(p *utils.Pagination, filter *UserFilter) ([]User, int64, error) {
	var users []User
	var total int64

	query := r.db.Model(&User{})

	if filter != nil {
		if filter.RoleID != nil {
			query = query.Where("role_id = ?", *filter.RoleID)
		}
		if filter.DivisionID != nil {
			query = query.Where("division_id = ?", *filter.DivisionID)
		}
		if filter.PositionID != nil {
			query = query.Where("position_id = ?", *filter.PositionID)
		}
		if filter.VendorID != nil {
			query = query.Where("vendor_id = ?", *filter.VendorID)
		}
	}

	if p != nil && p.Search != "" {
		pattern := "%" + strings.ToLower(p.Search) + "%"
		query = query.Where("LOWER(name) LIKE ? OR LOWER(email) LIKE ?", pattern, pattern)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if p != nil {
		sortBy := "created_at"
		switch strings.ToLower(p.SortBy) {
		case "name", "email", "created_at", "updated_at":
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

	if err := query.Preload("Role").Preload("Division").Preload("Position").Preload("Vendor").Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

func (r *repository) FindByID(id uuid.UUID) (*User, error) {
	var u User
	if err := r.db.Preload("Role").Preload("Division").Preload("Position").Preload("Vendor").First(&u, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &u, nil
}

func (r *repository) FindByEmail(email string) (*User, error) {
	var u User
	if err := r.db.Preload("Role").Preload("Division").Preload("Position").Preload("Vendor").First(&u, "LOWER(email) = LOWER(?)", email).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &u, nil
}

func (r *repository) Update(u *User) error {
	return r.db.Omit("Role", "Division", "Position", "Vendor").Save(u).Error
}

func (r *repository) Delete(id uuid.UUID) error {
	result := r.db.Delete(&User{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrUserNotFound
	}
	return nil
}

func (r *repository) CountTransactions(id uuid.UUID) (prCount int64, poCount int64, err error) {
	if err := r.db.Table("purchase_requisitions").Where("requester_id = ? OR approver_id = ?", id, id).Count(&prCount).Error; err != nil {
		return 0, 0, err
	}
	if err := r.db.Table("purchase_orders").Where("purchasing_id = ?", id).Count(&poCount).Error; err != nil {
		return 0, 0, err
	}
	return prCount, poCount, nil
}
