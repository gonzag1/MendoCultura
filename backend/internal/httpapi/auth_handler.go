package httpapi

import (
	"net/http"
	"strings"
	"time"

	"MendoCultura/internal/domain"
	"MendoCultura/internal/security"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type registerUserRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
	DNI      string `json:"dni"`
}

type registerOrganizerRequest struct {
	Name         string `json:"name"`
	Email        string `json:"email"`
	Password     string `json:"password"`
	DNI          string `json:"dni"`
	BusinessName string `json:"businessName"`
	TaxID        string `json:"taxId"`
	Locality     string `json:"locality"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (s *Server) registerUser(c *gin.Context) {
	var req registerUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, "Los datos enviados no son válidos.")
		return
	}

	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	if req.Name == "" || req.Email == "" || len(req.Password) < 6 {
		badRequest(c, "Nombre, email y contraseña de al menos 6 caracteres son obligatorios.")
		return
	}

	hash, err := security.HashPassword(req.Password)
	if err != nil {
		serverError(c, "No se pudo proteger la contraseña.")
		return
	}

	user := domain.User{Name: req.Name, Email: req.Email, PasswordHash: hash, DNI: req.DNI, Role: domain.RoleUser, Status: domain.AccountActive}
	if err := s.db.Create(&user).Error; err != nil {
		conflict(c, "Ya existe una cuenta con ese email.")
		return
	}

	s.audit(&user.ID, "USER_REGISTERED", "users", user.ID, "Cuenta de usuario creada")
	c.JSON(http.StatusCreated, gin.H{"user": user})
}

func (s *Server) registerOrganizer(c *gin.Context) {
	var req registerOrganizerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, "Los datos enviados no son válidos.")
		return
	}

	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	if req.Name == "" || req.Email == "" || len(req.Password) < 6 || req.BusinessName == "" || req.TaxID == "" || req.Locality == "" {
		badRequest(c, "Los datos personales y fiscales del organizador son obligatorios.")
		return
	}

	hash, err := security.HashPassword(req.Password)
	if err != nil {
		serverError(c, "No se pudo proteger la contraseña.")
		return
	}

	var user domain.User
	err = s.db.Transaction(func(tx *gorm.DB) error {
		user = domain.User{Name: req.Name, Email: req.Email, PasswordHash: hash, DNI: req.DNI, Role: domain.RoleOrganizer, Status: domain.AccountActive}
		if err := tx.Create(&user).Error; err != nil {
			return err
		}

		profile := domain.OrganizerProfile{
			UserID:       user.ID,
			BusinessName: req.BusinessName,
			TaxID:        req.TaxID,
			Locality:     req.Locality,
			Status:       domain.AccountPending,
		}
		return tx.Create(&profile).Error
	})
	if err != nil {
		conflict(c, "Ya existe una cuenta con ese email o CUIT.")
		return
	}

	s.audit(&user.ID, "ORGANIZER_REGISTERED", "organizer_profiles", user.ID, "Solicitud de organizador creada")
	c.JSON(http.StatusCreated, gin.H{"user": user, "organizerStatus": domain.AccountPending})
}

func (s *Server) login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, "Los datos enviados no son válidos.")
		return
	}

	var user domain.User
	if err := s.db.Where("email = ?", strings.TrimSpace(strings.ToLower(req.Email))).First(&user).Error; err != nil {
		unauthorized(c, "Email o contraseña incorrectos.")
		return
	}
	if user.Status != domain.AccountActive {
		forbidden(c, "La cuenta no está activa.")
		return
	}
	if !security.CheckPassword(user.PasswordHash, req.Password) {
		unauthorized(c, "Email o contraseña incorrectos.")
		return
	}

	token, err := security.GenerateJWT(security.Claims{
		UserID: user.ID,
		Email:  user.Email,
		Role:   user.Role,
		Exp:    time.Now().Add(24 * time.Hour).Unix(),
	}, s.config.JWTSecret)
	if err != nil {
		serverError(c, "No se pudo crear la sesión.")
		return
	}

	s.audit(&user.ID, "USER_LOGIN", "users", user.ID, "Inicio de sesión correcto")
	c.JSON(http.StatusOK, gin.H{"token": token, "user": user})
}

func (s *Server) getMe(c *gin.Context) {
	user, _ := currentUser(c)
	c.JSON(http.StatusOK, gin.H{"user": user})
}
