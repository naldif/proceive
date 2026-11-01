package vendor

import (
	"errors"
	"strings"

	"proceive/pkg/response"
	"proceive/pkg/utils"

	"github.com/google/uuid"
)

var (
	ErrVendorEmailAlreadyExists = errors.New("email vendor sudah digunakan")
	ErrUserEmailAlreadyExists   = errors.New("email vendor sudah terdaftar pada akun pengguna lain")
	ErrVendorRoleNotFound       = errors.New("role VENDOR belum dikonfigurasi dalam sistem")
	ErrVendorInUse              = errors.New("vendor tidak dapat dihapus karena masih terhubung dengan data transaksi pesanan pembelian (PO)")
)

type Service interface {
	Create(req *CreateVendorRequest) (*VendorResponse, error)
	GetAll(p *utils.Pagination, isActive *bool, category string) ([]VendorResponse, response.PaginationMeta, error)
	GetByID(id uuid.UUID) (*VendorResponse, error)
	Update(id uuid.UUID, req *UpdateVendorRequest) (*VendorResponse, error)
	Delete(id uuid.UUID) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

// Create membuat data rekanan vendor baru beserta akun user login.
func (s *service) Create(req *CreateVendorRequest) (*VendorResponse, error) {
	email := strings.ToLower(strings.TrimSpace(req.Email))

	existing, err := s.repo.FindByEmail(email)
	if err == nil && existing != nil {
		return nil, ErrVendorEmailAlreadyExists
	}

	usedInUsers, err := s.repo.IsEmailUsedInUsers(email, nil)
	if err != nil {
		return nil, err
	}
	if usedInUsers {
		return nil, ErrUserEmailAlreadyExists
	}

	vendorRoleID, err := s.repo.GetVendorRoleID()
	if err != nil {
		return nil, ErrVendorRoleNotFound
	}

	pwd := req.Password
	if strings.TrimSpace(pwd) == "" {
		pwd = "password123"
	}
	hashedPassword, err := utils.HashPassword(pwd)
	if err != nil {
		return nil, err
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	v := Vendor{
		Name:          strings.TrimSpace(req.Name),
		Category:      strings.TrimSpace(req.Category),
		ContactPerson: strings.TrimSpace(req.ContactPerson),
		Email:         email,
		Address:       strings.TrimSpace(req.Address),
		IsActive:      isActive,
	}

	if err := s.repo.CreateWithUser(&v, hashedPassword, vendorRoleID); err != nil {
		return nil, err
	}

	res := ToVendorResponse(&v)
	return &res, nil
}

// GetAll mengambil daftar rekanan vendor dengan pagination dan filter.
func (s *service) GetAll(p *utils.Pagination, isActive *bool, category string) ([]VendorResponse, response.PaginationMeta, error) {
	vendors, total, err := s.repo.FindAll(p, isActive, category)
	if err != nil {
		return nil, response.PaginationMeta{}, err
	}

	var meta response.PaginationMeta
	if p != nil {
		meta = p.Meta(total)
	}

	return ToVendorResponseList(vendors), meta, nil
}

// GetByID mengambil detail rekanan vendor berdasarkan ID.
func (s *service) GetByID(id uuid.UUID) (*VendorResponse, error) {
	v, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	res := ToVendorResponse(v)
	return &res, nil
}

// Update memperbarui data master vendor dan menyinkronkan data user terkait.
func (s *service) Update(id uuid.UUID, req *UpdateVendorRequest) (*VendorResponse, error) {
	v, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	if req.Email != "" {
		email := strings.ToLower(strings.TrimSpace(req.Email))
		if !strings.EqualFold(email, v.Email) {
			existing, err := s.repo.FindByEmail(email)
			if err == nil && existing != nil && existing.ID != id {
				return nil, ErrVendorEmailAlreadyExists
			}
			usedInUsers, err := s.repo.IsEmailUsedInUsers(email, &id)
			if err != nil {
				return nil, err
			}
			if usedInUsers {
				return nil, ErrUserEmailAlreadyExists
			}
			v.Email = email
		}
	}

	if req.Name != "" {
		v.Name = strings.TrimSpace(req.Name)
	}
	if req.Category != "" {
		v.Category = strings.TrimSpace(req.Category)
	}
	if req.ContactPerson != "" {
		v.ContactPerson = strings.TrimSpace(req.ContactPerson)
	}
	if req.Address != "" {
		v.Address = strings.TrimSpace(req.Address)
	}
	if req.IsActive != nil {
		v.IsActive = *req.IsActive
	}

	if err := s.repo.UpdateWithUser(v); err != nil {
		return nil, err
	}

	res := ToVendorResponse(v)
	return &res, nil
}

// Delete menghapus rekanan vendor dan akun user terkait jika tidak memiliki transaksi PO.
func (s *service) Delete(id uuid.UUID) error {
	if _, err := s.repo.FindByID(id); err != nil {
		return err
	}

	poCount, err := s.repo.CountPurchaseOrders(id)
	if err != nil {
		return err
	}
	if poCount > 0 {
		return ErrVendorInUse
	}

	return s.repo.DeleteWithUser(id)
}
