package config

import (
	"os"
	"strconv"
	"strings"
)

// Config holds all configuration values for the application
type Config struct {
	// Application settings
	AppEnv      string
	Port        string
	Host        string
	Debug       bool
	LogLevel    string
	JWTSecret   string
	CORSOrigins []string

	// Database settings
	DatabaseURL string

	// Redis settings
	RedisURL string

	// NFL API settings
	MySportsAPIKey string
	NFLSeasonYear  int

	// Background job settings
	EnableBackgroundJobs bool
	ScoreUpdateInterval  int // seconds

	// Chat settings
	ChatMessageLimit int
	ChatRateLimit    int // messages per minute per user
}

// Load creates a new Config instance with values from environment variables
func Load() *Config {
	return &Config{
		AppEnv:      getEnv("APP_ENV", "development"),
		Port:        getEnv("APP_PORT", "8080"),
		Host:        getEnv("APP_HOST", "0.0.0.0"),
		Debug:       getEnvBool("DEBUG", true),
		LogLevel:    getEnv("LOG_LEVEL", "debug"),
		JWTSecret:   getEnv("JWT_SECRET", "your-jwt-secret-key-change-this-in-production"),
		CORSOrigins: getEnvStringSlice("CORS_ORIGINS", []string{"http://localhost:3000", "http://127.0.0.1:3000"}),

		DatabaseURL: getEnv("DATABASE_URL", "postgres://touchdown_user:touchdown_dev@postgres:5432/touchdown_tally?sslmode=disable"),

		RedisURL: getEnv("REDIS_URL", "redis://:touchdown_redis@redis:6379/0"),

		MySportsAPIKey: getEnv("MYSPORTSFEEDS_API_KEY", ""),
		NFLSeasonYear:  getEnvInt("NFL_SEASON_YEAR", 2024),

		EnableBackgroundJobs: getEnvBool("ENABLE_BACKGROUND_JOBS", true),
		ScoreUpdateInterval:  getEnvInt("SCORE_UPDATE_INTERVAL", 300), // 5 minutes

		ChatMessageLimit: getEnvInt("CHAT_MESSAGE_LIMIT", 500),
		ChatRateLimit:    getEnvInt("CHAT_RATE_LIMIT", 10),
	}
}

// Helper functions for environment variable parsing

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		parsed, err := strconv.ParseBool(value)
		if err != nil {
			return defaultValue
		}
		return parsed
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil {
			return defaultValue
		}
		return parsed
	}
	return defaultValue
}

func getEnvStringSlice(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		return strings.Split(value, ",")
	}
	return defaultValue
}
