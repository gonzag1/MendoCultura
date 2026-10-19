package handlers

import (
	"MendoCultura/middleware"
	"MendoCultura/utils"
	"errors"
	"net/http"
	"strconv"
	"time"

	"MendoCultura/models"
	"MendoCultura/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type reserveTicketsRequest struct {
	EventID  uint `json:"eventId"`
	Quantity int  `json:"quantity"`
}

type checkInRequest struct {
	Code    string `json:"code"`
	EventID uint   `json:"eventId"`
}

func (s *Server) ListMyTickets(c *gin.Context) {
	user, _ := middleware.CurrentUser(c)
	var tickets []models.Ticket
	if err := s.db.Preload("Event").Where("user_id = ?", user.ID).Order("created_at desc").Find(&tickets).Error; err != nil {
		utils.ServerError(c, "No se pudieron consultar tus entradas.")
		return
	}
	var purchases []models.Purchase
	if err := s.db.Preload("Event").Where("user_id = ? AND status IN ?", user.ID, []string{models.PurchasePending, "PAYMENT_REVIEW", "EXPIRED"}).Order("created_at desc").Limit(20).Find(&purchases).Error; err != nil {
		utils.ServerError(c, "No se pudieron consultar tus compras.")
		return
	}
	c.JSON(http.StatusOK, gin.H{"tickets": tickets, "purchases": purchases})
}

func (s *Server) GetMyTicket(c *gin.Context) {
	user, _ := middleware.CurrentUser(c)
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	var ticket models.Ticket
	if err := s.db.Preload("Event").Where("id = ? AND user_id = ?", id, user.ID).First(&ticket).Error; err != nil {
		utils.NotFound(c, "La entrada no existe o no te pertenece.")
		return
	}
	c.JSON(http.StatusOK, gin.H{"ticket": ticket})
}

func (s *Server) CheckInTicket(c *gin.Context) {
	user, _ := middleware.CurrentUser(c)
	var req checkInRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Code == "" || req.EventID == 0 {
		utils.BadRequest(c, "El código y el evento son obligatorios.")
		return
	}

	ticketID, err := services.ValidateTicketCode(req.Code, s.config.JWTSecret)
	if err != nil {
		s.audit(&user.ID, "CHECKIN_FAILED", "tickets", "unknown", "Código inválido")
		utils.NotFound(c, "Entrada inexistente.")
		return
	}

	var ticket models.Ticket
	err = s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Preload("Event").First(&ticket, ticketID).Error; err != nil {
			return err
		}
		if ticket.EventID != req.EventID || ticket.Code != req.Code {
			return errors.New("ticket_not_found")
		}
		if ticket.Status == models.TicketUsed {
			return errors.New("ticket_used")
		}
		if ticket.Status == models.TicketCancelled {
			return errors.New("ticket_cancelled")
		}
		if ticket.Status != models.TicketValid {
			return errors.New("ticket_not_valid")
		}
		if ticket.Event.Status != models.EventPublished {
			return errors.New("event_not_active")
		}

		now := time.Now()
		ticket.Status = models.TicketUsed
		ticket.UsedAt = &now
		return tx.Save(&ticket).Error
	})

	if err != nil {
		entityID := strconv.FormatUint(uint64(ticketID), 10)
		switch err.Error() {
		case "ticket_used":
			s.audit(&user.ID, "CHECKIN_DUPLICATED", "tickets", entityID, "Intento de validar entrada ya utilizada")
			utils.Conflict(c, "Entrada ya utilizada.")
		case "ticket_cancelled":
			s.audit(&user.ID, "CHECKIN_FAILED", "tickets", entityID, "Entrada cancelada")
			utils.Conflict(c, "Entrada cancelada.")
		case "ticket_not_valid", "event_not_active":
			s.audit(&user.ID, "CHECKIN_FAILED", "tickets", entityID, "Entrada no válida")
			utils.Conflict(c, "Entrada no válida.")
		default:
			s.audit(&user.ID, "CHECKIN_FAILED", "tickets", entityID, "Entrada inexistente")
			utils.NotFound(c, "Entrada inexistente.")
		}
		return
	}

	s.audit(&user.ID, "CHECKIN_SUCCESS", "tickets", ticket.ID, "Entrada validada correctamente")
	c.JSON(http.StatusOK, gin.H{
		"message": "Entrada válida.",
		"ticket":  ticket,
	})
}
