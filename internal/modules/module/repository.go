package module

import (
	"errors"
	"strings"

	"proceive/pkg/utils"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrModuleNotFound = errors.New("modul tidak ditemukan")
)

type Repository interface {
	Create(m *Module) error
	FindAll(p *utils.Pagination) ([]Module, int64, error)
	FindByID(id uuid.UUID) (*Module, error)
	FindBySlug(slug string) (*Module, error)
	Update(m *Module) error
	Delete(id uuid.UUID) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(m *Module) error {
	return r.db.Create(m).Error
}

func (r *repository) FindAll(p *utils.Pagination) ([]Module, int64, error) {
	var modules []Module
	var total int64

	query := r.db.Model(&Module{})

	if p != nil && p.Search != "" {
		pattern := "%" + strings.ToLower(p.Search) + "%"
		query = query.Where("LOWER(name) LIKE ? OR LOWER(slug) LIKE ? OR LOWER(description) LIKE ?", pattern, pattern, pattern)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if p != nil {
		// Validasi kolom pengurutan untuk keamanan
		sortBy := "created_at"
		switch strings.ToLower(p.SortBy) {
		case "name", "slug", "created_at", "updated_at":
			sortBy = strings.ToLower(p.SortBy)
		}

		order := "DESC"
		if strings.ToUpper(p.Order) == "ASC" {
			order = "ASC"
		}

		query = query.Order(sortBy + " " + order).Offset(p.Offset()).Limit(p.Limit)
	} else {
		query = query.Order("name ASC")
	}

	err := query.Find(&modules).Error
	return modules, total, err
}

func (r *repository) FindByID(id uuid.UUID) (*Module, error) {
	var m Module
	err := r.db.Where("id = ?", id).First(&m).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrModuleNotFound
		}
		return nil, err
	}
	return &m, nil
}

func (r *repository) FindBySlug(slug string) (*Module, error) {
	var m Module
	err := r.db.Where("slug = ?", slug).First(&m).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrModuleNotFound
		}
		return nil, err
	}
	return &m, nil
}

func (r *repository) Update(m *Module) error {
	return r.db.Save(m).Error
}

func (r *repository) Delete(id uuid.UUID) error {
	res := r.db.Delete(&Module{}, "id = ?", id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrModuleNotFound
	}
	return nil
}
