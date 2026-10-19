package config

import "os"

type Config struct {
	Port            string
	DatabaseURL     string
	JWTSecret       string
	SeedData        bool
	MPAccessToken   string
	MPWebhookSecret string
	MPCollectorID   string
	MPTestMode      bool
	FrontendURL     string
	BackendURL      string
}

func Load() Config {
	return Config{
		Port:            getEnv("PORT", "8080"),
		DatabaseURL:     getEnv("DATABASE_URL", "postgres://mendocultura:mendocultura@localhost:5432/mendocultura?sslmode=disable"),
		JWTSecret:       getEnv("JWT_SECRET", "change-this-secret-in-production"),
		SeedData:        getEnv("SEED_DATA", "true") == "true",
		MPAccessToken:   os.Getenv("MP_ACCESS_TOKEN"),
		MPWebhookSecret: os.Getenv("MP_WEBHOOK_SECRET"),
		MPCollectorID:   os.Getenv("MP_COLLECTOR_ID"),
		MPTestMode:      getEnv("MP_TEST_MODE", "true") == "true",
		FrontendURL:     getEnv("FRONTEND_URL", "http://localhost:4321"),
		BackendURL:      os.Getenv("BACKEND_URL"),
	}
}

func getEnv(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
