package main

import (
	"log"

	"MendoCultura/internal/config"
	"MendoCultura/internal/database"
	"MendoCultura/internal/httpapi"
	"MendoCultura/internal/seed"
)

func main() {
	cfg := config.Load()

	db, err := database.Open(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database connection failed: %v", err)
	}

	if err := database.Migrate(db); err != nil {
		log.Fatalf("database migration failed: %v", err)
	}

	if cfg.SeedData {
		if err := seed.Run(db); err != nil {
			log.Fatalf("database seed failed: %v", err)
		}
	}

	router := httpapi.NewRouter(db, cfg)
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
