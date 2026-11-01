package goodsreceipt

import (
	"errors"
	"net/http"
	"os"
	"path/filepath"
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

// UploadSuratJalan memproses unggah dokumen surat jalan dari vendor untuk PO tertentu.
func (ctrl *Controller) UploadSuratJalan(c *gin.Context) {
	poID, err := uuid.Parse(c.Param("purchase_order_id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Format ID Purchase Order tidak valid")
		return
	}

	suratJalanNo := strings.TrimSpace(c.PostForm("surat_jalan_no"))
	if suratJalanNo == "" {
		suratJalanNo = strings.TrimSpace(c.Query("surat_jalan_no"))
	}
	if suratJalanNo == "" {
		response.Error(c, http.StatusBadRequest, "Nomor Surat Jalan (surat_jalan_no) wajib diisi")
		return
	}

	fileHeader, err := c.FormFile("file")
	if err != nil && c.Request.MultipartForm != nil {
		for _, files := range c.Request.MultipartForm.File {
			if len(files) > 0 {
				fileHeader = files[0]
				break
			}
		}
	}

	if fileHeader == nil {
		response.Error(c, http.StatusBadRequest, "File dokumen Surat Jalan (file) wajib diunggah")
		return
	}

	currentUserID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "Sesi autentikasi tidak valid")
		return
	}
	userRole, _ := middleware.GetCurrentUserRole(c)

	req := UploadSuratJalanRequest{
		SuratJalanNo: suratJalanNo,
	}

	res, err := ctrl.service.UploadSuratJalan(c.Request.Context(), poID, &req, fileHeader, currentUserID, userRole)
	if err != nil {
		if errors.Is(err, ErrPONotFound) {
			response.Error(c, http.StatusNotFound, err.Error())
			return
		}
		if errors.Is(err, ErrPOInvalidStatusForSJ) {
			response.Error(c, http.StatusBadRequest, err.Error())
			return
		}
		if errors.Is(err, ErrUnauthorizedVendor) {
			response.Error(c, http.StatusForbidden, err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.SuccessWithMessage(c, http.StatusCreated, "Surat Jalan berhasil diunggah", res)
}

// VerifySuratJalan memproses persetujuan atau permintaan revisi surat jalan oleh purchasing.
func (ctrl *Controller) VerifySuratJalan(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Format ID Goods Receipt tidak valid")
		return
	}

	var req VerifySuratJalanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errMap := response.FormatValidationErrors(err)
		response.ValidationError(c, "Data input verifikasi Surat Jalan tidak valid", errMap)
		return
	}

	currentUserID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "Sesi autentikasi tidak valid")
		return
	}
	userRole, _ := middleware.GetCurrentUserRole(c)

	res, err := ctrl.service.VerifySuratJalan(id, &req, currentUserID, userRole)
	if err != nil {
		if errors.Is(err, ErrGRNotFound) {
			response.Error(c, http.StatusNotFound, err.Error())
			return
		}
		if errors.Is(err, ErrGRNotUploaded) {
			response.Error(c, http.StatusBadRequest, err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, "Gagal memproses verifikasi Surat Jalan")
		return
	}

	msg := "Dokumen Surat Jalan berhasil disetujui (SJ_APPROVED)"
	if res.Status == StatusSJRevised {
		msg = "Dokumen Surat Jalan telah ditolak dan diminta revisi (SJ_REVISED)"
	}

	response.SuccessWithMessage(c, http.StatusOK, msg, res)
}

// CompleteGR memproses penyelesaian penerimaan fisik barang oleh PPIC dan menerbitkan nomor GR.
func (ctrl *Controller) CompleteGR(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Format ID Goods Receipt tidak valid")
		return
	}

	var req CompleteGRRequest
	fileHeader, _ := c.FormFile("file")

	contentType := c.Request.Header.Get("Content-Type")
	if strings.Contains(strings.ToLower(contentType), "application/json") {
		if err := c.ShouldBindJSON(&req); err != nil {
			response.ValidationError(c, "Format data request tidak valid", response.FormatValidationErrors(err))
			return
		}
	} else {
		req.Notes = strings.TrimSpace(c.PostForm("notes"))
		if req.Notes == "" {
			req.Notes = strings.TrimSpace(c.Query("notes"))
		}
		if fileHeader == nil && c.Request.MultipartForm != nil {
			for _, files := range c.Request.MultipartForm.File {
				if len(files) > 0 {
					fileHeader = files[0]
					break
				}
			}
		}
	}

	currentUserID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "Sesi autentikasi tidak valid")
		return
	}
	userRole, _ := middleware.GetCurrentUserRole(c)

	res, err := ctrl.service.CompleteGR(c.Request.Context(), id, &req, fileHeader, currentUserID, userRole)
	if err != nil {
		if errors.Is(err, ErrGRNotFound) {
			response.Error(c, http.StatusNotFound, err.Error())
			return
		}
		if errors.Is(err, ErrGRNotApproved) || errors.Is(err, ErrGRAlreadyCompleted) {
			response.Error(c, http.StatusBadRequest, err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.SuccessWithMessage(c, http.StatusOK, "Penerimaan fisik barang selesai dan nomor Goods Receipt telah diterbitkan", res)
}

// GetAll mengambil daftar seluruh Goods Receipt dengan filter dan pagination.
func (ctrl *Controller) GetAll(c *gin.Context) {
	p := utils.ParsePagination(c)

	currentUserID, _ := middleware.GetCurrentUserID(c)
	userRole, _ := middleware.GetCurrentUserRole(c)

	filter := GRFilter{
		Status: c.Query("status"),
	}

	res, meta, err := ctrl.service.GetAll(p, &filter, currentUserID, userRole)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Gagal mengambil daftar Goods Receipt")
		return
	}

	response.SuccessWithMeta(c, http.StatusOK, "Daftar Goods Receipt berhasil diambil", res, meta)
}

// GetByID mengambil detail data Goods Receipt berdasarkan ID.
func (ctrl *Controller) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Format ID Goods Receipt tidak valid")
		return
	}

	currentUserID, _ := middleware.GetCurrentUserID(c)
	userRole, _ := middleware.GetCurrentUserRole(c)

	res, err := ctrl.service.GetByID(id, currentUserID, userRole)
	if err != nil {
		if errors.Is(err, ErrGRNotFound) {
			response.Error(c, http.StatusNotFound, err.Error())
			return
		}
		if errors.Is(err, ErrUnauthorizedGRView) {
			response.Error(c, http.StatusForbidden, err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, "Gagal mengambil detail Goods Receipt")
		return
	}

	response.SuccessWithMessage(c, http.StatusOK, "Detail Goods Receipt berhasil diambil", res)
}

// ViewDocument membuka atau mengarahkan ke file dokumen Surat Jalan atau Goods Receipt.
func (ctrl *Controller) ViewDocument(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Format ID Goods Receipt tidak valid")
		return
	}

	currentUserID, _ := middleware.GetCurrentUserID(c)
	userRole, _ := middleware.GetCurrentUserRole(c)

	res, err := ctrl.service.GetByID(id, currentUserID, userRole)
	if err != nil {
		if errors.Is(err, ErrGRNotFound) {
			response.Error(c, http.StatusNotFound, err.Error())
			return
		}
		response.Error(c, http.StatusForbidden, "Tidak memiliki izin melihat dokumen Goods Receipt ini")
		return
	}

	docType := strings.ToLower(c.DefaultQuery("type", "sj"))
	docURL := res.SuratJalanURL
	subDir := "surat_jalan"
	if docType == "gr" || docType == "goods_receipt" {
		docURL = res.GRDocumentURL
		subDir = "goods_receipt"
	} else if docURL == "" && res.GRDocumentURL != "" {
		docURL = res.GRDocumentURL
		subDir = "goods_receipt"
	}

	if docURL == "" {
		response.Error(c, http.StatusNotFound, "Dokumen yang diminta belum diunggah")
		return
	}

	for strings.Contains(docURL, ".pdf.pdf") {
		docURL = strings.ReplaceAll(docURL, ".pdf.pdf", ".pdf")
	}

	filename := filepath.Base(docURL)
	localPath := filepath.Join("storage", "proceive", subDir, filename)
	if _, err := os.Stat(localPath); err == nil {
		c.File(localPath)
		return
	}
	if strings.HasPrefix(docURL, "/storage/") {
		p := strings.TrimPrefix(docURL, "/")
		if _, err := os.Stat(p); err == nil {
			c.File(p)
			return
		}
	}

	c.Redirect(http.StatusTemporaryRedirect, docURL)
}
