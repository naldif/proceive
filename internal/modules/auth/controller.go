package auth

import (
	"errors"
	"net/http"

	"proceive/internal/middleware"
	"proceive/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Controller struct {
	service Service
}

func NewController(service Service) *Controller {
	return &Controller{service: service}
}

// Login memproses autentikasi user dan mengembalikan token JWT.
func (ctrl *Controller) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBind(&req); err != nil {
		errMap := response.FormatValidationErrors(err)
		response.ValidationError(c, "Format input login tidak valid", errMap)
		return
	}

	res, err := ctrl.service.Login(&req)
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			response.Error(c, http.StatusUnauthorized, "Email atau password salah")
			return
		}
		response.Error(c, http.StatusInternalServerError, "Terjadi kesalahan internal pada server")
		return
	}

	// Simpan cookie access_token agar browser / client otomatis membawa auth
	c.SetCookie("access_token", res.Token, 86400*7, "/", "", false, false)

	response.SuccessWithMessage(c, http.StatusOK, "Login berhasil", res)
}

// GetProfile mengambil profil lengkap user yang sedang login.
func (ctrl *Controller) GetProfile(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists || userID == uuid.Nil {
		response.Error(c, http.StatusUnauthorized, "Sesi tidak terotentikasi")
		return
	}

	profile, err := ctrl.service.GetProfile(userID)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			response.Error(c, http.StatusNotFound, err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, "Gagal mengambil data profil pengguna")
		return
	}

	response.SuccessWithMessage(c, http.StatusOK, "Profil pengguna berhasil diambil", profile)
}

// Logout menghapus cookie token dan mengakhiri sesi user.
func (ctrl *Controller) Logout(c *gin.Context) {
	c.SetCookie("access_token", "", -1, "/", "", false, false)
	response.SuccessWithMessage(c, http.StatusOK, "Berhasil keluar dari sesi (logout)", nil)
}
