package permission

import (
	"errors"

	"proceive/internal/modules/module"
	"proceive/pkg/response"
	"proceive/pkg/utils"

	"github.com/google/uuid"
)

var (
	ErrPermissionCodeAlreadyExists = errors.New("kode permission sudah digunakan")
	ErrModuleNotFound              = errors.New("modul yang dipilih tidak ditemukan")
)

type Service interface {
	Create(req *CreatePermissionRequest) (*PermissionResponse, error)
	GetAll(moduleID *uuid.UUID, p *utils.Pagination) ([]PermissionResponse, response.PaginationMeta, error)
	GetByID(id uuid.UUID) (*PermissionResponse, error)
	Update(id uuid.UUID, req *UpdatePermissionRequest) (*PermissionResponse, error)
	Delete(id uuid.UUID) error
}

type service struct {
	repo       Repository
	moduleRepo module.Repository
}

func NewService(repo Repository, moduleRepo module.Repository) Service {
	return &service{
		repo:       repo,
		moduleRepo: moduleRepo,
	}
}

func (s *service) Create(req *CreatePermissionRequest) (*PermissionResponse, error) {
	modUUID, err := uuid.Parse(req.ModuleID)
	if err != nil {
		return nil, ErrModuleNotFound
	}

	// Pastikan modul valid
	mod, err := s.moduleRepo.FindByID(modUUID)
	if err != nil {
		return nil, ErrModuleNotFound
	}

	// Cek duplikasi kode permission
	existing, err := s.repo.FindByCode(req.Code)
	if err == nil && existing != nil {
		return nil, ErrPermissionCodeAlreadyExists
	}

	p := Permission{
		ModuleID:    modUUID,
		Name:        req.Name,
		Code:        req.Code,
		Description: req.Description,
		Module:      mod,
	}

	if err := s.repo.Create(&p); err != nil {
		return nil, err
	}

	res := ToPermissionResponse(&p)
	return &res, nil
}

func (s *service) GetAll(moduleID *uuid.UUID, p *utils.Pagination) ([]PermissionResponse, response.PaginationMeta, error) {
	permissions, total, err := s.repo.FindAll(moduleID, p)
	if err != nil {
		return nil, response.PaginationMeta{}, err
	}
	var meta response.PaginationMeta
	if p != nil {
		meta = p.Meta(total)
	}
	return ToPermissionResponseList(permissions), meta, nil
}

func (s *service) GetByID(id uuid.UUID) (*PermissionResponse, error) {
	p, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	res := ToPermissionResponse(p)
	return &res, nil
}

func (s *service) Update(id uuid.UUID, req *UpdatePermissionRequest) (*PermissionResponse, error) {
	p, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	if req.ModuleID != "" {
		modUUID, err := uuid.Parse(req.ModuleID)
		if err != nil {
			return nil, ErrModuleNotFound
		}
		if modUUID != p.ModuleID {
			mod, err := s.moduleRepo.FindByID(modUUID)
			if err != nil {
				return nil, ErrModuleNotFound
			}
			p.ModuleID = modUUID
			p.Module = mod
		}
	}

	if req.Code != "" && req.Code != p.Code {
		existing, err := s.repo.FindByCode(req.Code)
		if err == nil && existing != nil && existing.ID != id {
			return nil, ErrPermissionCodeAlreadyExists
		}
		p.Code = req.Code
	}

	if req.Name != "" {
		p.Name = req.Name
	}
	if req.Description != "" {
		p.Description = req.Description
	}

	if err := s.repo.Update(p); err != nil {
		return nil, err
	}

	res := ToPermissionResponse(p)
	return &res, nil
}

func (s *service) Delete(id uuid.UUID) error {
	return s.repo.Delete(id)
}
