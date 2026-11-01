package role

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

// Create membuat role baru.
func (ctrl *Controller) Create(c *gin.Context) {
	var req CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errMap := response.FormatValidationErrors(err)
		response.ValidationError(c, "Data input role tidak valid", errMap)
		return
	}

	res, err := ctrl.service.Create(&req)
	if err != nil {
		if errors.Is(err, ErrRoleCodeAlreadyExists) {
			response.Error(c, http.StatusConflict, err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, "Gagal membuat role")
		return
	}

	response.SuccessWithMessage(c, http.StatusCreated, "Role berhasil dibuat", res)
}

// GetAll mengambil daftar role dengan pagination.
func (ctrl *Controller) GetAll(c *gin.Context) {
	p := utils.ParsePagination(c)
	res, meta, err := ctrl.service.GetAll(p)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Gagal mengambil daftar role")
		return
	}

	response.SuccessWithMeta(c, http.StatusOK, "Daftar role berhasil diambil", res, meta)
}

// GetByID mengambil detail role beserta daftar permissions.
func (ctrl *Controller) GetByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Format ID role tidak valid")
		return
	}

	res, err := ctrl.service.GetByID(id)
	if err != nil {
		if errors.Is(err, ErrRoleNotFound) {
			response.Error(c, http.StatusNotFound, err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, "Gagal mengambil data role")
		return
	}

	response.SuccessWithMessage(c, http.StatusOK, "Detail role berhasil diambil", res)
}

// Update memperbarui data role.
func (ctrl *Controller) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Format ID role tidak valid")
		return
	}

	var req UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errMap := response.FormatValidationErrors(err)
		response.ValidationError(c, "Data input pembaruan role tidak valid", errMap)
		return
	}

	res, err := ctrl.service.Update(id, &req)
	if err != nil {
		if errors.Is(err, ErrRoleNotFound) {
			response.Error(c, http.StatusNotFound, err.Error())
			return
		}
		if errors.Is(err, ErrRoleCodeAlreadyExists) {
			response.Error(c, http.StatusConflict, err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, "Gagal memperbarui role")
		return
	}

	response.SuccessWithMessage(c, http.StatusOK, "Role berhasil diperbarui", res)
}

// Delete menghapus data role.
func (ctrl *Controller) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Format ID role tidak valid")
		return
	}

	if err := ctrl.service.Delete(id); err != nil {
		if errors.Is(err, ErrRoleNotFound) {
			response.Error(c, http.StatusNotFound, err.Error())
			return
		}
		if errors.Is(err, ErrCannotDeleteAdminRole) {
			response.Error(c, http.StatusBadRequest, err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, "Gagal menghapus role")
		return
	}

	response.SuccessWithMessage(c, http.StatusOK, "Role berhasil dihapus", nil)
}

// AssignPermissions menugaskan daftar permission ke role.
func (ctrl *Controller) AssignPermissions(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Format ID role tidak valid")
		return
	}

	var req AssignPermissionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errMap := response.FormatValidationErrors(err)
		response.ValidationError(c, "Data daftar permission_ids tidak valid", errMap)
		return
	}

	if err := ctrl.service.AssignPermissions(id, &req); err != nil {
		if errors.Is(err, ErrRoleNotFound) {
			response.Error(c, http.StatusNotFound, err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, "Gagal menetapkan permissions ke role")
		return
	}

	perms, err := ctrl.service.GetRolePermissions(id)
	if err != nil {
		response.SuccessWithMessage(c, http.StatusOK, "Permissions berhasil diperbarui", nil)
		return
	}

	response.SuccessWithMessage(c, http.StatusOK, "Permissions role berhasil diperbarui", perms)
}

// GetRolePermissions mengambil daftar permission yang dimiliki role.
func (ctrl *Controller) GetRolePermissions(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Format ID role tidak valid")
		return
	}

	perms, err := ctrl.service.GetRolePermissions(id)
	if err != nil {
		if errors.Is(err, ErrRoleNotFound) {
			response.Error(c, http.StatusNotFound, err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, "Gagal mengambil permissions role")
		return
	}

	response.SuccessWithMessage(c, http.StatusOK, "Daftar permissions role berhasil diambil", perms)
}
