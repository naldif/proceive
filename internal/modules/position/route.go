package position

import (
	"proceive/internal/config"
	"proceive/internal/middleware"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RegisterRoutes mendaftarkan seluruh endpoint CRUD jabatan
func RegisterRoutes(rg *gin.RouterGroup, db *gorm.DB, cfg *config.Config) {
	repo := NewRepository(db)
	svc := NewService(repo)
	ctrl := NewController(svc)

	positionGroup := rg.Group("/positions")
	positionGroup.Use(middleware.Auth(cfg.JWT.Secret))
	{
		positionGroup.GET("", ctrl.GetAll)
		positionGroup.GET("/:id", ctrl.GetByID)

		adminOnly := positionGroup.Group("")
		adminOnly.Use(middleware.RequireRole("ADMIN"))
		{
			adminOnly.POST("", ctrl.Create)
			adminOnly.PUT("/:id", ctrl.Update)
			adminOnly.DELETE("/:id", ctrl.Delete)
		}
	}
}
