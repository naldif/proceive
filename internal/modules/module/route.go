package module

import (
	"proceive/internal/config"
	"proceive/internal/middleware"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RegisterRoutes mendaftarkan seluruh endpoint CRUD modul
func RegisterRoutes(rg *gin.RouterGroup, db *gorm.DB, cfg *config.Config) {
	repo := NewRepository(db)
	svc := NewService(repo)
	ctrl := NewController(svc)

	moduleGroup := rg.Group("/modules")
	moduleGroup.Use(middleware.Auth(cfg.JWT.Secret))
	moduleGroup.Use(middleware.RequireRole("ADMIN"))
	{
		moduleGroup.POST("", ctrl.Create)
		moduleGroup.GET("", ctrl.GetAll)
		moduleGroup.GET("/:id", ctrl.GetByID)
		moduleGroup.PUT("/:id", ctrl.Update)
		moduleGroup.DELETE("/:id", ctrl.Delete)
	}
}
