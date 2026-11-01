package module

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

// Create membuat modul baru.
func (ctrl *Controller) Create(c *gin.Context) {
	var req CreateModuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errMap := response.FormatValidationErrors(err)
		response.ValidationError(c, "Data input modul tidak valid", errMap)
		return
	}

	res, err := ctrl.service.Create(&req)
	if err != nil {
		if errors.Is(err, ErrSlugAlreadyExists) {
			response.Error(c, http.StatusConflict, err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, "Gagal membuat modul")
		return
	}

	response.SuccessWithMessage(c, http.StatusCreated, "Modul berhasil dibuat", res)
}

// GetAll mengambil daftar modul dengan pagination.
func (ctrl *Controller) GetAll(c *gin.Context) {
	p := utils.ParsePagination(c)
	res, meta, err := ctrl.service.GetAll(p)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Gagal mengambil daftar modul")
		return
	}

	response.SuccessWithMeta(c, http.StatusOK, "Daftar modul berhasil diambil", res, meta)
}

// GetByID mengambil detail modul berdasarkan ID.
func (ctrl *Controller) GetByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Format ID modul tidak valid")
		return
	}

	res, err := ctrl.service.GetByID(id)
	if err != nil {
		if errors.Is(err, ErrModuleNotFound) {
			response.Error(c, http.StatusNotFound, err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, "Gagal mengambil data modul")
		return
	}

	response.SuccessWithMessage(c, http.StatusOK, "Detail modul berhasil diambil", res)
}

// Update memperbarui data modul.
func (ctrl *Controller) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Format ID modul tidak valid")
		return
	}

	var req UpdateModuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errMap := response.FormatValidationErrors(err)
		response.ValidationError(c, "Data input pembaruan modul tidak valid", errMap)
		return
	}

	res, err := ctrl.service.Update(id, &req)
	if err != nil {
		if errors.Is(err, ErrModuleNotFound) {
			response.Error(c, http.StatusNotFound, err.Error())
			return
		}
		if errors.Is(err, ErrSlugAlreadyExists) {
			response.Error(c, http.StatusConflict, err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, "Gagal memperbarui modul")
		return
	}

	response.SuccessWithMessage(c, http.StatusOK, "Modul berhasil diperbarui", res)
}

// Delete menghapus data modul.
func (ctrl *Controller) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Format ID modul tidak valid")
		return
	}

	if err := ctrl.service.Delete(id); err != nil {
		if errors.Is(err, ErrModuleNotFound) {
			response.Error(c, http.StatusNotFound, err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, "Gagal menghapus modul")
		return
	}

	response.SuccessWithMessage(c, http.StatusOK, "Modul berhasil dihapus", nil)
}
