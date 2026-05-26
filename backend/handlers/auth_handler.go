package handlers

import (
	"MendoCultura/middleware"
	"MendoCultura/utils"
	"net/http"
	"strings"
	"time"

	"MendoCultura/models"
	"MendoCultura/services"

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

func (s *Server) RegisterUser(c *gin.Context) {
	var req registerUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "Los datos enviados no son válidos.")
		return
	}

	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	if req.Name == "" || req.Email == "" || len(req.Password) < 6 {
		utils.BadRequest(c, "Nombre, email y contraseña de al menos 6 caracteres son obligatorios.")
		return
	}

	hash, err := services.HashPassword(req.Password)
	if err != nil {
		utils.ServerError(c, "No se pudo proteger la contraseña.")
		return
	}

	user := models.User{Name: req.Name, Email: req.Email, PasswordHash: hash, DNI: req.DNI, Role: models.RoleUser, Status: models.AccountActive}
	if err := s.db.Create(&user).Error; err != nil {
		utils.Conflict(c, "Ya existe una cuenta con ese email.")
		return
	}

	s.audit(&user.ID, "USER_REGISTERED", "users", user.ID, "Cuenta de usuario creada")
	c.JSON(http.StatusCreated, gin.H{"user": user})
}

func (s *Server) RegisterOrganizer(c *gin.Context) {
	var req registerOrganizerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "Los datos enviados no son válidos.")
		return
	}

	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	if req.Name == "" || req.Email == "" || len(req.Password) < 6 || req.BusinessName == "" || req.TaxID == "" || req.Locality == "" {
		utils.BadRequest(c, "Los datos personales y fiscales del organizador son obligatorios.")
		return
	}

	hash, err := services.HashPassword(req.Password)
	if err != nil {
		utils.ServerError(c, "No se pudo proteger la contraseña.")
		return
	}

	var user models.User
	err = s.db.Transaction(func(tx *gorm.DB) error {
		user = models.User{Name: req.Name, Email: req.Email, PasswordHash: hash, DNI: req.DNI, Role: models.RoleOrganizer, Status: models.AccountActive}
		if err := tx.Create(&user).Error; err != nil {
			return err
		}

		profile := models.OrganizerProfile{
			UserID:       user.ID,
			BusinessName: req.BusinessName,
			TaxID:        req.TaxID,
			Locality:     req.Locality,
			Status:       models.AccountPending,
		}
		return tx.Create(&profile).Error
	})
	if err != nil {
		utils.Conflict(c, "Ya existe una cuenta con ese email o CUIT.")
		return
	}

	s.audit(&user.ID, "ORGANIZER_REGISTERED", "organizer_profiles", user.ID, "Solicitud de organizador creada")
	c.JSON(http.StatusCreated, gin.H{"user": user, "organizerStatus": models.AccountPending})
}

func (s *Server) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "Los datos enviados no son válidos.")
		return
	}

	var user models.User
	if err := s.db.Where("email = ?", strings.TrimSpace(strings.ToLower(req.Email))).First(&user).Error; err != nil {
		utils.Unauthorized(c, "Email o contraseña incorrectos.")
		return
	}
	if user.Status != models.AccountActive {
		utils.Forbidden(c, "La cuenta no está activa.")
		return
	}
	if !services.CheckPassword(user.PasswordHash, req.Password) {
		utils.Unauthorized(c, "Email o contraseña incorrectos.")
		return
	}

	token, err := services.GenerateJWT(services.Claims{
		UserID: user.ID,
		Email:  user.Email,
		Role:   user.Role,
		Exp:    time.Now().Add(24 * time.Hour).Unix(),
	}, s.config.JWTSecret)
	if err != nil {
		utils.ServerError(c, "No se pudo crear la sesión.")
		return
	}

	s.audit(&user.ID, "USER_LOGIN", "users", user.ID, "Inicio de sesión correcto")
	c.JSON(http.StatusOK, gin.H{"token": token, "user": user})
}

func (s *Server) GetMe(c *gin.Context) {
	user, _ := middleware.CurrentUser(c)
	c.JSON(http.StatusOK, gin.H{"user": user})
}
