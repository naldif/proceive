package item

import (
	"errors"
	"strings"

	"proceive/pkg/utils"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrItemNotFound = errors.New("barang tidak ditemukan")
)

type Repository interface {
	Create(itm *Item) error
	FindAll(p *utils.Pagination) ([]Item, int64, error)
	FindByID(id uuid.UUID) (*Item, error)
	FindByCode(code string) (*Item, error)
	Update(itm *Item) error
	Delete(id uuid.UUID) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(itm *Item) error {
	return r.db.Create(itm).Error
}

func (r *repository) FindAll(p *utils.Pagination) ([]Item, int64, error) {
	var items []Item
	var total int64

	query := r.db.Model(&Item{})

	if p != nil && p.Search != "" {
		pattern := "%" + strings.ToLower(p.Search) + "%"
		query = query.Where("LOWER(code) LIKE ? OR LOWER(name) LIKE ? OR LOWER(category) LIKE ?", pattern, pattern, pattern)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if p != nil {
		sortBy := "created_at"
		switch strings.ToLower(p.SortBy) {
		case "code", "name", "category", "uom", "price_estimate", "created_at", "updated_at":
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

	if err := query.Find(&items).Error; err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

func (r *repository) FindByID(id uuid.UUID) (*Item, error) {
	var itm Item
	if err := r.db.First(&itm, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrItemNotFound
		}
		return nil, err
	}
	return &itm, nil
}

func (r *repository) FindByCode(code string) (*Item, error) {
	var itm Item
	if err := r.db.First(&itm, "LOWER(code) = LOWER(?)", code).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrItemNotFound
		}
		return nil, err
	}
	return &itm, nil
}

func (r *repository) Update(itm *Item) error {
	return r.db.Save(itm).Error
}

func (r *repository) Delete(id uuid.UUID) error {
	result := r.db.Delete(&Item{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrItemNotFound
	}
	return nil
}
