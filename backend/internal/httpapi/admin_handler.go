package httpapi

import (
	"fmt"
	"net/http"
	"strconv"

	"MendoCultura/internal/domain"

	"github.com/gin-gonic/gin"
)

func (s *Server) listUsers(c *gin.Context) {
	var users []domain.User
	if err := s.db.Order("created_at desc").Find(&users).Error; err != nil {
		serverError(c, "No se pudieron consultar los usuarios.")
		return
	}
	c.JSON(http.StatusOK, gin.H{"users": users})
}

func (s *Server) listOrganizers(c *gin.Context) {
	var profiles []domain.OrganizerProfile
	if err := s.db.Preload("User").Order("created_at desc").Find(&profiles).Error; err != nil {
		serverError(c, "No se pudieron consultar los organizadores.")
		return
	}
	c.JSON(http.StatusOK, gin.H{"organizers": profiles})
}

func (s *Server) updateOrganizerStatus(c *gin.Context) {
	user, _ := currentUser(c)
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	var req struct {
		Status string `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || !validAccountDecision(req.Status) {
		badRequest(c, "El estado del organizador debe ser APPROVED, REJECTED o SUSPENDED.")
		return
	}

	profileStatus := req.Status
	userStatus := domain.AccountActive
	if req.Status == "APPROVED" {
		profileStatus = "APPROVED"
	}
	if req.Status == domain.AccountRejected {
		userStatus = domain.AccountSuspended
	}
	if req.Status == domain.AccountSuspended {
		userStatus = domain.AccountSuspended
	}

	var profile domain.OrganizerProfile
	if err := s.db.First(&profile, id).Error; err != nil {
		notFound(c, "El organizador no existe.")
		return
	}
	if err := s.db.Model(&profile).Update("status", profileStatus).Error; err != nil {
		serverError(c, "No se pudo actualizar el organizador.")
		return
	}
	if err := s.db.Model(&domain.User{}).Where("id = ?", profile.UserID).Update("status", userStatus).Error; err != nil {
		serverError(c, "No se pudo actualizar la cuenta del organizador.")
		return
	}

	s.audit(&user.ID, "ORGANIZER_STATUS_UPDATED", "organizer_profiles", profile.ID, "Estado actualizado a "+profileStatus)
	c.JSON(http.StatusOK, gin.H{"status": profileStatus})
}

func (s *Server) getAdminMetrics(c *gin.Context) {
	var users int64
	var events int64
	var tickets int64
	var usedTickets int64
	var purchases int64
	s.db.Model(&domain.User{}).Count(&users)
	s.db.Model(&domain.Event{}).Count(&events)
	s.db.Model(&domain.Ticket{}).Count(&tickets)
	s.db.Model(&domain.Ticket{}).Where("status = ?", domain.TicketUsed).Count(&usedTickets)
	s.db.Model(&domain.Purchase{}).Where("status = ?", domain.PurchasePaid).Count(&purchases)

	c.JSON(http.StatusOK, gin.H{
		"users":       users,
		"events":      events,
		"tickets":     tickets,
		"usedTickets": usedTickets,
		"purchases":   purchases,
	})
}

func (s *Server) getOrganizerReports(c *gin.Context) {
	user, _ := currentUser(c)
	var events []domain.Event
	query := s.db
	if user.Role != domain.RoleAdmin {
		query = query.Where("organizer_id = ?", user.ID)
	}
	if err := query.Find(&events).Error; err != nil {
		serverError(c, "No se pudieron calcular las estadísticas.")
		return
	}

	type report struct {
		EventID          uint   `json:"eventId"`
		Title            string `json:"title"`
		SoldTickets      int    `json:"soldTickets"`
		AvailableTickets int    `json:"availableTickets"`
		EstimatedIncome  int64  `json:"estimatedIncome"`
		UsedTickets      int64  `json:"usedTickets"`
		OccupancyPercent int    `json:"occupancyPercent"`
	}

	reports := []report{}
	for _, event := range events {
		sold := event.Capacity - event.AvailableTickets
		var used int64
		s.db.Model(&domain.Ticket{}).Where("event_id = ? AND status = ?", event.ID, domain.TicketUsed).Count(&used)
		occupancy := 0
		if event.Capacity > 0 {
			occupancy = (sold * 100) / event.Capacity
		}
		reports = append(reports, report{
			EventID:          event.ID,
			Title:            event.Title,
			SoldTickets:      sold,
			AvailableTickets: event.AvailableTickets,
			EstimatedIncome:  int64(sold) * event.PriceCents,
			UsedTickets:      used,
			OccupancyPercent: occupancy,
		})
	}

	c.JSON(http.StatusOK, gin.H{"reports": reports})
}

func (s *Server) listAuditLogs(c *gin.Context) {
	var logs []domain.AuditLog
	query := s.db.Order("created_at desc").Limit(100)
	if action := c.Query("action"); action != "" {
		query = query.Where("action = ?", action)
	}
	if userID := c.Query("userId"); userID != "" {
		query = query.Where("user_id = ?", userID)
	}
	if err := query.Find(&logs).Error; err != nil {
		serverError(c, "No se pudieron consultar los registros de auditoría.")
		return
	}
	c.JSON(http.StatusOK, gin.H{"auditLogs": logs})
}

func (s *Server) audit(userID *uint, action string, entity string, entityID any, detail string) {
	log := domain.AuditLog{
		UserID:   userID,
		Action:   action,
		Entity:   entity,
		EntityID: fmt.Sprint(entityID),
		Detail:   detail,
	}
	_ = s.db.Create(&log).Error
}

func validAccountDecision(status string) bool {
	switch status {
	case "APPROVED", domain.AccountRejected, domain.AccountSuspended:
		return true
	default:
		return false
	}
}

func idString(id uint) string {
	return strconv.FormatUint(uint64(id), 10)
}
