package permission

import (
	"errors"
	"net/http"

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

// Create membuat permission baru.
func (ctrl *Controller) Create(c *gin.Context) {
	var req CreatePermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errMap := response.FormatValidationErrors(err)
		response.ValidationError(c, "Data input permission tidak valid", errMap)
		return
	}

	res, err := ctrl.service.Create(&req)
	if err != nil {
		if errors.Is(err, ErrModuleNotFound) {
			response.Error(c, http.StatusBadRequest, err.Error())
			return
		}
		if errors.Is(err, ErrPermissionCodeAlreadyExists) {
			response.Error(c, http.StatusConflict, err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, "Gagal membuat permission")
		return
	}

	response.SuccessWithMessage(c, http.StatusCreated, "Permission berhasil dibuat", res)
}

// GetAll mengambil daftar permission dengan filter dan pagination.
func (ctrl *Controller) GetAll(c *gin.Context) {
	p := utils.ParsePagination(c)
	var moduleIDPtr *uuid.UUID
	modIDStr := c.Query("module_id")
	if modIDStr != "" {
		modID, err := uuid.Parse(modIDStr)
		if err == nil {
			moduleIDPtr = &modID
		}
	}

	res, meta, err := ctrl.service.GetAll(moduleIDPtr, p)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Gagal mengambil daftar permission")
		return
	}

	response.SuccessWithMeta(c, http.StatusOK, "Daftar permission berhasil diambil", res, meta)
}

// GetByID mengambil detail permission berdasarkan ID.
func (ctrl *Controller) GetByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Format ID permission tidak valid")
		return
	}

	res, err := ctrl.service.GetByID(id)
	if err != nil {
		if errors.Is(err, ErrPermissionNotFound) {
			response.Error(c, http.StatusNotFound, err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, "Gagal mengambil data permission")
		return
	}

	response.SuccessWithMessage(c, http.StatusOK, "Detail permission berhasil diambil", res)
}

// Update memperbarui data permission.
func (ctrl *Controller) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Format ID permission tidak valid")
		return
	}

	var req UpdatePermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errMap := response.FormatValidationErrors(err)
		response.ValidationError(c, "Data input pembaruan permission tidak valid", errMap)
		return
	}

	res, err := ctrl.service.Update(id, &req)
	if err != nil {
		if errors.Is(err, ErrPermissionNotFound) {
			response.Error(c, http.StatusNotFound, err.Error())
			return
		}
		if errors.Is(err, ErrModuleNotFound) {
			response.Error(c, http.StatusBadRequest, err.Error())
			return
		}
		if errors.Is(err, ErrPermissionCodeAlreadyExists) {
			response.Error(c, http.StatusConflict, err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, "Gagal memperbarui permission")
		return
	}

	response.SuccessWithMessage(c, http.StatusOK, "Permission berhasil diperbarui", res)
}

// Delete menghapus permission berdasarkan ID.
func (ctrl *Controller) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Format ID permission tidak valid")
		return
	}

	if err := ctrl.service.Delete(id); err != nil {
		if errors.Is(err, ErrPermissionNotFound) {
			response.Error(c, http.StatusNotFound, err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, "Gagal menghapus permission")
		return
	}

	response.SuccessWithMessage(c, http.StatusOK, "Permission berhasil dihapus", nil)
}
