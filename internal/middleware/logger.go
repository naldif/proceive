package middleware

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

// Logger mencatat setiap HTTP request yang masuk dan selesai diproses oleh server
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		rawQuery := c.Request.URL.RawQuery

		// Lanjutkan ke handler berikutnya
		c.Next()

		latency := time.Since(start)
		statusCode := c.Writer.Status()
		clientIP := c.ClientIP()
		method := c.Request.Method

		fullPath := path
		if rawQuery != "" {
			fullPath = path + "?" + rawQuery
		}

		errorMessage := c.Errors.ByType(gin.ErrorTypePrivate).String()

		if errorMessage != "" {
			log.Printf("[HTTP] %3d | %12v | %15s | %-7s %s | Error: %s",
				statusCode,
				latency,
				clientIP,
				method,
				fullPath,
				errorMessage,
			)
		} else {
			log.Printf("[HTTP] %3d | %12v | %15s | %-7s %s",
				statusCode,
				latency,
				clientIP,
				method,
				fullPath,
			)
		}
	}
}
