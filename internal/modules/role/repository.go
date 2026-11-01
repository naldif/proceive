package role

import (
	"errors"
	"strings"

	"proceive/internal/modules/permission"
	"proceive/pkg/utils"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrRoleNotFound = errors.New("role tidak ditemukan")
)

type Repository interface {
	Create(r *Role) error
	FindAll(p *utils.Pagination) ([]Role, int64, error)
	FindByID(id uuid.UUID) (*Role, error)
	FindByCode(code string) (*Role, error)
	Update(r *Role) error
	Delete(id uuid.UUID) error
	AssignPermissions(roleID uuid.UUID, permissionIDs []uuid.UUID) error
	GetPermissionsByRoleID(roleID uuid.UUID) ([]permission.Permission, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(role *Role) error {
	return r.db.Create(role).Error
}

func (r *repository) FindAll(p *utils.Pagination) ([]Role, int64, error) {
	var roles []Role
	var total int64

	query := r.db.Model(&Role{})

	if p != nil && p.Search != "" {
		pattern := "%" + strings.ToLower(p.Search) + "%"
		query = query.Where("LOWER(name) LIKE ? OR LOWER(code) LIKE ? OR LOWER(description) LIKE ?", pattern, pattern, pattern)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

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
		query = query.Order("name ASC")
	}

	err := query.Find(&roles).Error
	return roles, total, err
}

func (r *repository) FindByID(id uuid.UUID) (*Role, error) {
	var role Role
	err := r.db.Preload("Permissions").Where("id = ?", id).First(&role).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrRoleNotFound
		}
		return nil, err
	}
	return &role, nil
}

func (r *repository) FindByCode(code string) (*Role, error) {
	var role Role
	err := r.db.Where("code = ?", code).First(&role).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrRoleNotFound
		}
		return nil, err
	}
	return &role, nil
}

func (r *repository) Update(role *Role) error {
	return r.db.Save(role).Error
}

func (r *repository) Delete(id uuid.UUID) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Hapus relasi pivot terlebih dahulu
		if err := tx.Where("role_id = ?", id).Delete(&RoleHasPermission{}).Error; err != nil {
			return err
		}

		res := tx.Delete(&Role{}, "id = ?", id)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return ErrRoleNotFound
		}
		return nil
	})
}

func (r *repository) AssignPermissions(roleID uuid.UUID, permissionIDs []uuid.UUID) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Pastikan role ada
		var role Role
		if err := tx.First(&role, "id = ?", roleID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrRoleNotFound
			}
			return err
		}

		// Hapus relasi permissions yang lama
		if err := tx.Where("role_id = ?", roleID).Delete(&RoleHasPermission{}).Error; err != nil {
			return err
		}

		// Jika array permissionIDs kosong, cukup simpan kondisi tanpa permission
		if len(permissionIDs) == 0 {
			return nil
		}

		// Insert batch relasi baru
		var pivotRows []RoleHasPermission
		for _, permID := range permissionIDs {
			pivotRows = append(pivotRows, RoleHasPermission{
				RoleID:       roleID,
				PermissionID: permID,
			})
		}

		return tx.Create(&pivotRows).Error
	})
}

func (r *repository) GetPermissionsByRoleID(roleID uuid.UUID) ([]permission.Permission, error) {
	var perms []permission.Permission
	err := r.db.Table("permissions").
		Joins("JOIN role_has_permissions ON role_has_permissions.permission_id = permissions.id").
		Where("role_has_permissions.role_id = ?", roleID).
		Order("permissions.code ASC").
		Find(&perms).Error
	return perms, err
}
