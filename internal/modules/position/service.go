package position

import (
	"errors"
	"strings"

	"proceive/pkg/response"
	"proceive/pkg/utils"

	"github.com/google/uuid"
)

var (
	ErrPositionCodeAlreadyExists = errors.New("kode jabatan sudah digunakan")
	ErrPositionInUse             = errors.New("jabatan tidak dapat dihapus karena masih digunakan oleh pengguna")
)

type Service interface {
	Create(req *CreatePositionRequest) (*PositionResponse, error)
	GetAll(p *utils.Pagination) ([]PositionResponse, response.PaginationMeta, error)
	GetByID(id uuid.UUID) (*PositionResponse, error)
	Update(id uuid.UUID, req *UpdatePositionRequest) (*PositionResponse, error)
	Delete(id uuid.UUID) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) Create(req *CreatePositionRequest) (*PositionResponse, error) {
	code := strings.TrimSpace(req.Code)
	existing, err := s.repo.FindByCode(code)
	if err == nil && existing != nil {
		return nil, ErrPositionCodeAlreadyExists
	}

	pos := Position{
		Code: code,
		Name: strings.TrimSpace(req.Name),
	}

	if err := s.repo.Create(&pos); err != nil {
		return nil, err
	}

	res := ToPositionResponse(&pos)
	return &res, nil
}

func (s *service) GetAll(p *utils.Pagination) ([]PositionResponse, response.PaginationMeta, error) {
	positions, total, err := s.repo.FindAll(p)
	if err != nil {
		return nil, response.PaginationMeta{}, err
	}

	var meta response.PaginationMeta
	if p != nil {
		meta = p.Meta(total)
	}

	return ToPositionResponseList(positions), meta, nil
}

func (s *service) GetByID(id uuid.UUID) (*PositionResponse, error) {
	pos, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	res := ToPositionResponse(pos)
	return &res, nil
}

func (s *service) Update(id uuid.UUID, req *UpdatePositionRequest) (*PositionResponse, error) {
	pos, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	if req.Code != "" {
		code := strings.TrimSpace(req.Code)
		if !strings.EqualFold(code, pos.Code) {
			existing, err := s.repo.FindByCode(code)
			if err == nil && existing != nil && existing.ID != id {
				return nil, ErrPositionCodeAlreadyExists
			}
			pos.Code = code
		}
	}

	if req.Name != "" {
		pos.Name = strings.TrimSpace(req.Name)
	}

	if err := s.repo.Update(pos); err != nil {
		return nil, err
	}

	res := ToPositionResponse(pos)
	return &res, nil
}

func (s *service) Delete(id uuid.UUID) error {
	// Pastikan jabatan ada
	if _, err := s.repo.FindByID(id); err != nil {
		return err
	}

	// Cek apakah ada user yang masih terhubung ke jabatan ini
	userCount, err := s.repo.CountUsersByPositionID(id)
	if err != nil {
		return err
	}
	if userCount > 0 {
		return ErrPositionInUse
	}

	return s.repo.Delete(id)
}
