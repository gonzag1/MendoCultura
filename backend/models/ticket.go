package models

import "time"

const (
	TicketPending   = "PENDING"
	TicketValid     = "VALID"
	TicketUsed      = "USED"
	TicketCancelled = "CANCELLED"
	TicketExpired   = "EXPIRED"
)

type Ticket struct {
	ID         uint       `json:"id" gorm:"primaryKey"`
	UserID     uint       `json:"userId" gorm:"index;not null"`
	User       User       `json:"user" gorm:"constraint:OnDelete:RESTRICT"`
	EventID    uint       `json:"eventId" gorm:"index;not null"`
	Event      Event      `json:"event" gorm:"constraint:OnDelete:RESTRICT"`
	PurchaseID uint       `json:"purchaseId" gorm:"index;not null"`
	Purchase   Purchase   `json:"purchase" gorm:"constraint:OnDelete:RESTRICT"`
	Code       string     `json:"code" gorm:"uniqueIndex;not null"`
	Status     string     `json:"status" gorm:"index;not null;default:VALID"`
	UsedAt     *time.Time `json:"usedAt"`
	CreatedAt  time.Time  `json:"createdAt"`
	UpdatedAt  time.Time  `json:"updatedAt"`
}
