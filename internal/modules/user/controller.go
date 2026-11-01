package user

import (
	"errors"
	"net/http"

	"proceive/internal/middleware"
	"proceive/pkg/response"
	"proceive/pkg/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Controller struct {
	service Service
}

func NewController(service Service) *Controller {
	return &Controller{service: service}
}

// Create membuat akun user baru.
func (ctrl *Controller) Create(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errMap := response.FormatValidationErrors(err)
		response.ValidationError(c, "Data input pengguna tidak valid", errMap)
		return
	}

	res, err := ctrl.service.Create(&req)
	if err != nil {
		if errors.Is(err, ErrEmailAlreadyExists) {
			response.Error(c, http.StatusConflict, err.Error())
			return
		}
		if errors.Is(err, ErrRoleNotFound) || errors.Is(err, ErrDivisionNotFound) ||
			errors.Is(err, ErrPositionNotFound) || errors.Is(err, ErrVendorNotFound) {
			response.Error(c, http.StatusBadRequest, err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, "Gagal membuat pengguna baru")
		return
	}

	response.SuccessWithMessage(c, http.StatusCreated, "Pengguna berhasil dibuat", res)
}

// GetAll mengambil daftar user dengan pagination dan filter.
func (ctrl *Controller) GetAll(c *gin.Context) {
	p := utils.ParsePagination(c)

	filter := UserFilter{}
	if rIDStr := c.Query("role_id"); rIDStr != "" {
		if rID, err := uuid.Parse(rIDStr); err == nil {
			filter.RoleID = &rID
		}
	}
	if dIDStr := c.Query("division_id"); dIDStr != "" {
		if dID, err := uuid.Parse(dIDStr); err == nil {
			filter.DivisionID = &dID
		}
	}
	if pIDStr := c.Query("position_id"); pIDStr != "" {
		if pID, err := uuid.Parse(pIDStr); err == nil {
			filter.PositionID = &pID
		}
	}
	if vIDStr := c.Query("vendor_id"); vIDStr != "" {
		if vID, err := uuid.Parse(vIDStr); err == nil {
			filter.VendorID = &vID
		}
	}

	res, meta, err := ctrl.service.GetAll(p, &filter)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Gagal mengambil daftar pengguna")
		return
	}

	response.SuccessWithMeta(c, http.StatusOK, "Daftar pengguna berhasil diambil", res, meta)
}

// GetByID mengambil detail user berdasarkan ID.
func (ctrl *Controller) GetByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Format ID pengguna tidak valid")
		return
	}

	res, err := ctrl.service.GetByID(id)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			response.Error(c, http.StatusNotFound, err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, "Gagal mengambil detail pengguna")
		return
	}

	response.SuccessWithMessage(c, http.StatusOK, "Detail pengguna berhasil diambil", res)
}

// Update memperbarui data user.
func (ctrl *Controller) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Format ID pengguna tidak valid")
		return
	}

	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errMap := response.FormatValidationErrors(err)
		response.ValidationError(c, "Data input pembaruan pengguna tidak valid", errMap)
		return
	}

	res, err := ctrl.service.Update(id, &req)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			response.Error(c, http.StatusNotFound, err.Error())
			return
		}
		if errors.Is(err, ErrEmailAlreadyExists) {
			response.Error(c, http.StatusConflict, err.Error())
			return
		}
		if errors.Is(err, ErrRoleNotFound) || errors.Is(err, ErrDivisionNotFound) ||
			errors.Is(err, ErrPositionNotFound) || errors.Is(err, ErrVendorNotFound) {
			response.Error(c, http.StatusBadRequest, err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, "Gagal memperbarui pengguna")
		return
	}

	response.SuccessWithMessage(c, http.StatusOK, "Pengguna berhasil diperbarui", res)
}

// Delete menghapus data user.
func (ctrl *Controller) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Format ID pengguna tidak valid")
		return
	}

	currentUserID, _ := middleware.GetCurrentUserID(c)

	if err := ctrl.service.Delete(id, currentUserID); err != nil {
		if errors.Is(err, ErrUserNotFound) {
			response.Error(c, http.StatusNotFound, err.Error())
			return
		}
		if errors.Is(err, ErrCannotDeleteSelf) || errors.Is(err, ErrCannotDeleteAdmin) || errors.Is(err, ErrUserHasTransactions) {
			response.Error(c, http.StatusForbidden, err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, "Gagal menghapus pengguna")
		return
	}

	response.SuccessWithMessage(c, http.StatusOK, "Pengguna berhasil dihapus", nil)
}
