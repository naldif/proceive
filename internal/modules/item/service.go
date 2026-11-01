package item

import (
	"errors"
	"strings"

	"proceive/pkg/response"
	"proceive/pkg/utils"

	"github.com/google/uuid"
)

var (
	ErrCodeAlreadyExists = errors.New("kode barang sudah terdaftar")
)

type Service interface {
	Create(req *CreateItemRequest) (*ItemResponse, error)
	GetAll(p *utils.Pagination) ([]ItemResponse, response.PaginationMeta, error)
	GetByID(id uuid.UUID) (*ItemResponse, error)
	Update(id uuid.UUID, req *UpdateItemRequest) (*ItemResponse, error)
	Delete(id uuid.UUID) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

// Create menambahkan item barang baru ke dalam katalog master.
func (s *service) Create(req *CreateItemRequest) (*ItemResponse, error) {
	code := strings.ToUpper(strings.TrimSpace(req.Code))

	existing, _ := s.repo.FindByCode(code)
	if existing != nil {
		return nil, ErrCodeAlreadyExists
	}

	itm := Item{
		Code:          code,
		Name:          strings.TrimSpace(req.Name),
		Category:      strings.TrimSpace(req.Category),
		UOM:           strings.ToUpper(strings.TrimSpace(req.UOM)),
		PriceEstimate: req.PriceEstimate,
	}

	if err := s.repo.Create(&itm); err != nil {
		return nil, err
	}

	res := ToItemResponse(&itm)
	return &res, nil
}

// GetAll mengambil daftar barang dengan pagination.
func (s *service) GetAll(p *utils.Pagination) ([]ItemResponse, response.PaginationMeta, error) {
	items, total, err := s.repo.FindAll(p)
	if err != nil {
		return nil, response.PaginationMeta{}, err
	}

	var meta response.PaginationMeta
	if p != nil {
		meta = p.Meta(total)
	}

	return ToItemResponseList(items), meta, nil
}

// GetByID mengambil detail item barang berdasarkan ID.
func (s *service) GetByID(id uuid.UUID) (*ItemResponse, error) {
	itm, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	res := ToItemResponse(itm)
	return &res, nil
}

// Update memperbarui data item barang.
func (s *service) Update(id uuid.UUID, req *UpdateItemRequest) (*ItemResponse, error) {
	itm, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	if req.Code != "" {
		code := strings.ToUpper(strings.TrimSpace(req.Code))
		if !strings.EqualFold(code, itm.Code) {
			existing, _ := s.repo.FindByCode(code)
			if existing != nil && existing.ID != id {
				return nil, ErrCodeAlreadyExists
			}
			itm.Code = code
		}
	}

	if req.Name != "" {
		itm.Name = strings.TrimSpace(req.Name)
	}
	if req.Category != "" {
		itm.Category = strings.TrimSpace(req.Category)
	}
	if req.UOM != "" {
		itm.UOM = strings.ToUpper(strings.TrimSpace(req.UOM))
	}
	if req.PriceEstimate != nil {
		itm.PriceEstimate = *req.PriceEstimate
	}

	if err := s.repo.Update(itm); err != nil {
		return nil, err
	}

	res := ToItemResponse(itm)
	return &res, nil
}

// Delete menghapus item barang dari katalog master.
func (s *service) Delete(id uuid.UUID) error {
	_, err := s.repo.FindByID(id)
	if err != nil {
		return err
	}

	return s.repo.Delete(id)
}
