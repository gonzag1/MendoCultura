package database

import (
	"MendoCultura/internal/domain"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Open(databaseURL string) (*gorm.DB, error) {
	return gorm.Open(postgres.Open(databaseURL), &gorm.Config{})
}

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&domain.User{},
		&domain.OrganizerProfile{},
		&domain.Event{},
		&domain.Purchase{},
		&domain.Ticket{},
		&domain.AuditLog{},
	)
}
