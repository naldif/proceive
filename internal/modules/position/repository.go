package position

import (
	"errors"
	"strings"

	"proceive/pkg/utils"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrPositionNotFound = errors.New("jabatan tidak ditemukan")
)

type Repository interface {
	Create(pos *Position) error
	FindAll(p *utils.Pagination) ([]Position, int64, error)
	FindByID(id uuid.UUID) (*Position, error)
	FindByCode(code string) (*Position, error)
	Update(pos *Position) error
	Delete(id uuid.UUID) error
	CountUsersByPositionID(id uuid.UUID) (int64, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(pos *Position) error {
	return r.db.Create(pos).Error
}

func (r *repository) FindAll(p *utils.Pagination) ([]Position, int64, error) {
	var positions []Position
	var total int64

	query := r.db.Model(&Position{})

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

	if err := query.Find(&positions).Error; err != nil {
		return nil, 0, err
	}

	return positions, total, nil
}

func (r *repository) FindByID(id uuid.UUID) (*Position, error) {
	var pos Position
	if err := r.db.First(&pos, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPositionNotFound
		}
		return nil, err
	}
	return &pos, nil
}

func (r *repository) FindByCode(code string) (*Position, error) {
	var pos Position
	if err := r.db.First(&pos, "LOWER(code) = LOWER(?)", code).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPositionNotFound
		}
		return nil, err
	}
	return &pos, nil
}

func (r *repository) Update(pos *Position) error {
	return r.db.Save(pos).Error
}

func (r *repository) Delete(id uuid.UUID) error {
	result := r.db.Delete(&Position{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrPositionNotFound
	}
	return nil
}

func (r *repository) CountUsersByPositionID(id uuid.UUID) (int64, error) {
	var count int64
	err := r.db.Table("users").Where("position_id = ?", id).Count(&count).Error
	return count, err
}
