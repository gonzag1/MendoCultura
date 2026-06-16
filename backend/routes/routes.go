package routes

import (
	"net/http"

	"MendoCultura/config"
	"MendoCultura/handlers"
	"MendoCultura/middleware"
	"MendoCultura/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func NewRouter(db *gorm.DB, cfg config.Config) *gin.Engine {
	server := handlers.NewServer(db, cfg)
	router := gin.Default()

	router.Use(middleware.CORS())

	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "MendoCultura funcionando"})
	})
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	router.Static("/uploads", "./uploads")

	api := router.Group("/api/v1")
	api.POST("/auth/register", server.RegisterUser)
	api.POST("/auth/register-organizer", server.RegisterOrganizer)
	api.POST("/auth/login", server.Login)
	api.GET("/events", server.ListPublicEvents)
	api.GET("/events/:id", server.GetPublicEvent)

	protected := api.Group("")
	protected.Use(middleware.Auth(db, cfg.JWTSecret))
	protected.GET("/me", server.GetMe)
	protected.POST("/tickets/reserve", middleware.RequireRoles(models.RoleUser, models.RoleOrganizer, models.RoleAdmin), server.ReserveTickets)
	protected.GET("/tickets", server.ListMyTickets)
	protected.GET("/tickets/:id", server.GetMyTicket)
	protected.POST("/checkin", middleware.RequireRoles(models.RoleOrganizer, models.RoleValidator, models.RoleAdmin), middleware.RequireApprovedOrganizer(db), server.CheckInTicket)

	organizer := protected.Group("/organizer")
	organizer.Use(middleware.RequireRoles(models.RoleOrganizer, models.RoleAdmin))
	organizer.Use(middleware.RequireApprovedOrganizer(db))
	organizer.GET("/events", server.ListOrganizerEvents)
	organizer.POST("/events", server.CreateOrganizerEvent)
	organizer.PUT("/events/:id", server.UpdateOrganizerEvent)
	organizer.POST("/events/upload-image", server.UploadOrganizerEventImage)
	organizer.GET("/reports", server.GetOrganizerReports)

	admin := protected.Group("/admin")
	admin.Use(middleware.RequireRoles(models.RoleAdmin))
	admin.GET("/users", server.ListUsers)
	admin.GET("/organizers", server.ListOrganizers)
	admin.PUT("/organizers/:id/status", server.UpdateOrganizerStatus)
	admin.GET("/events", server.ListAdminEvents)
	admin.PUT("/events/:id/status", server.UpdateEventStatus)
	admin.GET("/metrics", server.GetAdminMetrics)
	admin.GET("/audit-logs", server.ListAuditLogs)

	return router
}
