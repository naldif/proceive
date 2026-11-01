package division

import (
	"errors"
	"strings"

	"proceive/pkg/utils"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrDivisionNotFound = errors.New("divisi tidak ditemukan")
)

type Repository interface {
	Create(div *Division) error
	FindAll(p *utils.Pagination) ([]Division, int64, error)
	FindByID(id uuid.UUID) (*Division, error)
	FindByCode(code string) (*Division, error)
	Update(div *Division) error
	Delete(id uuid.UUID) error
	CountUsersByDivisionID(id uuid.UUID) (int64, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(div *Division) error {
	return r.db.Create(div).Error
}

func (r *repository) FindAll(p *utils.Pagination) ([]Division, int64, error) {
	var divisions []Division
	var total int64

	query := r.db.Model(&Division{})

	if p != nil && p.Search != "" {
		pattern := "%" + strings.ToLower(p.Search) + "%"
		query = query.Where("LOWER(code) LIKE ? OR LOWER(name) LIKE ?", pattern, pattern)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if p != nil {
		sortBy := "created_at"
		switch strings.ToLower(p.SortBy) {
		case "code", "name", "created_at", "updated_at":
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

	if err := query.Find(&divisions).Error; err != nil {
		return nil, 0, err
	}

	return divisions, total, nil
}

func (r *repository) FindByID(id uuid.UUID) (*Division, error) {
	var div Division
	if err := r.db.First(&div, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrDivisionNotFound
		}
		return nil, err
	}
	return &div, nil
}

func (r *repository) FindByCode(code string) (*Division, error) {
	var div Division
	if err := r.db.First(&div, "LOWER(code) = LOWER(?)", code).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrDivisionNotFound
		}
		return nil, err
	}
	return &div, nil
}

func (r *repository) Update(div *Division) error {
	return r.db.Save(div).Error
}

func (r *repository) Delete(id uuid.UUID) error {
	result := r.db.Delete(&Division{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrDivisionNotFound
	}
	return nil
}

func (r *repository) CountUsersByDivisionID(id uuid.UUID) (int64, error) {
	var count int64
	err := r.db.Table("users").Where("division_id = ?", id).Count(&count).Error
	return count, err
}
