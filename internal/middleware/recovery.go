package middleware

import (
	"log"
	"net/http"
	"runtime/debug"

	"proceive/pkg/response"

	"github.com/gin-gonic/gin"
)

// Recovery menangani unhandled panic/runtime error agar server tidak crash
// dan mengembalikan HTTP response 500 berformat JSON standar
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Cetak rincian panic beserta stack trace untuk kemudahan debugging
				log.Printf("🔥 [PANIC RECOVERED] Error: %v\nStack Trace:\n%s\n", err, string(debug.Stack()))

				// Kembalikan response JSON seragam ke client
				response.Error(c, http.StatusInternalServerError, "Terjadi kesalahan internal pada server")
				c.Abort()
			}
		}()
		c.Next()
	}
}
