package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// SetupDocsRoutes mendaftarkan endpoint untuk dokumentasi Scalar API Reference
func SetupDocsRoutes(router *gin.Engine) {
	// Endpoint untuk melayani file openapi.json
	router.GET("/docs/openapi.json", func(c *gin.Context) {
		c.File("docs/openapi.json")
	})

	// Endpoint dokumentasi interaktif Scalar
	router.GET("/docs", func(c *gin.Context) {
		c.File("docs/scalar.html")
	})

	// Alias /scalar ke /docs
	router.GET("/scalar", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/docs")
	})
}
