package user

import (
	"proceive/internal/config"
	"proceive/internal/middleware"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RegisterRoutes mendaftarkan seluruh endpoint CRUD pengguna sistem
func RegisterRoutes(rg *gin.RouterGroup, db *gorm.DB, cfg *config.Config) {
	repo := NewRepository(db)
	svc := NewService(repo, db)
	ctrl := NewController(svc)

	userGroup := rg.Group("/users")
	userGroup.Use(middleware.Auth(cfg.JWT.Secret))
	userGroup.Use(middleware.RequireRole("ADMIN"))
	{
		userGroup.POST("", ctrl.Create)
		userGroup.GET("", ctrl.GetAll)
		userGroup.GET("/:id", ctrl.GetByID)
		userGroup.PUT("/:id", ctrl.Update)
		userGroup.DELETE("/:id", ctrl.Delete)
	}
}
