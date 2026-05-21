package domain

import (
	"time"
)

const (
	RoleUser      = "USER"
	RoleOrganizer = "ORGANIZER"
	RoleValidator = "VALIDATOR"
	RoleAdmin     = "ADMIN"
)

const (
	AccountActive    = "ACTIVE"
	AccountPending   = "PENDING"
	AccountRejected  = "REJECTED"
	AccountSuspended = "SUSPENDED"
)

const (
	EventDraft     = "DRAFT"
	EventPublished = "PUBLISHED"
	EventPaused    = "PAUSED"
	EventCancelled = "CANCELLED"
	EventFinished  = "FINISHED"
)

const (
	PurchasePending   = "PENDING"
	PurchasePaid      = "PAID"
	PurchaseCancelled = "CANCELLED"
	PurchaseRejected  = "REJECTED"
)

const (
	TicketPending   = "PENDING"
	TicketValid     = "VALID"
	TicketUsed      = "USED"
	TicketCancelled = "CANCELLED"
	TicketExpired   = "EXPIRED"
)

type User struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	Name         string    `json:"name" gorm:"not null"`
	Email        string    `json:"email" gorm:"uniqueIndex;not null"`
	PasswordHash string    `json:"-" gorm:"not null"`
	DNI          string    `json:"dni" gorm:"index"`
	Role         string    `json:"role" gorm:"not null;default:USER"`
	Status       string    `json:"status" gorm:"not null;default:ACTIVE"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

type OrganizerProfile struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	UserID       uint      `json:"userId" gorm:"uniqueIndex;not null"`
	User         User      `json:"user" gorm:"constraint:OnDelete:CASCADE"`
	BusinessName string    `json:"businessName" gorm:"not null"`
	TaxID        string    `json:"taxId" gorm:"uniqueIndex;not null"`
	Locality     string    `json:"locality" gorm:"not null"`
	Status       string    `json:"status" gorm:"not null;default:PENDING"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

type Event struct {
	ID                  uint      `json:"id" gorm:"primaryKey"`
	OrganizerID         uint      `json:"organizerId" gorm:"index;not null"`
	Organizer           User      `json:"organizer" gorm:"constraint:OnDelete:RESTRICT"`
	Title               string    `json:"title" gorm:"not null"`
	Description         string    `json:"description" gorm:"type:text"`
	ExtendedDescription string    `json:"extendedDescription" gorm:"type:text"`
	Category            string    `json:"category" gorm:"index;not null"`
	Department          string    `json:"department" gorm:"index;not null"`
	Venue               string    `json:"venue" gorm:"not null"`
	Address             string    `json:"address" gorm:"type:text"`
	ImageURL            string    `json:"imageUrl" gorm:"type:text"`
	GalleryImages       string    `json:"galleryImages" gorm:"type:text"`
	MapURL              string    `json:"mapUrl" gorm:"type:text"`
	Latitude            *float64  `json:"latitude"`
	Longitude           *float64  `json:"longitude"`
	TicketType          string    `json:"ticketType" gorm:"not null;default:General"`
	ImportantInfo       string    `json:"importantInfo" gorm:"type:text"`
	Recommendations     string    `json:"recommendations" gorm:"type:text"`
	StartAt             time.Time `json:"startAt" gorm:"index;not null"`
	PriceCents          int64     `json:"priceCents" gorm:"not null"`
	Capacity            int       `json:"capacity" gorm:"not null"`
	AvailableTickets    int       `json:"availableTickets" gorm:"not null"`
	Status              string    `json:"status" gorm:"index;not null;default:DRAFT"`
	CreatedAt           time.Time `json:"createdAt"`
	UpdatedAt           time.Time `json:"updatedAt"`
}

type Purchase struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	UserID     uint      `json:"userId" gorm:"index;not null"`
	User       User      `json:"user" gorm:"constraint:OnDelete:RESTRICT"`
	EventID    uint      `json:"eventId" gorm:"index;not null"`
	Event      Event     `json:"event" gorm:"constraint:OnDelete:RESTRICT"`
	Quantity   int       `json:"quantity" gorm:"not null"`
	TotalCents int64     `json:"totalCents" gorm:"not null"`
	Status     string    `json:"status" gorm:"index;not null;default:PAID"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

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

type AuditLog struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	UserID    *uint     `json:"userId" gorm:"index"`
	Action    string    `json:"action" gorm:"index;not null"`
	Entity    string    `json:"entity" gorm:"index;not null"`
	EntityID  string    `json:"entityId"`
	Detail    string    `json:"detail" gorm:"type:text"`
	CreatedAt time.Time `json:"createdAt" gorm:"index"`
}
