package purchaserequisition

import (
	"proceive/internal/config"
	"proceive/internal/middleware"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RegisterRoutes mendaftarkan seluruh endpoint modul Purchase Requisition
func RegisterRoutes(rg *gin.RouterGroup, db *gorm.DB, cfg *config.Config) {
	repo := NewRepository(db)
	svc := NewService(repo, db)
	ctrl := NewController(svc)

	prGroup := rg.Group("/purchase-requisitions")
	prGroup.Use(middleware.Auth(cfg.JWT.Secret))
	{
		// Melihat daftar & rincian PR (service memfilter visibilitas berdasarkan role & divisi pengguna)
		prGroup.GET("", ctrl.GetAll)
		prGroup.GET("/:id", ctrl.GetByID)

		// Pengajuan & pembaruan PR: dapat diajukan oleh REQUESTER, APPROVER (misal SPV), atau ADMIN
		prGroup.POST("", middleware.RequireRole("ADMIN", "REQUESTER", "APPROVER"), ctrl.Create)
		prGroup.PUT("/:id", middleware.RequireRole("ADMIN", "REQUESTER", "APPROVER"), ctrl.Update)
		prGroup.DELETE("/:id", middleware.RequireRole("ADMIN", "REQUESTER", "APPROVER"), ctrl.Delete)

		// Persetujuan & Penolakan PR: hanya dapat dilakukan oleh APPROVER atau ADMIN (validasi satu divisi ditegakkan di service)
		prGroup.PUT("/:id/approve", middleware.RequireRole("ADMIN", "APPROVER"), ctrl.Approve)
		prGroup.PUT("/:id/reject", middleware.RequireRole("ADMIN", "APPROVER"), ctrl.Reject)
	}
}
