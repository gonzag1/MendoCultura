package models

import "time"

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
