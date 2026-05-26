package models

import "time"

const (
	EventDraft     = "DRAFT"
	EventPublished = "PUBLISHED"
	EventPaused    = "PAUSED"
	EventCancelled = "CANCELLED"
	EventFinished  = "FINISHED"
)

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
