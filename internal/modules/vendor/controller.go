package vendor

import (
	"errors"
	"net/http"
	"strconv"

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

// Create membuat data rekanan vendor baru.
func (ctrl *Controller) Create(c *gin.Context) {
	var req CreateVendorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errMap := response.FormatValidationErrors(err)
		response.ValidationError(c, "Data input vendor tidak valid", errMap)
		return
	}

	res, err := ctrl.service.Create(&req)
	if err != nil {
		if errors.Is(err, ErrVendorEmailAlreadyExists) || errors.Is(err, ErrUserEmailAlreadyExists) {
			response.Error(c, http.StatusConflict, err.Error())
			return
		}
		if errors.Is(err, ErrVendorRoleNotFound) {
			response.Error(c, http.StatusInternalServerError, err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, "Gagal membuat vendor baru")
		return
	}

	response.SuccessWithMessage(c, http.StatusCreated, "Vendor berhasil dibuat", res)
}

// GetAll mengambil daftar rekanan vendor dengan pagination dan filter.
func (ctrl *Controller) GetAll(c *gin.Context) {
	p := utils.ParsePagination(c)

	var isActivePtr *bool
	if activeStr := c.Query("is_active"); activeStr != "" {
		if activeBool, err := strconv.ParseBool(activeStr); err == nil {
			isActivePtr = &activeBool
		}
	}

	category := c.Query("category")

	res, meta, err := ctrl.service.GetAll(p, isActivePtr, category)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Gagal mengambil daftar vendor")
		return
	}

	response.SuccessWithMeta(c, http.StatusOK, "Daftar vendor berhasil diambil", res, meta)
}

// GetByID mengambil detail rekanan vendor berdasarkan ID.
func (ctrl *Controller) GetByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Format ID vendor tidak valid")
		return
	}

	res, err := ctrl.service.GetByID(id)
	if err != nil {
		if errors.Is(err, ErrVendorNotFound) {
			response.Error(c, http.StatusNotFound, err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, "Gagal mengambil detail vendor")
		return
	}

	response.SuccessWithMessage(c, http.StatusOK, "Detail vendor berhasil diambil", res)
}

// Update memperbarui data rekanan vendor.
func (ctrl *Controller) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Format ID vendor tidak valid")
		return
	}

	var req UpdateVendorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errMap := response.FormatValidationErrors(err)
		response.ValidationError(c, "Data input pembaruan vendor tidak valid", errMap)
		return
	}

	res, err := ctrl.service.Update(id, &req)
	if err != nil {
		if errors.Is(err, ErrVendorNotFound) {
			response.Error(c, http.StatusNotFound, err.Error())
			return
		}
		if errors.Is(err, ErrVendorEmailAlreadyExists) || errors.Is(err, ErrUserEmailAlreadyExists) {
			response.Error(c, http.StatusConflict, err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, "Gagal memperbarui data vendor")
		return
	}

	response.SuccessWithMessage(c, http.StatusOK, "Vendor berhasil diperbarui", res)
}

// Delete menghapus data rekanan vendor.
func (ctrl *Controller) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Format ID vendor tidak valid")
		return
	}

	if err := ctrl.service.Delete(id); err != nil {
		if errors.Is(err, ErrVendorNotFound) {
			response.Error(c, http.StatusNotFound, err.Error())
			return
		}
		if errors.Is(err, ErrVendorInUse) {
			response.Error(c, http.StatusConflict, err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, "Gagal menghapus vendor")
		return
	}

	response.SuccessWithMessage(c, http.StatusOK, "Vendor berhasil dihapus", nil)
}
