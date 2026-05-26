package database

import (
	"MendoCultura/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Open(databaseURL string) (*gorm.DB, error) {
	return gorm.Open(postgres.Open(databaseURL), &gorm.Config{})
}

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.User{},
		&models.OrganizerProfile{},
		&models.Event{},
		&models.Purchase{},
		&models.Ticket{},
		&models.AuditLog{},
	)
}
