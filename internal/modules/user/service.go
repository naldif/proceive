package user

import (
	"errors"
	"strings"

	"proceive/internal/modules/division"
	"proceive/internal/modules/position"
	"proceive/internal/modules/role"
	"proceive/internal/modules/vendor"
	"proceive/pkg/response"
	"proceive/pkg/utils"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrEmailAlreadyExists  = errors.New("email sudah digunakan oleh pengguna lain")
	ErrRoleNotFound        = errors.New("role yang dipilih tidak ditemukan")
	ErrDivisionNotFound    = errors.New("divisi yang dipilih tidak ditemukan")
	ErrPositionNotFound    = errors.New("jabatan yang dipilih tidak ditemukan")
	ErrVendorNotFound      = errors.New("vendor yang dipilih tidak ditemukan")
	ErrCannotDeleteAdmin   = errors.New("pengguna dengan role ADMIN tidak dapat dihapus")
	ErrCannotDeleteSelf    = errors.New("tidak dapat menghapus akun Anda sendiri yang sedang aktif")
	ErrUserHasTransactions = errors.New("pengguna tidak dapat dihapus karena memiliki riwayat transaksi (PR/PO)")
)

type Service interface {
	Create(req *CreateUserRequest) (*UserResponse, error)
	GetAll(p *utils.Pagination, filter *UserFilter) ([]UserResponse, response.PaginationMeta, error)
	GetByID(id uuid.UUID) (*UserResponse, error)
	Update(id uuid.UUID, req *UpdateUserRequest) (*UserResponse, error)
	Delete(id uuid.UUID, currentUserID uuid.UUID) error
}

type service struct {
	repo Repository
	db   *gorm.DB
}

func NewService(repo Repository, db *gorm.DB) Service {
	return &service{
		repo: repo,
		db:   db,
	}
}

// Create membuat pengguna baru dengan validasi role, divisi, jabatan, atau vendor.
func (s *service) Create(req *CreateUserRequest) (*UserResponse, error) {
	roleUUID, err := uuid.Parse(req.RoleID)
	if err != nil {
		return nil, ErrRoleNotFound
	}
	var r role.Role
	if err := s.db.First(&r, "id = ?", roleUUID).Error; err != nil {
		return nil, ErrRoleNotFound
	}

	var divUUIDPtr *uuid.UUID
	if req.DivisionID != nil && *req.DivisionID != "" {
		parsed, err := uuid.Parse(*req.DivisionID)
		if err != nil {
			return nil, ErrDivisionNotFound
		}
		var d division.Division
		if err := s.db.First(&d, "id = ?", parsed).Error; err != nil {
			return nil, ErrDivisionNotFound
		}
		divUUIDPtr = &parsed
	}

	var posUUIDPtr *uuid.UUID
	if req.PositionID != nil && *req.PositionID != "" {
		parsed, err := uuid.Parse(*req.PositionID)
		if err != nil {
			return nil, ErrPositionNotFound
		}
		var pos position.Position
		if err := s.db.First(&pos, "id = ?", parsed).Error; err != nil {
			return nil, ErrPositionNotFound
		}
		posUUIDPtr = &parsed
	}

	var vendorUUIDPtr *uuid.UUID
	var vendorObj *vendor.Vendor
	if req.VendorID != nil && *req.VendorID != "" {
		parsed, err := uuid.Parse(*req.VendorID)
		if err != nil {
			return nil, ErrVendorNotFound
		}
		var v vendor.Vendor
		if err := s.db.First(&v, "id = ?", parsed).Error; err != nil {
			return nil, ErrVendorNotFound
		}
		vendorUUIDPtr = &parsed
		vendorObj = &v
	}

	userName := strings.TrimSpace(req.Name)
	userEmail := strings.ToLower(strings.TrimSpace(req.Email))
	if vendorObj != nil {
		userName = vendorObj.Name
		userEmail = strings.ToLower(strings.TrimSpace(vendorObj.Email))
	}

	existing, err := s.repo.FindByEmail(userEmail)
	if err == nil && existing != nil {
		return nil, ErrEmailAlreadyExists
	}

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	userObj := User{
		Name:         userName,
		Email:        userEmail,
		PasswordHash: hashedPassword,
		RoleID:       roleUUID,
		DivisionID:   divUUIDPtr,
		PositionID:   posUUIDPtr,
		VendorID:     vendorUUIDPtr,
	}

	if err := s.repo.Create(&userObj); err != nil {
		return nil, err
	}

	created, err := s.repo.FindByID(userObj.ID)
	if err != nil {
		res := ToUserResponse(&userObj)
		return &res, nil
	}

	res := ToUserResponse(created)
	return &res, nil
}

// GetAll mengambil daftar pengguna dengan pagination dan filter.
func (s *service) GetAll(p *utils.Pagination, filter *UserFilter) ([]UserResponse, response.PaginationMeta, error) {
	users, total, err := s.repo.FindAll(p, filter)
	if err != nil {
		return nil, response.PaginationMeta{}, err
	}

	var meta response.PaginationMeta
	if p != nil {
		meta = p.Meta(total)
	}

	return ToUserResponseList(users), meta, nil
}

// GetByID mengambil data pengguna berdasarkan ID.
func (s *service) GetByID(id uuid.UUID) (*UserResponse, error) {
	u, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	res := ToUserResponse(u)
	return &res, nil
}

// Update memperbarui data pengguna.
func (s *service) Update(id uuid.UUID, req *UpdateUserRequest) (*UserResponse, error) {
	u, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	if req.VendorID != nil {
		if *req.VendorID != "" {
			parsed, err := uuid.Parse(*req.VendorID)
			if err != nil {
				return nil, ErrVendorNotFound
			}
			var v vendor.Vendor
			if err := s.db.First(&v, "id = ?", parsed).Error; err != nil {
				return nil, ErrVendorNotFound
			}
			u.VendorID = &parsed
			u.Name = v.Name
			u.Email = strings.ToLower(strings.TrimSpace(v.Email))
		} else {
			u.VendorID = nil
		}
	} else if u.VendorID != nil && u.Vendor != nil {
		u.Name = u.Vendor.Name
		u.Email = strings.ToLower(strings.TrimSpace(u.Vendor.Email))
	}

	if u.VendorID == nil && req.Email != "" {
		email := strings.ToLower(strings.TrimSpace(req.Email))
		if !strings.EqualFold(email, u.Email) {
			existing, err := s.repo.FindByEmail(email)
			if err == nil && existing != nil && existing.ID != id {
				return nil, ErrEmailAlreadyExists
			}
			u.Email = email
		}
	}

	if u.VendorID == nil && req.Name != "" {
		u.Name = strings.TrimSpace(req.Name)
	}

	if req.Password != "" {
		hashed, err := utils.HashPassword(req.Password)
		if err != nil {
			return nil, err
		}
		u.PasswordHash = hashed
	}

	if req.RoleID != "" {
		roleUUID, err := uuid.Parse(req.RoleID)
		if err != nil {
			return nil, ErrRoleNotFound
		}
		var r role.Role
		if err := s.db.First(&r, "id = ?", roleUUID).Error; err != nil {
			return nil, ErrRoleNotFound
		}
		u.RoleID = roleUUID
	}

	if req.DivisionID != nil {
		if *req.DivisionID != "" {
			parsed, err := uuid.Parse(*req.DivisionID)
			if err != nil {
				return nil, ErrDivisionNotFound
			}
			var d division.Division
			if err := s.db.First(&d, "id = ?", parsed).Error; err != nil {
				return nil, ErrDivisionNotFound
			}
			u.DivisionID = &parsed
		} else {
			u.DivisionID = nil
		}
	}

	if req.PositionID != nil {
		if *req.PositionID != "" {
			parsed, err := uuid.Parse(*req.PositionID)
			if err != nil {
				return nil, ErrPositionNotFound
			}
			var pos position.Position
			if err := s.db.First(&pos, "id = ?", parsed).Error; err != nil {
				return nil, ErrPositionNotFound
			}
			u.PositionID = &parsed
		} else {
			u.PositionID = nil
		}
	}

	if err := s.repo.Update(u); err != nil {
		return nil, err
	}

	updated, err := s.repo.FindByID(u.ID)
	if err != nil {
		res := ToUserResponse(u)
		return &res, nil
	}

	res := ToUserResponse(updated)
	return &res, nil
}

// Delete menghapus pengguna setelah memastikan bukan akun sendiri, bukan ADMIN, dan tidak memiliki transaksi.
func (s *service) Delete(id uuid.UUID, currentUserID uuid.UUID) error {
	if id == currentUserID {
		return ErrCannotDeleteSelf
	}

	u, err := s.repo.FindByID(id)
	if err != nil {
		return err
	}

	if u.Role != nil && strings.EqualFold(u.Role.Code, "ADMIN") {
		return ErrCannotDeleteAdmin
	}

	prCount, poCount, err := s.repo.CountTransactions(id)
	if err != nil {
		return err
	}
	if prCount > 0 || poCount > 0 {
		return ErrUserHasTransactions
	}

	return s.repo.Delete(id)
}
