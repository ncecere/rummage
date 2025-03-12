// Package config provides configuration functionality for the application.
package config

import (
	"os"
	"strconv"
	"time"
)

// Config represents the application configuration.
type Config struct {
	// Server configuration
	Port    string
	BaseURL string

	// Redis configuration
	RedisURL string

	// Scraper configuration
	DefaultTimeout     time.Duration
	DefaultWaitTime    time.Duration
	MaxConcurrentJobs  int
	JobExpirationHours int
}

// LoadConfig loads the configuration from environment variables.
func LoadConfig() *Config {
	cfg := &Config{
		// Server configuration
		Port:    getEnv("PORT", "8080"),
		BaseURL: getEnv("BASE_URL", ""),

		// Redis configuration
		RedisURL: getEnv("REDIS_URL", "redis://localhost:6379"),

		// Scraper configuration
		DefaultTimeout:     time.Duration(getEnvAsInt("DEFAULT_TIMEOUT_MS", 30000)) * time.Millisecond,
		DefaultWaitTime:    time.Duration(getEnvAsInt("DEFAULT_WAIT_TIME_MS", 0)) * time.Millisecond,
		MaxConcurrentJobs:  getEnvAsInt("MAX_CONCURRENT_JOBS", 10),
		JobExpirationHours: getEnvAsInt("JOB_EXPIRATION_HOURS", 24),
	}

	// If BASE_URL is not set, derive it from PORT
	if cfg.BaseURL == "" {
		cfg.BaseURL = "http://localhost:" + cfg.Port
	}

	return cfg
}

// getEnv gets an environment variable or returns a default value.
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// getEnvAsInt gets an environment variable as an integer or returns a default value.
func getEnvAsInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}

	return value
}

// getEnvAsBool gets an environment variable as a boolean or returns a default value.
func getEnvAsBool(key string, defaultValue bool) bool {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.ParseBool(valueStr)
	if err != nil {
		return defaultValue
	}

	return value
}
