package config

import (
	"os"
	"strconv"
)

// Config holds all runtime configuration for the server.
type Config struct {
	Port                   string
	DatabaseURL            string
	JWTSecret              string
	JWTExpiry              string
	FrontendOrigin         string
	FrontendURL            string
	RateLimitAuthPerMinute int
	RateLimitAuthBurst     int
	ShutdownTimeout        int
}

// Load reads configuration from environment variables, applying defaults.
func Load() *Config {
	return &Config{
		Port:                   getenv("PORT", "8080"),
		DatabaseURL:            getenv("DATABASE_URL", "postgres://quotetrack:quotetrack@localhost:5432/quotetrack?sslmode=disable"),
		JWTSecret:              getenv("JWT_SECRET", "dev-only-secret-change-me"),
		JWTExpiry:              getenv("JWT_EXPIRY", "168h"),
		FrontendOrigin:         getenv("FRONTEND_ORIGIN", "http://localhost:5173"),
		FrontendURL:            getenv("FRONTEND_URL", "http://localhost:5173"),
		RateLimitAuthPerMinute: getenvInt("RATE_LIMIT_AUTH_PER_MINUTE", 5),
		RateLimitAuthBurst:     getenvInt("RATE_LIMIT_AUTH_BURST", 5),
		ShutdownTimeout:        10,
	}
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// getenvInt parses an int env var, falling back on missing or invalid values.
func getenvInt(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil || n < 1 {
		return fallback
	}
	return n
}
