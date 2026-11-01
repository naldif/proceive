package auth

import (
	"proceive/internal/modules/user"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository interface {
	FindByEmail(email string) (*user.User, error)
	FindByID(id uuid.UUID) (*user.User, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) FindByEmail(email string) (*user.User, error) {
	var u user.User
	err := r.db.
		Preload("Role").
		Preload("Division").
		Preload("Position").
		Preload("Vendor").
		Where("email = ?", email).
		First(&u).Error
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *repository) FindByID(id uuid.UUID) (*user.User, error) {
	var u user.User
	err := r.db.
		Preload("Role").
		Preload("Division").
		Preload("Position").
		Preload("Vendor").
		Where("id = ?", id).
		First(&u).Error
	if err != nil {
		return nil, err
	}
	return &u, nil
}
