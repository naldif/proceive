package role

import (
	"errors"

	"proceive/pkg/response"
	"proceive/pkg/utils"

	"github.com/google/uuid"
)

var (
	ErrRoleCodeAlreadyExists = errors.New("kode role sudah digunakan")
	ErrCannotDeleteAdminRole = errors.New("role ADMIN sistem tidak boleh dihapus")
)

type Service interface {
	Create(req *CreateRoleRequest) (*RoleDetailResponse, error)
	GetAll(p *utils.Pagination) ([]RoleResponse, response.PaginationMeta, error)
	GetByID(id uuid.UUID) (*RoleDetailResponse, error)
	Update(id uuid.UUID, req *UpdateRoleRequest) (*RoleResponse, error)
	Delete(id uuid.UUID) error
	AssignPermissions(roleID uuid.UUID, req *AssignPermissionsRequest) error
	GetRolePermissions(roleID uuid.UUID) ([]PermissionItemResponse, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) Create(req *CreateRoleRequest) (*RoleDetailResponse, error) {
	// Cek duplikasi kode role
	existing, err := s.repo.FindByCode(req.Code)
	if err == nil && existing != nil {
		return nil, ErrRoleCodeAlreadyExists
	}

	role := Role{
		Code:        req.Code,
		Name:        req.Name,
		Description: req.Description,
	}

	if err := s.repo.Create(&role); err != nil {
		return nil, err
	}

	// Jika ada permissionIDs yang disertakan saat create, langsung assign
	if len(req.PermissionIDs) > 0 {
		permUUIDs := parseUUIDStrings(req.PermissionIDs)
		if err := s.repo.AssignPermissions(role.ID, permUUIDs); err != nil {
			return nil, err
		}
	}

	// Ambil role lengkap dengan permissions
	detail, err := s.repo.FindByID(role.ID)
	if err != nil {
		return nil, err
	}

	res := ToRoleDetailResponse(detail)
	return &res, nil
}

func (s *service) GetAll(p *utils.Pagination) ([]RoleResponse, response.PaginationMeta, error) {
	roles, total, err := s.repo.FindAll(p)
	if err != nil {
		return nil, response.PaginationMeta{}, err
	}
	var meta response.PaginationMeta
	if p != nil {
		meta = p.Meta(total)
	}
	return ToRoleResponseList(roles), meta, nil
}

func (s *service) GetByID(id uuid.UUID) (*RoleDetailResponse, error) {
	role, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	res := ToRoleDetailResponse(role)
	return &res, nil
}

func (s *service) Update(id uuid.UUID, req *UpdateRoleRequest) (*RoleResponse, error) {
	role, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	if req.Code != "" && req.Code != role.Code {
		existing, err := s.repo.FindByCode(req.Code)
		if err == nil && existing != nil && existing.ID != id {
			return nil, ErrRoleCodeAlreadyExists
		}
		role.Code = req.Code
	}

	if req.Name != "" {
		role.Name = req.Name
	}
	if req.Description != "" {
		role.Description = req.Description
	}

	if err := s.repo.Update(role); err != nil {
		return nil, err
	}

	res := ToRoleResponse(role)
	return &res, nil
}

func (s *service) Delete(id uuid.UUID) error {
	role, err := s.repo.FindByID(id)
	if err != nil {
		return err
	}

	if role.Code == "ADMIN" {
		return ErrCannotDeleteAdminRole
	}

	return s.repo.Delete(id)
}

func (s *service) AssignPermissions(roleID uuid.UUID, req *AssignPermissionsRequest) error {
	permUUIDs := parseUUIDStrings(req.PermissionIDs)
	return s.repo.AssignPermissions(roleID, permUUIDs)
}

func parseUUIDStrings(strs []string) []uuid.UUID {
	res := make([]uuid.UUID, 0, len(strs))
	for _, s := range strs {
		if s != "" {
			if id, err := uuid.Parse(s); err == nil {
				res = append(res, id)
			}
		}
	}
	return res
}

func (s *service) GetRolePermissions(roleID uuid.UUID) ([]PermissionItemResponse, error) {
	// Pastikan role ada
	if _, err := s.repo.FindByID(roleID); err != nil {
		return nil, err
	}

	perms, err := s.repo.GetPermissionsByRoleID(roleID)
	if err != nil {
		return nil, err
	}

	return ToPermissionItemResponseList(perms), nil
}
