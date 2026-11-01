package purchaserequisition

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

// Create membuat pengajuan Purchase Requisition baru.
func (ctrl *Controller) Create(c *gin.Context) {
	var req CreatePRRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errMap := response.FormatValidationErrors(err)
		response.ValidationError(c, "Data input pengajuan PR tidak valid", errMap)
		return
	}

	currentUserID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "Sesi autentikasi tidak valid")
		return
	}

	res, err := ctrl.service.Create(&req, currentUserID)
	if err != nil {
		if errors.Is(err, ErrItemNotFound) {
			response.Error(c, http.StatusBadRequest, err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, "Gagal membuat pengajuan PR")
		return
	}

	response.SuccessWithMessage(c, http.StatusCreated, "Pengajuan Purchase Requisition berhasil dibuat", res)
}

// GetAll mengambil daftar Purchase Requisition dengan filter dan pagination.
func (ctrl *Controller) GetAll(c *gin.Context) {
	p := utils.ParsePagination(c)

	currentUserID, _ := middleware.GetCurrentUserID(c)
	userRole, _ := middleware.GetCurrentUserRole(c)

	filter := PRFilter{
		Status: c.Query("status"),
	}

	res, meta, err := ctrl.service.GetAll(p, &filter, currentUserID, userRole)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Gagal mengambil daftar Purchase Requisition")
		return
	}

	response.SuccessWithMeta(c, http.StatusOK, "Daftar Purchase Requisition berhasil diambil", res, meta)
}

// GetByID mengambil detail Purchase Requisition berdasarkan ID.
func (ctrl *Controller) GetByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Format ID PR tidak valid")
		return
	}

	currentUserID, _ := middleware.GetCurrentUserID(c)
	userRole, _ := middleware.GetCurrentUserRole(c)

	res, err := ctrl.service.GetByID(id, currentUserID, userRole)
	if err != nil {
		if errors.Is(err, ErrPRNotFound) {
			response.Error(c, http.StatusNotFound, err.Error())
			return
		}
		if errors.Is(err, ErrUnauthorizedPRView) {
			response.Error(c, http.StatusForbidden, err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, "Gagal mengambil detail Purchase Requisition")
		return
	}

	response.SuccessWithMessage(c, http.StatusOK, "Detail Purchase Requisition berhasil diambil", res)
}

// Update memperbarui data Purchase Requisition draft.
func (ctrl *Controller) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Format ID PR tidak valid")
		return
	}

	var req UpdatePRRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errMap := response.FormatValidationErrors(err)
		response.ValidationError(c, "Data input pembaruan PR tidak valid", errMap)
		return
	}

	currentUserID, _ := middleware.GetCurrentUserID(c)
	userRole, _ := middleware.GetCurrentUserRole(c)

	res, err := ctrl.service.Update(id, &req, currentUserID, userRole)
	if err != nil {
		if errors.Is(err, ErrPRNotFound) {
			response.Error(c, http.StatusNotFound, err.Error())
			return
		}
		if errors.Is(err, ErrPRNotDraft) || errors.Is(err, ErrItemNotFound) {
			response.Error(c, http.StatusBadRequest, err.Error())
			return
		}
		if errors.Is(err, ErrUnauthorizedPRAction) {
			response.Error(c, http.StatusForbidden, err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, "Gagal memperbarui Purchase Requisition")
		return
	}

	response.SuccessWithMessage(c, http.StatusOK, "Purchase Requisition berhasil diperbarui", res)
}

// Approve menyetujui atau menolak pengajuan Purchase Requisition.
func (ctrl *Controller) Approve(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Format ID PR tidak valid")
		return
	}

	var req ApprovePRRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errMap := response.FormatValidationErrors(err)
		response.ValidationError(c, "Data input persetujuan PR tidak valid", errMap)
		return
	}
	if req.Status == "" {
		req.Status = StatusApproved
	}

	currentUserID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "Sesi autentikasi tidak valid")
		return
	}
	userRole, _ := middleware.GetCurrentUserRole(c)

	res, err := ctrl.service.Approve(id, &req, currentUserID, userRole)
	if err != nil {
		if errors.Is(err, ErrPRNotFound) {
			response.Error(c, http.StatusNotFound, err.Error())
			return
		}
		if errors.Is(err, ErrPRNotPending) {
			response.Error(c, http.StatusBadRequest, err.Error())
			return
		}
		if errors.Is(err, ErrCannotApproveOwnPR) || errors.Is(err, ErrDifferentDivision) {
			response.Error(c, http.StatusForbidden, err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, "Gagal memproses persetujuan Purchase Requisition")
		return
	}

	msg := "Purchase Requisition berhasil disetujui"
	if res.Status == StatusRejected {
		msg = "Purchase Requisition telah ditolak"
	}

	response.SuccessWithMessage(c, http.StatusOK, msg, res)
}

// Reject menolak pengajuan Purchase Requisition secara eksplisit.
func (ctrl *Controller) Reject(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Format ID PR tidak valid")
		return
	}

	var req ApprovePRRequest
	_ = c.ShouldBindJSON(&req)
	req.Status = StatusRejected

	currentUserID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "Sesi autentikasi tidak valid")
		return
	}
	userRole, _ := middleware.GetCurrentUserRole(c)

	res, err := ctrl.service.Approve(id, &req, currentUserID, userRole)
	if err != nil {
		if errors.Is(err, ErrPRNotFound) {
			response.Error(c, http.StatusNotFound, err.Error())
			return
		}
		if errors.Is(err, ErrPRNotPending) {
			response.Error(c, http.StatusBadRequest, err.Error())
			return
		}
		if errors.Is(err, ErrCannotApproveOwnPR) || errors.Is(err, ErrDifferentDivision) {
			response.Error(c, http.StatusForbidden, err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, "Gagal memproses penolakan Purchase Requisition")
		return
	}

	response.SuccessWithMessage(c, http.StatusOK, "Purchase Requisition telah ditolak", res)
}

// Delete menghapus pengajuan Purchase Requisition draft.
func (ctrl *Controller) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Format ID PR tidak valid")
		return
	}

	currentUserID, _ := middleware.GetCurrentUserID(c)
	userRole, _ := middleware.GetCurrentUserRole(c)

	if err := ctrl.service.Delete(id, currentUserID, userRole); err != nil {
		if errors.Is(err, ErrPRNotFound) {
			response.Error(c, http.StatusNotFound, err.Error())
			return
		}
		if errors.Is(err, ErrPRNotDraft) {
			response.Error(c, http.StatusBadRequest, err.Error())
			return
		}
		if errors.Is(err, ErrUnauthorizedPRAction) {
			response.Error(c, http.StatusForbidden, err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, "Gagal menghapus Purchase Requisition")
		return
	}

	response.SuccessWithMessage(c, http.StatusOK, "Purchase Requisition berhasil dihapus", nil)
}
