package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"MendoCultura/middleware"
	"MendoCultura/models"
	"MendoCultura/services"
	"MendoCultura/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (s *Server) paymentReady() bool {
	back, e1 := url.Parse(s.config.BackendURL)
	front, e2 := url.Parse(s.config.FrontendURL)
	return s.config.MPAccessToken != "" && s.config.MPWebhookSecret != "" && s.config.MPCollectorID != "" && e1 == nil && e2 == nil && back.Scheme == "https" && back.Host != "" && front.Scheme == "https" && front.Host != ""
}

// All stock operations lock the event first, then purchases, in the same order.
// Expired holds are reclaimed on the next checkout for this event.
func releaseExpired(tx *gorm.DB, event *models.Event) error {
	var expired []models.Purchase
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("event_id = ? AND status = ? AND expires_at < ?", event.ID, models.PurchasePending, time.Now()).Find(&expired).Error; err != nil {
		return err
	}
	for i := range expired {
		event.AvailableTickets += expired[i].Quantity
		if err := tx.Model(&expired[i]).Update("status", "EXPIRED").Error; err != nil {
			return err
		}
	}
	return tx.Save(event).Error
}

// Public catalog reads also reclaim holds, so a sold-out page can reopen
// after abandoned checkouts expire, without requiring a buyer to POST first.
func (s *Server) expirePaymentHolds() error {
	var ids []uint
	if err := s.db.Model(&models.Purchase{}).Where("status = ? AND expires_at < ?", models.PurchasePending, time.Now()).Distinct("event_id").Order("event_id").Pluck("event_id", &ids).Error; err != nil {
		return err
	}
	for _, id := range ids {
		if err := s.db.Transaction(func(tx *gorm.DB) error {
			var event models.Event
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&event, id).Error; err != nil {
				return err
			}
			return releaseExpired(tx, &event)
		}); err != nil {
			return err
		}
	}
	return nil
}

func (s *Server) CreateCheckout(c *gin.Context) {
	if !s.paymentReady() {
		c.JSON(http.StatusServiceUnavailable, gin.H{"message": "Mercado Pago todavía no está configurado."})
		return
	}
	user, _ := middleware.CurrentUser(c)
	var req reserveTicketsRequest
	if c.ShouldBindJSON(&req) != nil || req.EventID == 0 || req.Quantity < 1 || req.Quantity > 4 {
		utils.BadRequest(c, "Podés comprar entre 1 y 4 entradas por evento.")
		return
	}
	var purchase models.Purchase
	var event models.Event
	random := make([]byte, 24)
	if _, err := rand.Read(random); err != nil {
		utils.ServerError(c, "No se pudo iniciar el pago.")
		return
	}
	expires := time.Now().UTC().Add(30 * time.Minute)
	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&event, req.EventID).Error; err != nil {
			return err
		}
		if err := releaseExpired(tx, &event); err != nil {
			return err
		}
		if event.Status != models.EventPublished || !event.StartAt.After(time.Now()) {
			return errors.New("unavailable")
		}
		// At most one open checkout per user/event; retry reuses its provider URL.
		if err := tx.Where("user_id = ? AND event_id = ? AND status = ?", user.ID, event.ID, models.PurchasePending).First(&purchase).Error; err == nil {
			return nil
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		if !event.StartAt.After(expires) || event.PriceCents <= 0 || event.PriceCents > math.MaxInt64/int64(req.Quantity) || event.AvailableTickets < req.Quantity {
			return errors.New("unavailable")
		}
		purchase = models.Purchase{UserID: user.ID, EventID: event.ID, Quantity: req.Quantity, TotalCents: event.PriceCents * int64(req.Quantity), Status: models.PurchasePending, ExpiresAt: &expires, PaymentReference: hex.EncodeToString(random)}
		if err := tx.Create(&purchase).Error; err != nil {
			return err
		}
		event.AvailableTickets -= req.Quantity
		return tx.Save(&event).Error
	})
	if err != nil {
		utils.Conflict(c, "No hay cupos suficientes o el evento no está disponible para el pago.")
		return
	}
	if purchase.CheckoutURL != "" {
		c.JSON(http.StatusOK, gin.H{"purchase": purchase, "checkoutUrl": purchase.CheckoutURL})
		return
	}
	// A timed-out request may have succeeded remotely. Never create a second
	// preference for the same purchase after an uncertain response or crash.
	claim := s.db.Model(&models.Purchase{}).Where("id = ? AND preference_attempted = ?", purchase.ID, false).Update("preference_attempted", true)
	if claim.Error != nil || claim.RowsAffected != 1 {
		utils.Conflict(c, "El checkout está en preparación o necesita revisión. Si ya pagaste, no vuelvas a pagar; revisá Mis entradas.")
		return
	}
	// Serialize preference creation across duplicate requests without locking stock.
	err = s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&purchase, purchase.ID).Error; err != nil {
			return err
		}
		if purchase.Status != models.PurchasePending {
			return errors.New("expired")
		}
		if purchase.CheckoutURL != "" {
			return nil
		}
		back := strings.TrimRight(s.config.FrontendURL, "/") + "/payment-result?purchaseId=" + strconv.FormatUint(uint64(purchase.ID), 10)
		unitCents := purchase.TotalCents / int64(purchase.Quantity)
		payload := map[string]any{
			"items": []any{map[string]any{"id": fmt.Sprint(event.ID), "title": event.Title, "quantity": purchase.Quantity, "currency_id": "ARS", "unit_price": json.Number(fmt.Sprintf("%d.%02d", unitCents/100, unitCents%100))}},
			"payer": map[string]any{"email": user.Email}, "external_reference": purchase.PaymentReference,
			"notification_url": strings.TrimRight(s.config.BackendURL, "/") + "/api/v1/payments/webhook",
			"back_urls":        map[string]string{"success": back, "failure": back, "pending": back}, "auto_return": "approved",
			"expires": true, "expiration_date_to": purchase.ExpiresAt.Format(time.RFC3339),
			"payment_methods": map[string]any{"excluded_payment_types": []any{map[string]string{"id": "ticket"}, map[string]string{"id": "atm"}}, "installments": 1},
		}
		var pref services.MPPreference
		mp := s.mercadoPago()
		if err := mp.Request(c.Request.Context(), http.MethodPost, "/checkout/preferences", payload, &pref); err != nil {
			return err
		}
		purchase.PreferenceID = pref.ID
		purchase.CheckoutURL = pref.InitPoint
		if s.config.MPTestMode {
			purchase.CheckoutURL = pref.SandboxInitPoint
		}
		target, err := url.Parse(purchase.CheckoutURL)
		if err != nil || target.Scheme != "https" || !(target.Hostname() == "www.mercadopago.com.ar" || target.Hostname() == "sandbox.mercadopago.com.ar") || pref.ID == "" {
			return errors.New("invalid checkout URL")
		}
		return tx.Save(&purchase).Error
	})
	if err != nil {
		utils.ServerError(c, "No se pudo abrir Mercado Pago. La reserva vence en 30 minutos. Si ya pagaste, no vuelvas a pagar y contactá a MendoCultura.")
		return
	}
	c.JSON(http.StatusCreated, gin.H{"purchase": purchase, "checkoutUrl": purchase.CheckoutURL})
}

func (s *Server) GetMyPurchase(c *gin.Context) {
	user, _ := middleware.CurrentUser(c)
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	var purchase models.Purchase
	if err := s.db.Where("id = ? AND user_id = ?", id, user.ID).First(&purchase).Error; err != nil {
		utils.NotFound(c, "Compra inexistente.")
		return
	}
	// Return only authoritative local status, never trust redirect query parameters.
	c.JSON(http.StatusOK, gin.H{"purchase": purchase})
}

func (s *Server) PaymentWebhook(c *gin.Context) {
	if !s.paymentReady() {
		c.Status(http.StatusServiceUnavailable)
		return
	}
	id := c.Query("data.id")
	if !services.ValidMPSignature(c.GetHeader("x-signature"), c.GetHeader("x-request-id"), id, s.config.MPWebhookSecret) {
		c.Status(http.StatusUnauthorized)
		return
	}
	if _, err := strconv.ParseUint(id, 10, 64); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}
	var notice struct {
		Type string `json:"type"`
	}
	if c.ShouldBindJSON(&notice) != nil {
		c.Status(http.StatusBadRequest)
		return
	}
	if notice.Type != "payment" {
		c.Status(http.StatusOK)
		return
	}
	var payment services.MPPayment
	mp := s.mercadoPago()
	if err := mp.Request(c.Request.Context(), http.MethodGet, "/v1/payments/"+id, nil, &payment); err != nil {
		c.Status(http.StatusBadGateway)
		return
	}
	if fmt.Sprint(payment.ID) != id || s.validatePaymentEnvironment(c, payment) != nil {
		c.Status(http.StatusBadRequest)
		return
	}
	var purchase models.Purchase
	if err := s.db.Where("payment_reference = ? AND payment_reference <> ''", payment.ExternalReference).First(&purchase).Error; errors.Is(err, gorm.ErrRecordNotFound) {
		c.Status(http.StatusOK)
		return
	} else if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}
	cents, err := services.AmountCents(payment.TransactionAmount)
	if err != nil || cents != purchase.TotalCents || payment.CurrencyID != "ARS" {
		c.Status(http.StatusBadRequest)
		return
	}
	err = s.applyPayment(&purchase, payment)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}
	c.Status(http.StatusOK)
}

func (s *Server) validatePaymentEnvironment(c *gin.Context, payment services.MPPayment) error {
	if fmt.Sprint(payment.CollectorID) != s.config.MPCollectorID {
		return errors.New("wrong collector")
	}
	var seller services.MPUser
	if s.config.MPTestMode && payment.LiveMode {
		mp := s.mercadoPago()
		if err := mp.Request(c.Request.Context(), http.MethodGet, "/users/me", nil, &seller); err != nil {
			return err
		}
	}
	if !services.ValidMPEnvironment(s.config.MPTestMode, payment.LiveMode, payment.CollectorID, seller) {
		return errors.New("wrong payment environment")
	}
	return nil
}

// Reconcile an existing purchase using the provider API. No money is moved.
// The browser supplies only its own purchase ID; reference, amount and seller
// are taken from the database and verified independently for every result.
func (s *Server) SyncMyPurchase(c *gin.Context) {
	if !s.paymentReady() {
		c.JSON(http.StatusServiceUnavailable, gin.H{"message": "Mercado Pago todavía no está configurado."})
		return
	}
	user, _ := middleware.CurrentUser(c)
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	var purchase models.Purchase
	if err := s.db.Where("id = ? AND user_id = ?", id, user.ID).First(&purchase).Error; err != nil {
		utils.NotFound(c, "Compra inexistente.")
		return
	}
	if purchase.PaymentReference == "" {
		utils.Conflict(c, "Esta compra no tiene una referencia de Mercado Pago.")
		return
	}
	mp := s.mercadoPago()
	var search struct {
		Results []struct {
			ID int64 `json:"id"`
		} `json:"results"`
	}
	path := "/v1/payments/search?external_reference=" + url.QueryEscape(purchase.PaymentReference) + "&sort=date_created&criteria=desc&limit=50"
	if err := mp.Request(c.Request.Context(), http.MethodGet, path, nil, &search); err != nil {
		utils.ServerError(c, "No se pudo consultar Mercado Pago. Reintentá en unos segundos.")
		return
	}
	for _, candidate := range search.Results {
		var payment services.MPPayment
		if err := mp.Request(c.Request.Context(), http.MethodGet, fmt.Sprintf("/v1/payments/%d", candidate.ID), nil, &payment); err != nil {
			utils.ServerError(c, "No se pudo verificar el pago en Mercado Pago.")
			return
		}
		if payment.ID != candidate.ID || payment.ExternalReference != purchase.PaymentReference {
			utils.Conflict(c, "El pago no corresponde a esta compra.")
			return
		}
		if err := s.validatePaymentEnvironment(c, payment); err != nil {
			utils.Conflict(c, "La cuenta o el ambiente del pago no coinciden con la configuración.")
			return
		}
		cents, err := services.AmountCents(payment.TransactionAmount)
		if err != nil || cents != purchase.TotalCents || payment.CurrencyID != "ARS" {
			utils.Conflict(c, "El importe o la moneda del pago no coinciden con la compra.")
			return
		}
		if err := s.applyPayment(&purchase, payment); err != nil {
			utils.ServerError(c, "No se pudo actualizar la compra.")
			return
		}
	}
	c.JSON(http.StatusOK, gin.H{"purchase": purchase})
}

func (s *Server) applyPayment(purchase *models.Purchase, payment services.MPPayment) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var event models.Event
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&event, purchase.EventID).Error; err != nil {
			return err
		}
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(purchase, purchase.ID).Error; err != nil {
			return err
		}
		id := fmt.Sprint(payment.ID)
		if purchase.PaymentID != "" && purchase.PaymentID != id {
			if payment.Status == "approved" {
				return tx.Create(&models.AuditLog{Action: "DUPLICATE_PAYMENT_REVIEW", Entity: "purchases", EntityID: fmt.Sprint(purchase.ID), Detail: "Pago adicional " + id + "; revisar devolución manual en Mercado Pago."}).Error
			}
			return nil
		}
		refunded := int64(0)
		if payment.AmountRefunded != "" {
			var err error
			refunded, err = services.AmountCents(payment.AmountRefunded)
			if err != nil {
				return err
			}
		}
		if payment.Status == "refunded" || payment.Status == "charged_back" || refunded > 0 {
			if purchase.Status == models.PurchasePending {
				event.AvailableTickets += purchase.Quantity
				if err := tx.Save(&event).Error; err != nil {
					return err
				}
			}
			purchase.PaymentID = id
			purchase.Status = "REFUNDED"
			if err := tx.Model(&models.Ticket{}).Where("purchase_id = ?", purchase.ID).Update("status", models.TicketCancelled).Error; err != nil {
				return err
			}
			return tx.Save(purchase).Error
		}
		if payment.Status != "approved" || purchase.Status == models.PurchasePaid || purchase.Status == "REFUNDED" || purchase.Status == "PAYMENT_REVIEW" {
			return nil
		}
		held := purchase.Status == models.PurchasePending
		purchase.PaymentID = id
		if event.Status != models.EventPublished || !event.StartAt.After(time.Now()) || (!held && event.AvailableTickets < purchase.Quantity) {
			if held {
				event.AvailableTickets += purchase.Quantity
				if err := tx.Save(&event).Error; err != nil {
					return err
				}
			}
			// Money received after expiry with no stock requires an operator refund.
			purchase.Status = "PAYMENT_REVIEW"
			if err := tx.Create(&models.AuditLog{Action: "PAYMENT_REVIEW", Entity: "purchases", EntityID: fmt.Sprint(purchase.ID), Detail: "Pago aprobado sin disponibilidad; requiere devolución manual en Mercado Pago."}).Error; err != nil {
				return err
			}
			return tx.Save(purchase).Error
		}
		if !held {
			event.AvailableTickets -= purchase.Quantity
			if err := tx.Save(&event).Error; err != nil {
				return err
			}
		}
		for i := 0; i < purchase.Quantity; i++ {
			random := make([]byte, 24)
			if _, err := rand.Read(random); err != nil {
				return err
			}
			ticket := models.Ticket{UserID: purchase.UserID, EventID: event.ID, PurchaseID: purchase.ID, Status: models.TicketValid, Code: "pending-" + hex.EncodeToString(random)}
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
		}
		purchase.Status = models.PurchasePaid
		return tx.Save(purchase).Error
	})
}
