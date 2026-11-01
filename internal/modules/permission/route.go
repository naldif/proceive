package permission

import (
	"proceive/internal/config"
	"proceive/internal/middleware"
	"proceive/internal/modules/module"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RegisterRoutes mendaftarkan seluruh endpoint CRUD permission
func RegisterRoutes(rg *gin.RouterGroup, db *gorm.DB, cfg *config.Config) {
	moduleRepo := module.NewRepository(db)
	repo := NewRepository(db)
	svc := NewService(repo, moduleRepo)
	ctrl := NewController(svc)

	permGroup := rg.Group("/permissions")
	permGroup.Use(middleware.Auth(cfg.JWT.Secret))
	permGroup.Use(middleware.RequireRole("ADMIN"))
	{
		permGroup.POST("", ctrl.Create)
		permGroup.GET("", ctrl.GetAll)
		permGroup.GET("/:id", ctrl.GetByID)
		permGroup.PUT("/:id", ctrl.Update)
		permGroup.DELETE("/:id", ctrl.Delete)
	}
}
