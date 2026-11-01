package purchaseorder

import (
	"proceive/internal/config"
	"proceive/internal/middleware"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RegisterRoutes mendaftarkan seluruh endpoint modul Purchase Order
func RegisterRoutes(rg *gin.RouterGroup, db *gorm.DB, cfg *config.Config) {
	repo := NewRepository(db)
	svc := NewService(repo, db)
	ctrl := NewController(svc)

	// Unduh / lihat dokumen PDF resmi Purchase Order (dapat diakses via link browser & Postman)
	rg.GET("/purchase-orders/documents/:filename", ctrl.GetDocument)

	poGroup := rg.Group("/purchase-orders")
	poGroup.Use(middleware.Auth(cfg.JWT.Secret))
	{
		// Melihat daftar & detail PO (Vendor otomatis hanya melihat PO miliknya)
		poGroup.GET("", ctrl.GetAll)
		poGroup.GET("/:id", ctrl.GetByID)

		// Penerbitan PO dari PR yang sudah disetujui: hanya oleh ADMIN atau PURCHASING
		poGroup.POST("/from-purchase-requisition/:purchase_requisition_id", middleware.RequireRole("ADMIN", "PURCHASING"), ctrl.CreateFromPR)

		// Pengelolaan status & penghapusan PO
		poGroup.PUT("/:id/status", middleware.RequireRole("ADMIN", "PURCHASING"), ctrl.UpdateStatus)
		poGroup.DELETE("/:id", middleware.RequireRole("ADMIN", "PURCHASING"), ctrl.Delete)
	}
}
