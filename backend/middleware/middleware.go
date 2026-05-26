package middleware

import (
	"net/http"
	"strings"

	"MendoCultura/models"
	"MendoCultura/services"
	"MendoCultura/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const currentUserKey = "currentUser"

func Auth(db *gorm.DB, jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			utils.Unauthorized(c, "Tenés que iniciar sesión para continuar.")
			return
		}

		claims, err := services.ParseJWT(strings.TrimPrefix(header, "Bearer "), jwtSecret)
		if err != nil {
			utils.Unauthorized(c, "La sesión no es válida o expiró.")
			return
		}

		var user models.User
		if err := db.First(&user, claims.UserID).Error; err != nil {
			utils.Unauthorized(c, "El usuario de la sesión no existe.")
			return
		}
		if user.Status != models.AccountActive {
			utils.Forbidden(c, "La cuenta no está activa.")
			return
		}

		c.Set(currentUserKey, user)
		c.Next()
	}
}

func RequireRoles(roles ...string) gin.HandlerFunc {
	allowed := map[string]bool{}
	for _, role := range roles {
		allowed[role] = true
	}

	return func(c *gin.Context) {
		user, ok := CurrentUser(c)
		if !ok {
			utils.Unauthorized(c, "Tenés que iniciar sesión para continuar.")
			return
		}
		if !allowed[user.Role] {
			utils.Forbidden(c, "No tenés permisos para realizar esta acción.")
			return
		}
		c.Next()
	}
}

func RequireApprovedOrganizer(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := CurrentUser(c)
		if !ok {
			utils.Unauthorized(c, "Tenés que iniciar sesión para continuar.")
			return
		}
		if user.Role == models.RoleAdmin || user.Role == models.RoleValidator {
			c.Next()
			return
		}
		if user.Role != models.RoleOrganizer {
			utils.Forbidden(c, "No tenés permisos para realizar esta acción.")
			return
		}

		var profile models.OrganizerProfile
		if err := db.Where("user_id = ? AND status = ?", user.ID, "APPROVED").First(&profile).Error; err != nil {
			utils.Forbidden(c, "El organizador todavía no está aprobado.")
			return
		}
		c.Next()
	}
}

func CurrentUser(c *gin.Context) (models.User, bool) {
	value, exists := c.Get(currentUserKey)
	if !exists {
		return models.User{}, false
	}
	user, ok := value.(models.User)
	return user, ok
}

func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
