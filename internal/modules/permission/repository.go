package permission

import (
	"errors"
	"strings"

	"proceive/pkg/utils"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrPermissionNotFound = errors.New("permission tidak ditemukan")
)

type Repository interface {
	Create(p *Permission) error
	FindAll(moduleID *uuid.UUID, p *utils.Pagination) ([]Permission, int64, error)
	FindByID(id uuid.UUID) (*Permission, error)
	FindByCode(code string) (*Permission, error)
	Update(p *Permission) error
	Delete(id uuid.UUID) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(p *Permission) error {
	return r.db.Create(p).Error
}

func (r *repository) FindAll(moduleID *uuid.UUID, p *utils.Pagination) ([]Permission, int64, error) {
	var permissions []Permission
	var total int64

	query := r.db.Model(&Permission{})

	if moduleID != nil && *moduleID != uuid.Nil {
		query = query.Where("module_id = ?", *moduleID)
	}

	if p != nil && p.Search != "" {
		pattern := "%" + strings.ToLower(p.Search) + "%"
		query = query.Where("LOWER(name) LIKE ? OR LOWER(code) LIKE ? OR LOWER(description) LIKE ?", pattern, pattern, pattern)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	query = query.Preload("Module")

	if p != nil {
		sortBy := "created_at"
		switch strings.ToLower(p.SortBy) {
		case "name", "code", "created_at", "updated_at":
			sortBy = strings.ToLower(p.SortBy)
		}

		order := "DESC"
		if strings.ToUpper(p.Order) == "ASC" {
			order = "ASC"
		}

		query = query.Order(sortBy + " " + order).Offset(p.Offset()).Limit(p.Limit)
	} else {
		query = query.Order("code ASC")
	}

	err := query.Find(&permissions).Error
	return permissions, total, err
}

func (r *repository) FindByID(id uuid.UUID) (*Permission, error) {
	var p Permission
	err := r.db.Preload("Module").Where("id = ?", id).First(&p).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPermissionNotFound
		}
		return nil, err
	}
	return &p, nil
}

func (r *repository) FindByCode(code string) (*Permission, error) {
	var p Permission
	err := r.db.Where("code = ?", code).First(&p).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPermissionNotFound
		}
		return nil, err
	}
	return &p, nil
}

func (r *repository) Update(p *Permission) error {
	return r.db.Save(p).Error
}

func (r *repository) Delete(id uuid.UUID) error {
	res := r.db.Delete(&Permission{}, "id = ?", id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrPermissionNotFound
	}
	return nil
}
