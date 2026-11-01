package division

import (
	"errors"
	"strings"

	"proceive/pkg/response"
	"proceive/pkg/utils"

	"github.com/google/uuid"
)

var (
	ErrDivisionCodeAlreadyExists = errors.New("kode divisi sudah digunakan")
	ErrDivisionInUse             = errors.New("divisi tidak dapat dihapus karena masih digunakan oleh pengguna")
)

type Service interface {
	Create(req *CreateDivisionRequest) (*DivisionResponse, error)
	GetAll(p *utils.Pagination) ([]DivisionResponse, response.PaginationMeta, error)
	GetByID(id uuid.UUID) (*DivisionResponse, error)
	Update(id uuid.UUID, req *UpdateDivisionRequest) (*DivisionResponse, error)
	Delete(id uuid.UUID) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) Create(req *CreateDivisionRequest) (*DivisionResponse, error) {
	code := strings.TrimSpace(req.Code)
	existing, err := s.repo.FindByCode(code)
	if err == nil && existing != nil {
		return nil, ErrDivisionCodeAlreadyExists
	}

	div := Division{
		Code: code,
		Name: strings.TrimSpace(req.Name),
	}

	if err := s.repo.Create(&div); err != nil {
		return nil, err
	}

	res := ToDivisionResponse(&div)
	return &res, nil
}

func (s *service) GetAll(p *utils.Pagination) ([]DivisionResponse, response.PaginationMeta, error) {
	divisions, total, err := s.repo.FindAll(p)
	if err != nil {
		return nil, response.PaginationMeta{}, err
	}

	var meta response.PaginationMeta
	if p != nil {
		meta = p.Meta(total)
	}

	return ToDivisionResponseList(divisions), meta, nil
}

func (s *service) GetByID(id uuid.UUID) (*DivisionResponse, error) {
	div, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	res := ToDivisionResponse(div)
	return &res, nil
}

func (s *service) Update(id uuid.UUID, req *UpdateDivisionRequest) (*DivisionResponse, error) {
	div, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	if req.Code != "" {
		code := strings.TrimSpace(req.Code)
		if !strings.EqualFold(code, div.Code) {
			existing, err := s.repo.FindByCode(code)
			if err == nil && existing != nil && existing.ID != id {
				return nil, ErrDivisionCodeAlreadyExists
			}
			div.Code = code
		}
	}

	if req.Name != "" {
		div.Name = strings.TrimSpace(req.Name)
	}

	if err := s.repo.Update(div); err != nil {
		return nil, err
	}

	res := ToDivisionResponse(div)
	return &res, nil
}

func (s *service) Delete(id uuid.UUID) error {
	// Pastikan divisi ada
	if _, err := s.repo.FindByID(id); err != nil {
		return err
	}

	// Cek apakah ada user yang masih terhubung ke divisi ini
	userCount, err := s.repo.CountUsersByDivisionID(id)
	if err != nil {
		return err
	}
	if userCount > 0 {
		return ErrDivisionInUse
	}

	return s.repo.Delete(id)
}
