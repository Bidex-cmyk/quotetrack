package config

import "os"

// Config holds all runtime configuration for the server.
type Config struct {
	Port            string
	DatabaseURL     string
	JWTSecret       string
	JWTExpiry       string
	FrontendOrigin  string
	ShutdownTimeout int
}

// Load reads configuration from environment variables, applying defaults.
func Load() *Config {
	return &Config{
		Port:            getenv("PORT", "8080"),
		DatabaseURL:     getenv("DATABASE_URL", "postgres://quotetrack:quotetrack@localhost:5432/quotetrack?sslmode=disable"),
		JWTSecret:       getenv("JWT_SECRET", "dev-only-secret-change-me"),
		JWTExpiry:       getenv("JWT_EXPIRY", "168h"),
		FrontendOrigin:  getenv("FRONTEND_ORIGIN", "http://localhost:5173"),
		ShutdownTimeout: 10,
	}
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}