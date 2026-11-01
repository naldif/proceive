package goodsreceipt

import (
	"proceive/internal/config"
	"proceive/internal/middleware"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RegisterRoutes mendaftarkan seluruh endpoint modul Goods Receipt
func RegisterRoutes(rg *gin.RouterGroup, db *gorm.DB, cfg *config.Config) {
	repo := NewRepository(db)
	svc := NewService(repo, db, cfg)
	ctrl := NewController(svc)

	grGroup := rg.Group("/goods-receipts")
	grGroup.Use(middleware.Auth(cfg.JWT.Secret))
	{
		// Melihat daftar & detail Goods Receipt
		grGroup.GET("", ctrl.GetAll)
		grGroup.GET("/:id", ctrl.GetByID)
		grGroup.GET("/:id/document", ctrl.ViewDocument)

		// 1. Vendor: Upload dokumen fisik Surat Jalan
		grGroup.POST("/upload-surat-jalan/:purchase_order_id", middleware.RequireRole("ADMIN", "VENDOR"), ctrl.UploadSuratJalan)

		// 2. Purchasing: Verifikasi kesesuaian dokumen Surat Jalan (SJ_APPROVED / SJ_REVISED)
		grGroup.PUT("/:id/verify-surat-jalan", middleware.RequireRole("ADMIN", "PURCHASING"), ctrl.VerifySuratJalan)

		// 3. PPIC / Warehouse: Selesaikan penerimaan fisik barang dan terbitkan nomor Goods Receipt
		grGroup.PUT("/:id/complete", middleware.RequireRole("ADMIN", "PPIC"), ctrl.CompleteGR)
	}
}
