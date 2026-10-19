package models

import "time"

const (
	PurchasePending   = "PENDING"
	PurchasePaid      = "PAID"
	PurchaseCancelled = "CANCELLED"
	PurchaseRejected  = "REJECTED"
)

type Purchase struct {
	PreferenceAttempted bool       `json:"-"`
	PaymentReference    string     `json:"-" gorm:"index"`
	PreferenceID        string     `json:"preferenceId"`
	CheckoutURL         string     `json:"checkoutUrl"`
	PaymentID           string     `json:"paymentId" gorm:"index"`
	ExpiresAt           *time.Time `json:"expiresAt" gorm:"index"`
	ID                  uint       `json:"id" gorm:"primaryKey"`
	UserID              uint       `json:"userId" gorm:"index;not null"`
	User                User       `json:"user" gorm:"constraint:OnDelete:RESTRICT"`
	EventID             uint       `json:"eventId" gorm:"index;not null"`
	Event               Event      `json:"event" gorm:"constraint:OnDelete:RESTRICT"`
	Quantity            int        `json:"quantity" gorm:"not null"`
	TotalCents          int64      `json:"totalCents" gorm:"not null"`
	Status              string     `json:"status" gorm:"index;not null;default:PAID"`
	CreatedAt           time.Time  `json:"createdAt"`
	UpdatedAt           time.Time  `json:"updatedAt"`
}
