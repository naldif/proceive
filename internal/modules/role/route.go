package role

import (
	"proceive/internal/config"
	"proceive/internal/middleware"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RegisterRoutes mendaftarkan seluruh endpoint CRUD role dan assign permissions
func RegisterRoutes(rg *gin.RouterGroup, db *gorm.DB, cfg *config.Config) {
	repo := NewRepository(db)
	svc := NewService(repo)
	ctrl := NewController(svc)

	roleGroup := rg.Group("/roles")
	roleGroup.Use(middleware.Auth(cfg.JWT.Secret))
	roleGroup.Use(middleware.RequireRole("ADMIN"))
	{
		roleGroup.POST("", ctrl.Create)
		roleGroup.GET("", ctrl.GetAll)
		roleGroup.GET("/:id", ctrl.GetByID)
		roleGroup.PUT("/:id", ctrl.Update)
		roleGroup.DELETE("/:id", ctrl.Delete)

		// Endpoint assign dan get role_has_permissions
		roleGroup.POST("/:id/permissions", ctrl.AssignPermissions)
		roleGroup.GET("/:id/permissions", ctrl.GetRolePermissions)
	}
}
