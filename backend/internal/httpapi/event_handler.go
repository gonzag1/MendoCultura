package httpapi

import (
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode"

	"MendoCultura/internal/domain"

	"github.com/gin-gonic/gin"
)

type eventRequest struct {
	Title               string   `json:"title"`
	Description         string   `json:"description"`
	ExtendedDescription string   `json:"extendedDescription"`
	Category            string   `json:"category"`
	Department          string   `json:"department"`
	Venue               string   `json:"venue"`
	Address             string   `json:"address"`
	ImageURL            string   `json:"imageUrl"`
	GalleryImages       string   `json:"galleryImages"`
	MapURL              string   `json:"mapUrl"`
	Latitude            *float64 `json:"latitude"`
	Longitude           *float64 `json:"longitude"`
	TicketType          string   `json:"ticketType"`
	ImportantInfo       string   `json:"importantInfo"`
	Recommendations     string   `json:"recommendations"`
	StartAt             string   `json:"startAt"`
	PriceCents          int64    `json:"priceCents"`
	Capacity            int      `json:"capacity"`
	Status              string   `json:"status"`
}

func (s *Server) listPublicEvents(c *gin.Context) {
	var events []domain.Event
	if err := s.db.Where("status = ?", domain.EventPublished).Order("start_at asc").Find(&events).Error; err != nil {
		serverError(c, "No se pudieron consultar los eventos.")
		return
	}

	search := normalizeText(c.Query("q"))
	category := normalizeText(c.Query("category"))
	department := normalizeText(c.Query("department"))
	filtered := make([]domain.Event, 0, len(events))

	for _, event := range events {
		if category != "" && normalizeText(event.Category) != category {
			continue
		}
		if department != "" && normalizeText(event.Department) != department {
			continue
		}
		if search != "" {
			haystack := strings.Join([]string{event.Title, event.Description, event.Category, event.Department, event.Venue, event.Address}, " ")
			if !strings.Contains(normalizeText(haystack), search) {
				continue
			}
		}
		filtered = append(filtered, event)
	}

	c.JSON(http.StatusOK, gin.H{"events": filtered})
}

func (s *Server) getPublicEvent(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}

	var event domain.Event
	if err := s.db.Where("id = ? AND status = ?", id, domain.EventPublished).First(&event).Error; err != nil {
		notFound(c, "El evento no existe o no está publicado.")
		return
	}
	c.JSON(http.StatusOK, gin.H{"event": event})
}

func (s *Server) listOrganizerEvents(c *gin.Context) {
	user, _ := currentUser(c)
	var events []domain.Event
	query := s.db.Order("start_at desc")
	if user.Role != domain.RoleAdmin {
		query = query.Where("organizer_id = ?", user.ID)
	}
	if err := query.Find(&events).Error; err != nil {
		serverError(c, "No se pudieron consultar tus eventos.")
		return
	}
	c.JSON(http.StatusOK, gin.H{"events": events})
}

func (s *Server) createOrganizerEvent(c *gin.Context) {
	user, _ := currentUser(c)
	var req eventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, "Los datos del evento no son válidos.")
		return
	}

	event, ok := buildEventFromRequest(req)
	if !ok {
		badRequest(c, "Título, categoría, departamento, lugar, fecha futura, precio y capacidad válida son obligatorios.")
		return
	}
	event.OrganizerID = user.ID
	event.AvailableTickets = event.Capacity
	if event.Status == "" {
		event.Status = domain.EventDraft
	}
	if !validEventStatus(event.Status) {
		badRequest(c, "El estado del evento no es válido.")
		return
	}

	if err := s.db.Create(&event).Error; err != nil {
		serverError(c, "No se pudo crear el evento.")
		return
	}

	s.audit(&user.ID, "EVENT_CREATED", "events", event.ID, "Evento creado")
	c.JSON(http.StatusCreated, gin.H{"event": event})
}

func (s *Server) updateOrganizerEvent(c *gin.Context) {
	user, _ := currentUser(c)
	id, ok := parseID(c, "id")
	if !ok {
		return
	}

	var event domain.Event
	query := s.db.Where("id = ?", id)
	if user.Role != domain.RoleAdmin {
		query = query.Where("organizer_id = ?", user.ID)
	}
	if err := query.First(&event).Error; err != nil {
		notFound(c, "El evento no existe o no te pertenece.")
		return
	}

	var req eventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, "Los datos del evento no son válidos.")
		return
	}

	updated, ok := buildEventFromRequest(req)
	if !ok {
		badRequest(c, "Los datos del evento no son válidos.")
		return
	}
	if updated.Capacity < event.Capacity-event.AvailableTickets {
		badRequest(c, "No podés reducir la capacidad por debajo de las entradas vendidas.")
		return
	}
	if !validEventStatus(updated.Status) {
		badRequest(c, "El estado del evento no es válido.")
		return
	}

	sold := event.Capacity - event.AvailableTickets
	event.Title = updated.Title
	event.Description = updated.Description
	event.ExtendedDescription = updated.ExtendedDescription
	event.Category = updated.Category
	event.Department = updated.Department
	event.Venue = updated.Venue
	event.Address = updated.Address
	event.ImageURL = updated.ImageURL
	event.GalleryImages = updated.GalleryImages
	event.MapURL = updated.MapURL
	event.Latitude = updated.Latitude
	event.Longitude = updated.Longitude
	event.TicketType = updated.TicketType
	event.ImportantInfo = updated.ImportantInfo
	event.Recommendations = updated.Recommendations
	event.StartAt = updated.StartAt
	event.PriceCents = updated.PriceCents
	event.Capacity = updated.Capacity
	event.AvailableTickets = updated.Capacity - sold
	event.Status = updated.Status

	if err := s.db.Save(&event).Error; err != nil {
		serverError(c, "No se pudo actualizar el evento.")
		return
	}

	s.audit(&user.ID, "EVENT_UPDATED", "events", event.ID, "Evento actualizado")
	c.JSON(http.StatusOK, gin.H{"event": event})
}

func (s *Server) listAdminEvents(c *gin.Context) {
	var events []domain.Event
	if err := s.db.Order("created_at desc").Find(&events).Error; err != nil {
		serverError(c, "No se pudieron consultar los eventos.")
		return
	}
	c.JSON(http.StatusOK, gin.H{"events": events})
}

func (s *Server) updateEventStatus(c *gin.Context) {
	user, _ := currentUser(c)
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	var req struct {
		Status string `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || !validEventStatus(req.Status) {
		badRequest(c, "El estado del evento no es válido.")
		return
	}

	if err := s.db.Model(&domain.Event{}).Where("id = ?", id).Update("status", req.Status).Error; err != nil {
		serverError(c, "No se pudo actualizar el estado del evento.")
		return
	}
	s.audit(&user.ID, "EVENT_STATUS_UPDATED", "events", id, "Estado de evento actualizado a "+req.Status)
	c.JSON(http.StatusOK, gin.H{"status": req.Status})
}

func buildEventFromRequest(req eventRequest) (domain.Event, bool) {
	startAt, err := time.Parse(time.RFC3339, req.StartAt)
	if err != nil {
		return domain.Event{}, false
	}
	ticketType := strings.TrimSpace(req.TicketType)
	if ticketType == "" {
		ticketType = "General"
	}
	if req.Title == "" || req.Category == "" || req.Department == "" || req.Venue == "" || req.Capacity <= 0 || req.PriceCents < 0 || startAt.Before(time.Now()) {
		return domain.Event{}, false
	}
	status := req.Status
	if status == "" {
		status = domain.EventDraft
	}
	return domain.Event{
		Title:               strings.TrimSpace(req.Title),
		Description:         strings.TrimSpace(req.Description),
		ExtendedDescription: strings.TrimSpace(req.ExtendedDescription),
		Category:            strings.TrimSpace(req.Category),
		Department:          strings.TrimSpace(req.Department),
		Venue:               strings.TrimSpace(req.Venue),
		Address:             strings.TrimSpace(req.Address),
		ImageURL:            strings.TrimSpace(req.ImageURL),
		GalleryImages:       strings.TrimSpace(req.GalleryImages),
		MapURL:              strings.TrimSpace(req.MapURL),
		Latitude:            req.Latitude,
		Longitude:           req.Longitude,
		TicketType:          ticketType,
		ImportantInfo:       strings.TrimSpace(req.ImportantInfo),
		Recommendations:     strings.TrimSpace(req.Recommendations),
		StartAt:             startAt,
		PriceCents:          req.PriceCents,
		Capacity:            req.Capacity,
		Status:              status,
	}, true
}

func validEventStatus(status string) bool {
	switch status {
	case domain.EventDraft, domain.EventPublished, domain.EventPaused, domain.EventCancelled, domain.EventFinished:
		return true
	default:
		return false
	}
}

func parseID(c *gin.Context, param string) (uint, bool) {
	value, err := strconv.ParseUint(c.Param(param), 10, 64)
	if err != nil {
		badRequest(c, "El identificador no es válido.")
		return 0, false
	}
	return uint(value), true
}

func normalizeText(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	var builder strings.Builder
	builder.Grow(len(value))
	for _, r := range value {
		switch r {
		case 'á', 'à', 'ä', 'â':
			r = 'a'
		case 'é', 'è', 'ë', 'ê':
			r = 'e'
		case 'í', 'ì', 'ï', 'î':
			r = 'i'
		case 'ó', 'ò', 'ö', 'ô':
			r = 'o'
		case 'ú', 'ù', 'ü', 'û':
			r = 'u'
		case 'ñ':
			r = 'n'
		}
		if unicode.IsSpace(r) {
			r = ' '
		}
		builder.WriteRune(r)
	}
	return strings.Join(strings.Fields(builder.String()), " ")
}
