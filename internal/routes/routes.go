package routes

import (
	"net/http"

	"proceive/internal/config"
	"proceive/internal/modules/auth"
	"proceive/internal/modules/division"
	goodsreceipt "proceive/internal/modules/goods_receipt"
	"proceive/internal/modules/item"
	"proceive/internal/modules/module"
	"proceive/internal/modules/permission"
	"proceive/internal/modules/position"
	purchaseorder "proceive/internal/modules/purchase_order"
	purchaserequisition "proceive/internal/modules/purchase_requisition"
	"proceive/internal/modules/role"
	"proceive/internal/modules/user"
	"proceive/internal/modules/vendor"
	"proceive/pkg/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Setup menginisialisasi Gin engine dan mendaftarkan semua route dari seluruh modul.
func Setup(db *gorm.DB, cfg *config.Config) *gin.Engine {
	if cfg.App.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// Health check - tidak butuh auth
	router.GET("/health", func(c *gin.Context) {
		sqlDB, err := db.DB()
		dbStatus := "ok"
		if err != nil || sqlDB.Ping() != nil {
			dbStatus = "error"
		}
		response.SuccessWithMessage(c, http.StatusOK, "ok", gin.H{
			"app":      cfg.App.Name,
			"env":      cfg.App.Env,
			"database": dbStatus,
		})
	})

	// Rute Dokumentasi Scalar API Reference
	SetupDocsRoutes(router)

	// Melayani file upload statis (storage fallback)
	router.Static("/storage", "./storage")

	// Redirect root ke /docs
	router.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusTemporaryRedirect, "/docs")
	})

	// Semua API berada di bawah prefix /api/v1
	v1 := router.Group("/api/v1")

	// Daftarkan route tiap modul di sini.
	// Setiap modul memiliki fungsi RegisterRoutes(rg *gin.RouterGroup, db *gorm.DB, cfg *config.Config)
	registerModules(v1, db, cfg)

	// 404 handler
	router.NoRoute(func(c *gin.Context) {
		response.Error(c, http.StatusNotFound, "Endpoint tidak ditemukan")
	})

	return router
}

// registerModules mendaftarkan semua modul ke router group /api/v1.
// Tambahkan baris baru di sini setiap kali modul baru dibuat.
func registerModules(v1 *gin.RouterGroup, db *gorm.DB, cfg *config.Config) {
	// -- Auth --
	auth.RegisterRoutes(v1, db, cfg)
	//
	// -- Master Data (RBAC) --
	role.RegisterRoutes(v1, db, cfg)
	permission.RegisterRoutes(v1, db, cfg)
	module.RegisterRoutes(v1, db, cfg)
	//
	// -- Master Data (Organisasi) --
	user.RegisterRoutes(v1, db, cfg)
	division.RegisterRoutes(v1, db, cfg)
	position.RegisterRoutes(v1, db, cfg)
	//
	// -- Master Katalog --
	vendor.RegisterRoutes(v1, db, cfg)
	item.RegisterRoutes(v1, db, cfg)
	//
	// -- Transaksional --
	purchaserequisition.RegisterRoutes(v1, db, cfg)
	purchaseorder.RegisterRoutes(v1, db, cfg)
	goodsreceipt.RegisterRoutes(v1, db, cfg)
}
