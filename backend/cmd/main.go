package main

import (
	"log"

	"MendoCultura/config"
	"MendoCultura/database"
	"MendoCultura/routes"
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
		if err := database.Seed(db); err != nil {
			log.Fatalf("database seed failed: %v", err)
		}
	}

	router := routes.NewRouter(db, cfg)
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
