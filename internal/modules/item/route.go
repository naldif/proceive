package item

import (
	"proceive/internal/config"
	"proceive/internal/middleware"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RegisterRoutes mendaftarkan endpoint untuk modul Item
func RegisterRoutes(rg *gin.RouterGroup, db *gorm.DB, cfg *config.Config) {
	repo := NewRepository(db)
	svc := NewService(repo)
	ctrl := NewController(svc)

	itemsGroup := rg.Group("/items")
	itemsGroup.Use(middleware.Auth(cfg.JWT.Secret))
	{
		// Semua user terautentikasi dapat melihat katalog barang untuk membuat PR / PO
		itemsGroup.GET("", ctrl.GetAll)
		itemsGroup.GET("/:id", ctrl.GetByID)

		// Pengelolaan katalog barang hanya boleh oleh ADMIN dan PURCHASING
		itemsGroup.POST("", middleware.RequireRole("ADMIN", "PURCHASING"), ctrl.Create)
		itemsGroup.PUT("/:id", middleware.RequireRole("ADMIN", "PURCHASING"), ctrl.Update)
		itemsGroup.DELETE("/:id", middleware.RequireRole("ADMIN", "PURCHASING"), ctrl.Delete)
	}
}
