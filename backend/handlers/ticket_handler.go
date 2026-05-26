package handlers

import (
	"MendoCultura/middleware"
	"MendoCultura/utils"
	"database/sql"
	"errors"
	"fmt"
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

func (s *Server) ReserveTickets(c *gin.Context) {
	user, _ := middleware.CurrentUser(c)
	var req reserveTicketsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "Los datos de compra no son válidos.")
		return
	}
	if req.EventID == 0 || req.Quantity <= 0 || req.Quantity > 4 {
		utils.BadRequest(c, "Podés comprar entre 1 y 4 entradas por evento.")
		return
	}

	var purchase models.Purchase
	var tickets []models.Ticket
	err := s.db.Transaction(func(tx *gorm.DB) error {
		var event models.Event
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&event, req.EventID).Error; err != nil {
			return err
		}
		if event.Status != models.EventPublished || !event.StartAt.After(time.Now()) {
			return errors.New("event_not_available")
		}
		if event.AvailableTickets < req.Quantity {
			return errors.New("not_enough_tickets")
		}

		event.AvailableTickets -= req.Quantity
		if err := tx.Save(&event).Error; err != nil {
			return err
		}

		purchase = models.Purchase{
			UserID:     user.ID,
			EventID:    event.ID,
			Quantity:   req.Quantity,
			TotalCents: event.PriceCents * int64(req.Quantity),
			Status:     models.PurchasePaid,
		}
		if err := tx.Create(&purchase).Error; err != nil {
			return err
		}

		for i := 0; i < req.Quantity; i++ {
			ticket := models.Ticket{
				UserID:     user.ID,
				EventID:    event.ID,
				PurchaseID: purchase.ID,
				Status:     models.TicketValid,
				Code:       "pending",
			}
			if err := tx.Create(&ticket).Error; err != nil {
				return err
			}
			code, err := services.NewTicketCode(ticket.ID, s.config.JWTSecret)
			if err != nil {
				return err
			}
			ticket.Code = code
			if err := tx.Save(&ticket).Error; err != nil {
				return err
			}
			tickets = append(tickets, ticket)
		}
		return nil
	}, &sql.TxOptions{Isolation: sql.LevelSerializable})

	if err != nil {
		switch err.Error() {
		case "event_not_available":
			utils.Conflict(c, "El evento no está disponible para comprar entradas.")
		case "not_enough_tickets":
			utils.Conflict(c, "No hay cupos disponibles suficientes.")
		default:
			utils.ServerError(c, "No se pudo completar la compra simulada.")
		}
		return
	}

	s.audit(&user.ID, "TICKETS_RESERVED", "purchases", purchase.ID, fmt.Sprintf("Compra simulada de %d entradas", req.Quantity))
	c.JSON(http.StatusCreated, gin.H{"purchase": purchase, "tickets": tickets})
}

func (s *Server) ListMyTickets(c *gin.Context) {
	user, _ := middleware.CurrentUser(c)
	var tickets []models.Ticket
	if err := s.db.Preload("Event").Where("user_id = ?", user.ID).Order("created_at desc").Find(&tickets).Error; err != nil {
		utils.ServerError(c, "No se pudieron consultar tus entradas.")
		return
	}
	c.JSON(http.StatusOK, gin.H{"tickets": tickets})
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
