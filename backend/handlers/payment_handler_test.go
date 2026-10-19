package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"MendoCultura/config"
	"MendoCultura/models"
	"MendoCultura/services"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func paymentDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL is required for isolated PostgreSQL tests")
	}
	base, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	schema := fmt.Sprintf("payment_test_%d", time.Now().UnixNano())
	if err := base.Exec("CREATE SCHEMA " + schema).Error; err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { base.Exec("DROP SCHEMA " + schema + " CASCADE"); sqlDB, _ := base.DB(); sqlDB.Close() })
	parsed, err := url.Parse(dsn)
	if err != nil {
		t.Fatal(err)
	}
	query := parsed.Query()
	query.Set("search_path", schema)
	parsed.RawQuery = query.Encode()
	db, err := gorm.Open(postgres.Open(parsed.String()), &gorm.Config{DisableForeignKeyConstraintWhenMigrating: true})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { sqlDB, _ := db.DB(); sqlDB.Close() })
	if err := db.AutoMigrate(&models.User{}, &models.Event{}, &models.Purchase{}, &models.Ticket{}, &models.AuditLog{}); err != nil {
		t.Fatal(err)
	}
	return db
}
func paymentFixture(t *testing.T, db *gorm.DB, status string, available int) (models.Event, models.Purchase) {
	t.Helper()
	event := models.Event{OrganizerID: 1, Title: "Test", Category: "Music", Department: "Mendoza", Venue: "Test", StartAt: time.Now().Add(time.Hour), PriceCents: 10000, Capacity: 4, AvailableTickets: available, Status: models.EventPublished}
	if err := db.Create(&event).Error; err != nil {
		t.Fatal(err)
	}
	expires := time.Now().Add(-time.Minute)
	purchase := models.Purchase{UserID: 1, EventID: event.ID, Quantity: 2, TotalCents: 20000, Status: status, ExpiresAt: &expires, PaymentReference: fmt.Sprint(event.ID)}
	if err := db.Create(&purchase).Error; err != nil {
		t.Fatal(err)
	}
	return event, purchase
}
func approvedPayment() services.MPPayment {
	return services.MPPayment{ID: 123, Status: "approved", TransactionAmount: json.Number("200.00"), CurrencyID: "ARS"}
}

func TestPaymentDuplicateNotifications(t *testing.T) {
	db := paymentDB(t)
	event, purchase := paymentFixture(t, db, models.PurchasePending, 2)
	server := NewServer(db, config.Config{JWTSecret: "test"})
	var wg sync.WaitGroup
	errs := make(chan error, 8)
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); copy := purchase; errs <- server.applyPayment(&copy, approvedPayment()) }()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
	var count int64
	db.Model(&models.Ticket{}).Where("purchase_id = ?", purchase.ID).Count(&count)
	db.First(&event, event.ID)
	db.First(&purchase, purchase.ID)
	if count != 2 || event.AvailableTickets != 2 || purchase.Status != models.PurchasePaid {
		t.Fatalf("duplicate processing: tickets=%d available=%d status=%s", count, event.AvailableTickets, purchase.Status)
	}
	refund := approvedPayment()
	refund.Status = "refunded"
	if err := server.applyPayment(&purchase, refund); err != nil {
		t.Fatal(err)
	}
	if err := server.applyPayment(&purchase, approvedPayment()); err != nil {
		t.Fatal(err)
	}
	db.Model(&models.Ticket{}).Where("purchase_id = ? AND status = ?", purchase.ID, models.TicketValid).Count(&count)
	if count != 0 || purchase.Status != "REFUNDED" {
		t.Fatal("refund failed to invalidate tickets or stale approval revived them")
	}
}

func TestLatePaymentStock(t *testing.T) {
	db := paymentDB(t)
	server := NewServer(db, config.Config{JWTSecret: "test"})
	for _, available := range []int{0, 2} {
		event, purchase := paymentFixture(t, db, "EXPIRED", available)
		if err := server.applyPayment(&purchase, approvedPayment()); err != nil {
			t.Fatal(err)
		}
		db.First(&event, event.ID)
		expected := "PAYMENT_REVIEW"
		if available == 2 {
			expected = models.PurchasePaid
		}
		if purchase.Status != expected || event.AvailableTickets != 0 {
			t.Fatalf("late payment: %+v %+v", purchase, event)
		}
	}
}

func TestExpiredHoldReleasedOnce(t *testing.T) {
	db := paymentDB(t)
	event, _ := paymentFixture(t, db, models.PurchasePending, 2)
	for i := 0; i < 2; i++ {
		if err := db.Transaction(func(tx *gorm.DB) error {
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&event, event.ID).Error; err != nil {
				return err
			}
			return releaseExpired(tx, &event)
		}); err != nil {
			t.Fatal(err)
		}
	}
	db.First(&event, event.ID)
	if event.AvailableTickets != 4 {
		t.Fatalf("stock released more than once: %d", event.AvailableTickets)
	}
}

func TestLateConcurrentPaymentsDoNotOversell(t *testing.T) {
	db := paymentDB(t)
	event, first := paymentFixture(t, db, "EXPIRED", 2)
	second := first
	second.ID = 0
	second.PaymentReference = "other"
	if err := db.Create(&second).Error; err != nil {
		t.Fatal(err)
	}
	server := NewServer(db, config.Config{JWTSecret: "test"})
	var wg sync.WaitGroup
	errs := make(chan error, 2)
	for _, p := range []models.Purchase{first, second} {
		wg.Add(1)
		go func(p models.Purchase) { defer wg.Done(); errs <- server.applyPayment(&p, approvedPayment()) }(p)
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
	var tickets, reviews int64
	db.Model(&models.Ticket{}).Where("event_id = ?", event.ID).Count(&tickets)
	db.Model(&models.Purchase{}).Where("event_id = ? AND status = ?", event.ID, "PAYMENT_REVIEW").Count(&reviews)
	db.First(&event, event.ID)
	if tickets != 2 || reviews != 1 || event.AvailableTickets != 0 {
		t.Fatalf("overselling: tickets=%d reviews=%d available=%d", tickets, reviews, event.AvailableTickets)
	}
}

func TestCheckoutReusesHoldWhenSoldOut(t *testing.T) {
	db := paymentDB(t)
	event, purchase := paymentFixture(t, db, models.PurchasePending, 0)
	expires := time.Now().Add(20 * time.Minute)
	if err := db.Model(&purchase).Updates(map[string]any{"expires_at": expires, "checkout_url": "https://sandbox.mercadopago.com.ar/checkout/test"}).Error; err != nil {
		t.Fatal(err)
	}
	server := NewServer(db, config.Config{MPAccessToken: "test", MPWebhookSecret: "test", MPCollectorID: "1", BackendURL: "https://backend.example.com", FrontendURL: "https://frontend.example.com"})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("currentUser", models.User{ID: 1})
	c.Request = httptest.NewRequest(http.MethodPost, "/payments/checkout", strings.NewReader(fmt.Sprintf(`{"eventId":%d,"quantity":1}`, event.ID)))
	c.Request.Header.Set("Content-Type", "application/json")
	server.CreateCheckout(c)
	if w.Code != http.StatusOK {
		t.Fatalf("could not resume sold-out hold: %d %s", w.Code, w.Body.String())
	}
	var count int64
	db.Model(&models.Purchase{}).Count(&count)
	if count != 1 {
		t.Fatal("resuming created another purchase")
	}
}

func TestCatalogReclaimsExpiredHolds(t *testing.T) {
	db := paymentDB(t)
	event, purchase := paymentFixture(t, db, models.PurchasePending, 0)
	server := NewServer(db, config.Config{})
	if err := server.expirePaymentHolds(); err != nil {
		t.Fatal(err)
	}
	if err := server.expirePaymentHolds(); err != nil {
		t.Fatal(err)
	}
	db.First(&event, event.ID)
	db.First(&purchase, purchase.ID)
	if event.AvailableTickets != 2 || purchase.Status != "EXPIRED" {
		t.Fatal("catalog did not reclaim abandoned stock exactly once")
	}
}

func TestWebhookRejectsUnsignedPayment(t *testing.T) {
	server := NewServer(nil, config.Config{MPAccessToken: "test", MPWebhookSecret: "test", MPCollectorID: "1", BackendURL: "https://backend.example.com", FrontendURL: "https://frontend.example.com"})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/payments/webhook?data.id=123", strings.NewReader(`{"type":"payment","data":{"id":"123"}}`))
	server.PaymentWebhook(c)
	if c.Writer.Status() != http.StatusUnauthorized {
		t.Fatal("unsigned webhook accepted")
	}
}

type paymentTransport func(*http.Request) (*http.Response, error)

func (f paymentTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestSyncMyPurchase(t *testing.T) {
	db := paymentDB(t)
	for _, tc := range []struct {
		name, reference, amount string
		owner                   uint
		want                    int
	}{
		{"approved", "match", "200.00", 1, 200},
		{"wrong reference", "other", "200.00", 1, 409},
		{"wrong amount", "match", "1.00", 1, 409},
		{"other owner", "match", "200.00", 99, 404},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, purchase := paymentFixture(t, db, models.PurchasePending, 2)
			server := NewServer(db, config.Config{JWTSecret: "test", MPAccessToken: "test", MPCollectorID: "42", MPTestMode: true, MPWebhookSecret: "test", BackendURL: "https://backend.example.com", FrontendURL: "https://frontend.example.com"})
			server.mpClient = &http.Client{Transport: paymentTransport(func(r *http.Request) (*http.Response, error) {
				body := ""
				switch r.URL.Path {
				case "/v1/payments/search":
					if r.URL.Query().Get("external_reference") != purchase.PaymentReference {
						t.Fatal("search did not use database reference")
					}
					body = `{"results":[{"id":123}]}`
				case "/v1/payments/123":
					ref := purchase.PaymentReference
					if tc.reference != "match" {
						ref = tc.reference
					}
					body = fmt.Sprintf(`{"id":123,"status":"approved","external_reference":%q,"collector_id":42,"live_mode":true,"currency_id":"ARS","transaction_amount":%s}`, ref, tc.amount)
				case "/users/me":
					body = `{"id":42,"tags":["test_user"]}`
				default:
					t.Fatalf("unexpected provider path: %s", r.URL.Path)
				}
				return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header)}, nil
			})}
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Set("currentUser", models.User{ID: tc.owner})
			c.Params = gin.Params{{Key: "id", Value: fmt.Sprint(purchase.ID)}}
			c.Request = httptest.NewRequest(http.MethodPost, "/purchases/1/sync", nil)
			server.SyncMyPurchase(c)
			if c.Writer.Status() != tc.want {
				t.Fatalf("status=%d body=%s", c.Writer.Status(), w.Body.String())
			}
			db.First(&purchase, purchase.ID)
			if tc.want == 200 && purchase.Status != models.PurchasePaid {
				t.Fatal("approved payment not recovered")
			}
			if tc.want != 200 && purchase.Status != models.PurchasePending {
				t.Fatal("unverified payment issued tickets")
			}
		})
	}
}
