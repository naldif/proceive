package purchaserequisition

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
	ErrPRNotFound = errors.New("purchase requisition tidak ditemukan")
)

type PRFilter struct {
	Status      string
	RequesterID *uuid.UUID
	DivisionID  *uuid.UUID // Untuk filter approver yang hanya boleh melihat PR dari divisinya
}

type Repository interface {
	Create(pr *PurchaseRequisition) error
	FindAll(p *utils.Pagination, filter *PRFilter) ([]PurchaseRequisition, int64, error)
	FindByID(id uuid.UUID) (*PurchaseRequisition, error)
	Update(pr *PurchaseRequisition) error
	Delete(id uuid.UUID) error
	GeneratePRNumber() (string, error)
	ReplaceLines(prID uuid.UUID, lines []PRLine) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) GeneratePRNumber() (string, error) {
	prefix := fmt.Sprintf("PR-%s-", time.Now().Format("200601"))
	var count int64
	if err := r.db.Model(&PurchaseRequisition{}).Where("pr_number LIKE ?", prefix+"%").Count(&count).Error; err != nil {
		return "", err
	}

	seq := count + 1
	for {
		candidate := fmt.Sprintf("%s%04d", prefix, seq)
		var exists int64
		if err := r.db.Model(&PurchaseRequisition{}).Where("pr_number = ?", candidate).Count(&exists).Error; err != nil {
			return "", err
		}
		if exists == 0 {
			return candidate, nil
		}
		seq++
	}
}

func (r *repository) Create(pr *PurchaseRequisition) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 1. Simpan Header PR
		if err := tx.Create(pr).Error; err != nil {
			return err
		}
		return nil
	})
}

func (r *repository) FindAll(p *utils.Pagination, filter *PRFilter) ([]PurchaseRequisition, int64, error) {
	var prs []PurchaseRequisition
	var total int64

	query := r.db.Model(&PurchaseRequisition{}).
		Joins("LEFT JOIN users ON users.id = purchase_requisitions.requester_id")

	if filter != nil {
		if filter.Status != "" {
			query = query.Where("purchase_requisitions.status = ?", strings.ToUpper(strings.TrimSpace(filter.Status)))
		}
		if filter.RequesterID != nil {
			query = query.Where("purchase_requisitions.requester_id = ?", *filter.RequesterID)
		}
		if filter.DivisionID != nil {
			// Approver hanya melihat pengajuan dari divisi yang sama
			query = query.Where("users.division_id = ?", *filter.DivisionID)
		}
	}

	if p != nil && p.Search != "" {
		pattern := "%" + strings.ToLower(p.Search) + "%"
		query = query.Where("LOWER(purchase_requisitions.pr_number) LIKE ? OR LOWER(purchase_requisitions.notes) LIKE ? OR LOWER(users.name) LIKE ?",
			pattern, pattern, pattern)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if p != nil {
		sortBy := "purchase_requisitions.created_at"
		switch strings.ToLower(p.SortBy) {
		case "pr_number", "status", "created_at", "updated_at":
			sortBy = "purchase_requisitions." + strings.ToLower(p.SortBy)
		}

		order := "DESC"
		if strings.ToUpper(p.Order) == "ASC" {
			order = "ASC"
		}

		query = query.Order(sortBy + " " + order).Offset(p.Offset()).Limit(p.Limit)
	} else {
		query = query.Order("purchase_requisitions.created_at DESC")
	}

	err := query.
		Preload("Requester").
		Preload("Requester.Division").
		Preload("Approver").
		Preload("Lines").
		Preload("Lines.Item").
		Find(&prs).Error

	if err != nil {
		return nil, 0, err
	}

	return prs, total, nil
}

func (r *repository) FindByID(id uuid.UUID) (*PurchaseRequisition, error) {
	var pr PurchaseRequisition
	err := r.db.
		Preload("Requester").
		Preload("Requester.Division").
		Preload("Requester.Position").
		Preload("Approver").
		Preload("Approver.Division").
		Preload("Approver.Position").
		Preload("Lines").
		Preload("Lines.Item").
		First(&pr, "id = ?", id).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPRNotFound
		}
		return nil, err
	}

	return &pr, nil
}

func (r *repository) Update(pr *PurchaseRequisition) error {
	return r.db.Save(pr).Error
}

func (r *repository) ReplaceLines(prID uuid.UUID, lines []PRLine) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Hapus item lines lama
		if err := tx.Where("pr_id = ?", prID).Delete(&PRLine{}).Error; err != nil {
			return err
		}

		// Masukkan item lines baru
		if len(lines) > 0 {
			for i := range lines {
				lines[i].PRID = prID
			}
			if err := tx.Create(&lines).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func (r *repository) Delete(id uuid.UUID) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("pr_id = ?", id).Delete(&PRLine{}).Error; err != nil {
			return err
		}
		result := tx.Delete(&PurchaseRequisition{}, "id = ?", id)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return ErrPRNotFound
		}
		return nil
	})
}
