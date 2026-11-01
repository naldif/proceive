package module

import (
	"errors"

	"proceive/pkg/response"
	"proceive/pkg/utils"

	"github.com/google/uuid"
)

var (
	ErrSlugAlreadyExists = errors.New("slug modul sudah digunakan")
)

type Service interface {
	Create(req *CreateModuleRequest) (*ModuleResponse, error)
	GetAll(p *utils.Pagination) ([]ModuleResponse, response.PaginationMeta, error)
	GetByID(id uuid.UUID) (*ModuleResponse, error)
	Update(id uuid.UUID, req *UpdateModuleRequest) (*ModuleResponse, error)
	Delete(id uuid.UUID) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) Create(req *CreateModuleRequest) (*ModuleResponse, error) {
	// Cek duplikasi slug
	existing, err := s.repo.FindBySlug(req.Slug)
	if err == nil && existing != nil {
		return nil, ErrSlugAlreadyExists
	}

	m := Module{
		Name:        req.Name,
		Slug:        req.Slug,
		Description: req.Description,
	}

	if err := s.repo.Create(&m); err != nil {
		return nil, err
	}

	res := ToModuleResponse(&m)
	return &res, nil
}

func (s *service) GetAll(p *utils.Pagination) ([]ModuleResponse, response.PaginationMeta, error) {
	modules, total, err := s.repo.FindAll(p)
	if err != nil {
		return nil, response.PaginationMeta{}, err
	}
	var meta response.PaginationMeta
	if p != nil {
		meta = p.Meta(total)
	}
	return ToModuleResponseList(modules), meta, nil
}

func (s *service) GetByID(id uuid.UUID) (*ModuleResponse, error) {
	m, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	res := ToModuleResponse(m)
	return &res, nil
}

func (s *service) Update(id uuid.UUID, req *UpdateModuleRequest) (*ModuleResponse, error) {
	m, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	if req.Slug != "" && req.Slug != m.Slug {
		existing, err := s.repo.FindBySlug(req.Slug)
		if err == nil && existing != nil && existing.ID != id {
			return nil, ErrSlugAlreadyExists
		}
		m.Slug = req.Slug
	}

	if req.Name != "" {
		m.Name = req.Name
	}
	if req.Description != "" {
		m.Description = req.Description
	}

	if err := s.repo.Update(m); err != nil {
		return nil, err
	}

	res := ToModuleResponse(m)
	return &res, nil
}

func (s *service) Delete(id uuid.UUID) error {
	return s.repo.Delete(id)
}
