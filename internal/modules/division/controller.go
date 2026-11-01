package division

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

// Create membuat divisi baru.
func (ctrl *Controller) Create(c *gin.Context) {
	var req CreateDivisionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errMap := response.FormatValidationErrors(err)
		response.ValidationError(c, "Data input divisi tidak valid", errMap)
		return
	}

	res, err := ctrl.service.Create(&req)
	if err != nil {
		if errors.Is(err, ErrDivisionCodeAlreadyExists) {
			response.Error(c, http.StatusConflict, err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, "Gagal membuat divisi baru")
		return
	}

	response.SuccessWithMessage(c, http.StatusCreated, "Divisi berhasil dibuat", res)
}

// GetAll mengambil daftar divisi dengan pagination.
func (ctrl *Controller) GetAll(c *gin.Context) {
	p := utils.ParsePagination(c)
	res, meta, err := ctrl.service.GetAll(p)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Gagal mengambil daftar divisi")
		return
	}

	response.SuccessWithMeta(c, http.StatusOK, "Daftar divisi berhasil diambil", res, meta)
}

// GetByID mengambil detail divisi berdasarkan ID.
func (ctrl *Controller) GetByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Format ID divisi tidak valid")
		return
	}

	res, err := ctrl.service.GetByID(id)
	if err != nil {
		if errors.Is(err, ErrDivisionNotFound) {
			response.Error(c, http.StatusNotFound, err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, "Gagal mengambil detail divisi")
		return
	}

	response.SuccessWithMessage(c, http.StatusOK, "Detail divisi berhasil diambil", res)
}

// Update memperbarui data divisi.
func (ctrl *Controller) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Format ID divisi tidak valid")
		return
	}

	var req UpdateDivisionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errMap := response.FormatValidationErrors(err)
		response.ValidationError(c, "Data input pembaruan divisi tidak valid", errMap)
		return
	}

	res, err := ctrl.service.Update(id, &req)
	if err != nil {
		if errors.Is(err, ErrDivisionNotFound) {
			response.Error(c, http.StatusNotFound, err.Error())
			return
		}
		if errors.Is(err, ErrDivisionCodeAlreadyExists) {
			response.Error(c, http.StatusConflict, err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, "Gagal memperbarui divisi")
		return
	}

	response.SuccessWithMessage(c, http.StatusOK, "Divisi berhasil diperbarui", res)
}

// Delete menghapus data divisi.
func (ctrl *Controller) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Format ID divisi tidak valid")
		return
	}

	if err := ctrl.service.Delete(id); err != nil {
		if errors.Is(err, ErrDivisionNotFound) {
			response.Error(c, http.StatusNotFound, err.Error())
			return
		}
		if errors.Is(err, ErrDivisionInUse) {
			response.Error(c, http.StatusConflict, err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, "Gagal menghapus divisi")
		return
	}

	response.SuccessWithMessage(c, http.StatusOK, "Divisi berhasil dihapus", nil)
}
