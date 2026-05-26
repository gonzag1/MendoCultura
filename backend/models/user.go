package models

import "time"

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
