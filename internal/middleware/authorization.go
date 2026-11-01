package middleware

import (
	"net/http"
	"strings"

	"proceive/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// isAdmin memeriksa apakah role pengguna adalah ADMIN
func isAdmin(roleCode string) bool {
	return strings.EqualFold(strings.TrimSpace(roleCode), "ADMIN")
}

// RequireRole membatasi akses endpoint hanya untuk role tertentu.
// Jika role pengguna adalah ADMIN, otorisasi akan otomatis di-BYPASS (diberi akses penuh).
func RequireRole(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		roleCode, exists := GetCurrentUserRole(c)
		if !exists || roleCode == "" {
			response.Error(c, http.StatusUnauthorized, "Sesi tidak terotentikasi")
			c.Abort()
			return
		}

		// Role ADMIN memiliki hak akses penuh
		if isAdmin(roleCode) {
			c.Next()
			return
		}

		// Periksa apakah role pengguna cocok dengan salah satu allowedRoles
		for _, allowed := range allowedRoles {
			if strings.EqualFold(roleCode, allowed) {
				c.Next()
				return
			}
		}

		response.Error(c, http.StatusForbidden, "Akses ditolak: peran Anda tidak memiliki izin untuk mengakses resource ini")
		c.Abort()
	}
}

// RequirePermission membatasi akses endpoint berdasarkan kode permission tertentu di database.
// Jika role pengguna adalah ADMIN, otorisasi akan otomatis di-BYPASS (diberi akses penuh).
func RequirePermission(db *gorm.DB, permissionCode string) gin.HandlerFunc {
	return func(c *gin.Context) {
		roleCode, existsRole := GetCurrentUserRole(c)
		roleID, existsID := GetCurrentUserRoleID(c)

		if !existsRole || !existsID || roleID == uuid.Nil {
			response.Error(c, http.StatusUnauthorized, "Sesi tidak terotentikasi")
			c.Abort()
			return
		}

		// Role ADMIN memiliki hak akses penuh
		if isAdmin(roleCode) {
			c.Next()
			return
		}

		// Periksa apakah role pengguna memiliki permission tersebut pada tabel pivot role_has_permissions
		var count int64
		err := db.Table("role_has_permissions").
			Joins("JOIN permissions ON permissions.id = role_has_permissions.permission_id").
			Where("role_has_permissions.role_id = ? AND permissions.code = ?", roleID, permissionCode).
			Count(&count).Error

		if err != nil || count == 0 {
			response.Error(c, http.StatusForbidden, "Akses ditolak: Anda tidak memiliki izin akses untuk fitur ini")
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireAnyPermission memvalidasi apakah pengguna memiliki salah satu dari permission yang ditentukan.
// Jika role pengguna adalah ADMIN, otorisasi akan otomatis di-BYPASS.
func RequireAnyPermission(db *gorm.DB, permissionCodes ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		roleCode, existsRole := GetCurrentUserRole(c)
		roleID, existsID := GetCurrentUserRoleID(c)

		if !existsRole || !existsID || roleID == uuid.Nil {
			response.Error(c, http.StatusUnauthorized, "Sesi tidak terotentikasi")
			c.Abort()
			return
		}

		// Role ADMIN memiliki hak akses penuh
		if isAdmin(roleCode) {
			c.Next()
			return
		}

		if len(permissionCodes) == 0 {
			c.Next()
			return
		}

		var count int64
		err := db.Table("role_has_permissions").
			Joins("JOIN permissions ON permissions.id = role_has_permissions.permission_id").
			Where("role_has_permissions.role_id = ? AND permissions.code IN (?)", roleID, permissionCodes).
			Count(&count).Error

		if err != nil || count == 0 {
			response.Error(c, http.StatusForbidden, "Akses ditolak: Anda tidak memiliki izin yang sesuai")
			c.Abort()
			return
		}

		c.Next()
	}
}
