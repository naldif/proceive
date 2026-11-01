package division

import (
	"proceive/internal/config"
	"proceive/internal/middleware"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RegisterRoutes mendaftarkan seluruh endpoint CRUD divisi
func RegisterRoutes(rg *gin.RouterGroup, db *gorm.DB, cfg *config.Config) {
	repo := NewRepository(db)
	svc := NewService(repo)
	ctrl := NewController(svc)

	divisionGroup := rg.Group("/divisions")
	divisionGroup.Use(middleware.Auth(cfg.JWT.Secret))
	{
		divisionGroup.GET("", ctrl.GetAll)
		divisionGroup.GET("/:id", ctrl.GetByID)

		adminOnly := divisionGroup.Group("")
		adminOnly.Use(middleware.RequireRole("ADMIN"))
		{
			adminOnly.POST("", ctrl.Create)
			adminOnly.PUT("/:id", ctrl.Update)
			adminOnly.DELETE("/:id", ctrl.Delete)
		}
	}
}
