package auth

import (
	"errors"

	"proceive/internal/config"
	"proceive/pkg/utils"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrInvalidCredentials = errors.New("email atau password salah")
	ErrUserNotFound       = errors.New("pengguna tidak ditemukan")
)

type Service interface {
	Login(req *LoginRequest) (*LoginResponse, error)
	GetProfile(userID uuid.UUID) (*ProfileResponse, error)
}

type service struct {
	repo Repository
	cfg  *config.Config
}

func NewService(repo Repository, cfg *config.Config) Service {
	return &service{
		repo: repo,
		cfg:  cfg,
	}
}

// Login memvalidasi kredensial pengguna dan mengembalikan token JWT.
func (s *service) Login(req *LoginRequest) (*LoginResponse, error) {
	u, err := s.repo.FindByEmail(req.Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	if !utils.CheckPasswordHash(req.Password, u.PasswordHash) {
		return nil, ErrInvalidCredentials
	}

	roleCode := ""
	if u.Role != nil {
		roleCode = u.Role.Code
	}

	token, err := utils.GenerateToken(
		u.ID,
		u.Email,
		roleCode,
		u.RoleID,
		s.cfg.JWT.Secret,
		s.cfg.JWT.ExpirationHours,
	)
	if err != nil {
		return nil, err
	}

	divisionName := ""
	if u.Division != nil {
		divisionName = u.Division.Name
	}

	positionName := ""
	if u.Position != nil {
		positionName = u.Position.Name
	}

	roleName := roleCode
	if u.Role != nil && u.Role.Name != "" {
		roleName = u.Role.Name
	}

	return &LoginResponse{
		Token: token,
		User: UserInfoResponse{
			ID:         u.ID,
			Name:       u.Name,
			Email:      u.Email,
			Role:       roleName,
			RoleID:     u.RoleID,
			DivisionID: u.DivisionID,
			Division:   divisionName,
			PositionID: u.PositionID,
			Position:   positionName,
			VendorID:   u.VendorID,
		},
	}, nil
}

// GetProfile mengambil profil lengkap pengguna berdasarkan ID.
func (s *service) GetProfile(userID uuid.UUID) (*ProfileResponse, error) {
	u, err := s.repo.FindByID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	roleCode := ""
	roleName := ""
	if u.Role != nil {
		roleCode = u.Role.Code
		roleName = u.Role.Name
	}

	divisionName := ""
	if u.Division != nil {
		divisionName = u.Division.Name
	}

	positionName := ""
	if u.Position != nil {
		positionName = u.Position.Name
	}

	vendorName := ""
	if u.Vendor != nil {
		vendorName = u.Vendor.Name
	}

	return &ProfileResponse{
		ID:         u.ID,
		Name:       u.Name,
		Email:      u.Email,
		RoleID:     u.RoleID,
		Role:       roleName,
		RoleCode:   roleCode,
		DivisionID: u.DivisionID,
		Division:   divisionName,
		PositionID: u.PositionID,
		Position:   positionName,
		VendorID:   u.VendorID,
		VendorName: vendorName,
		CreatedAt:  u.CreatedAt,
		UpdatedAt:  u.UpdatedAt,
	}, nil
}
