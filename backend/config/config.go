package config

import "os"

type Config struct {
	Port        string
	DatabaseURL string
	JWTSecret   string
	SeedData    bool
}

func Load() Config {
	return Config{
		Port:        getEnv("PORT", "8080"),
		DatabaseURL: getEnv("DATABASE_URL", "postgres://mendocultura:mendocultura@localhost:5432/mendocultura?sslmode=disable"),
		JWTSecret:   getEnv("JWT_SECRET", "change-this-secret-in-production"),
		SeedData:    getEnv("SEED_DATA", "true") == "true",
	}
}

func getEnv(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
