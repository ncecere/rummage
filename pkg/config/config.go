// Package config provides configuration functionality for the application.
package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
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

// LoadConfig loads the configuration from environment variables and config files.
func LoadConfig() (*Config, error) {
	v := viper.New()

	// Set default values
	v.SetDefault("server.port", "8080")
	v.SetDefault("server.baseURL", "")
	v.SetDefault("redis.url", "redis://localhost:6379")
	v.SetDefault("scraper.defaultTimeoutMS", 30000)
	v.SetDefault("scraper.defaultWaitTimeMS", 0)
	v.SetDefault("scraper.maxConcurrentJobs", 10)
	v.SetDefault("scraper.jobExpirationHours", 24)

	// Set environment variable prefix and bind environment variables
	v.SetEnvPrefix("RUMMAGE")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Read config file
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	v.AddConfigPath("./config")
	v.AddConfigPath("/etc/rummage")
	v.AddConfigPath("$HOME/.rummage")

	// Read config file if it exists
	if err := v.ReadInConfig(); err != nil {
		// It's okay if config file doesn't exist
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	// Create config struct with default values
	cfg := &Config{
		// Server configuration
		Port:    v.GetString("server.port"),
		BaseURL: v.GetString("server.baseURL"),

		// Redis configuration
		RedisURL: v.GetString("redis.url"),

		// Scraper configuration
		DefaultTimeout:     time.Duration(getIntWithDefault(v, "scraper.defaultTimeoutMS", 30000)) * time.Millisecond,
		DefaultWaitTime:    time.Duration(getIntWithDefault(v, "scraper.defaultWaitTimeMS", 0)) * time.Millisecond,
		MaxConcurrentJobs:  getIntWithDefault(v, "scraper.maxConcurrentJobs", 10),
		JobExpirationHours: getIntWithDefault(v, "scraper.jobExpirationHours", 24),
	}

	// If BaseURL is not set, derive it from Port
	if cfg.BaseURL == "" {
		cfg.BaseURL = "http://localhost:" + cfg.Port
	}

	return cfg, nil
}

// getIntWithDefault gets an integer value from Viper, falling back to the provided default if the value is invalid.
func getIntWithDefault(v *viper.Viper, key string, defaultValue int) int {
	// Check if the key exists and is a valid integer
	if v.IsSet(key) {
		// Try to get the value as a string first
		strValue := v.GetString(key)
		// If it's not a valid integer, return the default
		if _, err := fmt.Sscanf(strValue, "%d", &defaultValue); err != nil {
			return defaultValue
		}
	}

	// Get the value as an int
	value := v.GetInt(key)
	if value == 0 && v.GetString(key) != "0" {
		// If the value is 0 but the string representation is not "0", it's likely invalid
		return defaultValue
	}

	return value
}
