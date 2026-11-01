package auth

import (
	"proceive/internal/config"
	"proceive/internal/middleware"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RegisterRoutes mendaftarkan endpoint untuk modul auth
func RegisterRoutes(rg *gin.RouterGroup, db *gorm.DB, cfg *config.Config) {
	repo := NewRepository(db)
	svc := NewService(repo, cfg)
	ctrl := NewController(svc)

	authGroup := rg.Group("/auth")
	{
		// Public route
		authGroup.POST("/login", ctrl.Login)

		// Protected routes (wajib Bearer Token)
		authProtected := authGroup.Group("")
		authProtected.Use(middleware.Auth(cfg.JWT.Secret))
		{
			authProtected.GET("/me", ctrl.GetProfile)
			authProtected.GET("/profile", ctrl.GetProfile)
			authProtected.POST("/logout", ctrl.Logout)
		}
	}
}
