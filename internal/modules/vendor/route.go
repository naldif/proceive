package vendor

import (
	"proceive/internal/config"
	"proceive/internal/middleware"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RegisterRoutes mendaftarkan seluruh endpoint CRUD vendor
func RegisterRoutes(rg *gin.RouterGroup, db *gorm.DB, cfg *config.Config) {
	repo := NewRepository(db)
	svc := NewService(repo)
	ctrl := NewController(svc)

	vendorGroup := rg.Group("/vendors")
	vendorGroup.Use(middleware.Auth(cfg.JWT.Secret))
	{
		vendorGroup.GET("", ctrl.GetAll)
		vendorGroup.GET("/:id", ctrl.GetByID)

		purchasingOrAdmin := vendorGroup.Group("")
		purchasingOrAdmin.Use(middleware.RequireRole("ADMIN", "PURCHASING"))
		{
			purchasingOrAdmin.POST("", ctrl.Create)
			purchasingOrAdmin.PUT("/:id", ctrl.Update)
			purchasingOrAdmin.DELETE("/:id", ctrl.Delete)
		}
	}
}
