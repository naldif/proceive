package purchaseorder

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

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

// CreateFromPR membuat Purchase Order baru dari PR yang telah disetujui.
func (ctrl *Controller) CreateFromPR(c *gin.Context) {
	prIDStr := c.Param("purchase_requisition_id")
	prID, err := uuid.Parse(prIDStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Format ID Purchase Requisition tidak valid")
		return
	}

	var req CreatePOFromPRRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errMap := response.FormatValidationErrors(err)
		response.ValidationError(c, "Data input penerbitan PO tidak valid", errMap)
		return
	}

	currentUserID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "Sesi autentikasi tidak valid")
		return
	}

	res, err := ctrl.service.CreateFromPR(prID, &req, currentUserID)
	if err != nil {
		if errors.Is(err, ErrPRNotFound) {
			response.Error(c, http.StatusNotFound, err.Error())
			return
		}
		if errors.Is(err, ErrPRNotApproved) || errors.Is(err, ErrVendorNotFound) || errors.Is(err, ErrVendorNotActive) {
			response.Error(c, http.StatusBadRequest, err.Error())
			return
		}
		if errors.Is(err, ErrPOAlreadyExists) {
			response.Error(c, http.StatusConflict, err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, "Gagal menerbitkan Purchase Order")
		return
	}

	response.SuccessWithMessage(c, http.StatusCreated, "Purchase Order berhasil diterbitkan dan dikirim ke vendor", res)
}

// GetAll mengambil daftar Purchase Order dengan filter dan pagination.
func (ctrl *Controller) GetAll(c *gin.Context) {
	p := utils.ParsePagination(c)

	currentUserID, _ := middleware.GetCurrentUserID(c)
	userRole, _ := middleware.GetCurrentUserRole(c)

	filter := POFilter{
		Status: c.Query("status"),
	}

	vendorIDStr := c.Query("vendor_id")
	if vendorIDStr != "" {
		if vID, err := uuid.Parse(vendorIDStr); err == nil {
			filter.VendorID = &vID
		}
	}

	res, meta, err := ctrl.service.GetAll(p, &filter, currentUserID, userRole)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Gagal mengambil daftar Purchase Order")
		return
	}

	response.SuccessWithMeta(c, http.StatusOK, "Daftar Purchase Order berhasil diambil", res, meta)
}

// GetByID mengambil detail Purchase Order berdasarkan ID.
func (ctrl *Controller) GetByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Format ID Purchase Order tidak valid")
		return
	}

	currentUserID, _ := middleware.GetCurrentUserID(c)
	userRole, _ := middleware.GetCurrentUserRole(c)

	res, err := ctrl.service.GetByID(id, currentUserID, userRole)
	if err != nil {
		if errors.Is(err, ErrPONotFound) {
			response.Error(c, http.StatusNotFound, err.Error())
			return
		}
		if errors.Is(err, ErrUnauthorizedPOView) {
			response.Error(c, http.StatusForbidden, err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, "Gagal mengambil detail Purchase Order")
		return
	}

	response.SuccessWithMessage(c, http.StatusOK, "Detail Purchase Order berhasil diambil", res)
}

// UpdateStatus memperbarui status Purchase Order.
func (ctrl *Controller) UpdateStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Format ID Purchase Order tidak valid")
		return
	}

	var req UpdatePOStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errMap := response.FormatValidationErrors(err)
		response.ValidationError(c, "Data input status PO tidak valid", errMap)
		return
	}

	currentUserID, _ := middleware.GetCurrentUserID(c)
	userRole, _ := middleware.GetCurrentUserRole(c)

	res, err := ctrl.service.UpdateStatus(id, &req, currentUserID, userRole)
	if err != nil {
		if errors.Is(err, ErrPONotFound) {
			response.Error(c, http.StatusNotFound, err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, "Gagal memperbarui status Purchase Order")
		return
	}

	response.SuccessWithMessage(c, http.StatusOK, "Status Purchase Order berhasil diperbarui", res)
}

// Delete menghapus draft Purchase Order.
func (ctrl *Controller) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Format ID Purchase Order tidak valid")
		return
	}

	currentUserID, _ := middleware.GetCurrentUserID(c)
	userRole, _ := middleware.GetCurrentUserRole(c)

	if err := ctrl.service.Delete(id, currentUserID, userRole); err != nil {
		if errors.Is(err, ErrPONotFound) {
			response.Error(c, http.StatusNotFound, err.Error())
			return
		}
		if errors.Is(err, ErrPONotDraft) {
			response.Error(c, http.StatusBadRequest, err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, "Gagal menghapus Purchase Order")
		return
	}

	response.SuccessWithMessage(c, http.StatusOK, "Purchase Order berhasil dihapus", nil)
}

// GetDocument menyajikan file dokumen PDF Purchase Order.
func (ctrl *Controller) GetDocument(c *gin.Context) {
	filename := c.Param("filename")
	poNumber := strings.TrimSuffix(filename, ".pdf")

	pdfBytes, po, err := ctrl.service.GeneratePDF(poNumber)
	if err != nil {
		if errors.Is(err, ErrPONotFound) {
			response.Error(c, http.StatusNotFound, "Dokumen Purchase Order tidak ditemukan")
			return
		}
		response.Error(c, http.StatusInternalServerError, "Gagal menghasilkan dokumen Purchase Order")
		return
	}

	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", fmt.Sprintf("inline; filename=\"%s.pdf\"", po.PONumber))
	c.Data(http.StatusOK, "application/pdf", pdfBytes)
}
