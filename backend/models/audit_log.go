package models

import "time"

type AuditLog struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	UserID    *uint     `json:"userId" gorm:"index"`
	Action    string    `json:"action" gorm:"index;not null"`
	Entity    string    `json:"entity" gorm:"index;not null"`
	EntityID  string    `json:"entityId"`
	Detail    string    `json:"detail" gorm:"type:text"`
	CreatedAt time.Time `json:"createdAt" gorm:"index"`
}
