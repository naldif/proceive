package middleware

import (
	"net/http"
	"strings"

	"proceive/pkg/response"
	"proceive/pkg/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	CtxUserIDKey   = "user_id"
	CtxEmailKey    = "email"
	CtxRoleIDKey   = "role_id"
	CtxRoleCodeKey = "role_code"
	CtxClaimsKey   = "claims"
)

// Auth memvalidasi token JWT dari header Authorization atau cookie access_token.
func Auth(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var tokenString string

		authHeader := strings.TrimSpace(c.GetHeader("Authorization"))
		if authHeader != "" {
			parts := strings.SplitN(authHeader, " ", 2)
			if strings.EqualFold(parts[0], "Bearer") && len(parts) == 2 {
				tokenString = strings.TrimSpace(parts[1])
			}
		}

		// Fallback ke cookie jika tidak ada di header
		if tokenString == "" {
			if cookieVal, err := c.Cookie("access_token"); err == nil {
				tokenString = strings.TrimSpace(cookieVal)
			}
		}

		if tokenString == "" {
			response.Error(c, http.StatusUnauthorized, "Token autentikasi tidak ditemukan")
			c.Abort()
			return
		}
		claims, err := utils.ValidateToken(tokenString, secret)
		if err != nil {
			response.Error(c, http.StatusUnauthorized, "Token tidak valid atau telah kedaluwarsa")
			c.Abort()
			return
		}

		// Simpan informasi pengguna yang terautentikasi ke Gin Context
		c.Set(CtxUserIDKey, claims.UserID)
		c.Set(CtxEmailKey, claims.Email)
		c.Set(CtxRoleIDKey, claims.RoleID)
		c.Set(CtxRoleCodeKey, claims.RoleCode)
		c.Set(CtxClaimsKey, claims)

		c.Next()
	}
}

// GetCurrentUserID mengambil ID (UUID) pengguna yang sedang login dari context Gin
func GetCurrentUserID(c *gin.Context) (uuid.UUID, bool) {
	val, exists := c.Get(CtxUserIDKey)
	if !exists {
		return uuid.Nil, false
	}
	id, ok := val.(uuid.UUID)
	return id, ok
}

// GetCurrentUserRole mengambil kode role pengguna (misal ADMIN, JASTIPER, BUYER) dari context Gin
func GetCurrentUserRole(c *gin.Context) (string, bool) {
	val, exists := c.Get(CtxRoleCodeKey)
	if !exists {
		return "", false
	}
	role, ok := val.(string)
	return role, ok
}

// GetCurrentUserRoleID mengambil ID (UUID) role pengguna dari context Gin
func GetCurrentUserRoleID(c *gin.Context) (uuid.UUID, bool) {
	val, exists := c.Get(CtxRoleIDKey)
	if !exists {
		return uuid.Nil, false
	}
	id, ok := val.(uuid.UUID)
	return id, ok
}

// GetCurrentClaims mengambil claims lengkap JWT dari context Gin
func GetCurrentClaims(c *gin.Context) (*utils.JWTClaims, bool) {
	val, exists := c.Get(CtxClaimsKey)
	if !exists {
		return nil, false
	}
	claims, ok := val.(*utils.JWTClaims)
	return claims, ok
}
