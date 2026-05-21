package httpapi

import (
	"strings"

	"MendoCultura/internal/domain"
	"MendoCultura/internal/security"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const currentUserKey = "currentUser"

func authMiddleware(db *gorm.DB, jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			unauthorized(c, "Tenés que iniciar sesión para continuar.")
			return
		}

		claims, err := security.ParseJWT(strings.TrimPrefix(header, "Bearer "), jwtSecret)
		if err != nil {
			unauthorized(c, "La sesión no es válida o expiró.")
			return
		}

		var user domain.User
		if err := db.First(&user, claims.UserID).Error; err != nil {
			unauthorized(c, "El usuario de la sesión no existe.")
			return
		}
		if user.Status != domain.AccountActive {
			forbidden(c, "La cuenta no está activa.")
			return
		}

		c.Set(currentUserKey, user)
		c.Next()
	}
}

func requireRoles(roles ...string) gin.HandlerFunc {
	allowed := map[string]bool{}
	for _, role := range roles {
		allowed[role] = true
	}

	return func(c *gin.Context) {
		user, ok := currentUser(c)
		if !ok {
			unauthorized(c, "Tenés que iniciar sesión para continuar.")
			return
		}
		if !allowed[user.Role] {
			forbidden(c, "No tenés permisos para realizar esta acción.")
			return
		}
		c.Next()
	}
}

func requireApprovedOrganizer(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := currentUser(c)
		if !ok {
			unauthorized(c, "Tenés que iniciar sesión para continuar.")
			return
		}
		if user.Role == domain.RoleAdmin || user.Role == domain.RoleValidator {
			c.Next()
			return
		}
		if user.Role != domain.RoleOrganizer {
			forbidden(c, "No tenés permisos para realizar esta acción.")
			return
		}

		var profile domain.OrganizerProfile
		if err := db.Where("user_id = ? AND status = ?", user.ID, "APPROVED").First(&profile).Error; err != nil {
			forbidden(c, "El organizador todavía no está aprobado.")
			return
		}
		c.Next()
	}
}

func currentUser(c *gin.Context) (domain.User, bool) {
	value, exists := c.Get(currentUserKey)
	if !exists {
		return domain.User{}, false
	}
	user, ok := value.(domain.User)
	return user, ok
}
