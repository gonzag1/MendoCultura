package httpapi

import (
	"net/http"

	"MendoCultura/internal/config"
	"MendoCultura/internal/domain"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Server struct {
	db     *gorm.DB
	config config.Config
}

func NewRouter(db *gorm.DB, cfg config.Config) *gin.Engine {
	server := &Server{db: db, config: cfg}
	router := gin.Default()

	router.Use(corsMiddleware())

	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "MendoCultura funcionando"})
	})
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	api := router.Group("/api/v1")
	api.POST("/auth/register", server.registerUser)
	api.POST("/auth/register-organizer", server.registerOrganizer)
	api.POST("/auth/login", server.login)
	api.GET("/events", server.listPublicEvents)
	api.GET("/events/:id", server.getPublicEvent)

	protected := api.Group("")
	protected.Use(authMiddleware(db, cfg.JWTSecret))
	protected.GET("/me", server.getMe)
	protected.POST("/tickets/reserve", requireRoles(domain.RoleUser, domain.RoleOrganizer, domain.RoleAdmin), server.reserveTickets)
	protected.GET("/tickets", server.listMyTickets)
	protected.GET("/tickets/:id", server.getMyTicket)
	protected.POST("/checkin", requireRoles(domain.RoleOrganizer, domain.RoleValidator, domain.RoleAdmin), requireApprovedOrganizer(db), server.checkInTicket)

	organizer := protected.Group("/organizer")
	organizer.Use(requireRoles(domain.RoleOrganizer, domain.RoleAdmin))
	organizer.Use(requireApprovedOrganizer(db))
	organizer.GET("/events", server.listOrganizerEvents)
	organizer.POST("/events", server.createOrganizerEvent)
	organizer.PUT("/events/:id", server.updateOrganizerEvent)
	organizer.GET("/reports", server.getOrganizerReports)

	admin := protected.Group("/admin")
	admin.Use(requireRoles(domain.RoleAdmin))
	admin.GET("/users", server.listUsers)
	admin.GET("/organizers", server.listOrganizers)
	admin.PUT("/organizers/:id/status", server.updateOrganizerStatus)
	admin.GET("/events", server.listAdminEvents)
	admin.PUT("/events/:id/status", server.updateEventStatus)
	admin.GET("/metrics", server.getAdminMetrics)
	admin.GET("/audit-logs", server.listAuditLogs)

	return router
}

func corsMiddleware() gin.HandlerFunc {
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
