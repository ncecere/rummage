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

	// Create config struct
	cfg := &Config{
		// Server configuration
		Port:    v.GetString("server.port"),
		BaseURL: v.GetString("server.baseURL"),

		// Redis configuration
		RedisURL: v.GetString("redis.url"),

		// Scraper configuration
		DefaultTimeout:     time.Duration(v.GetInt("scraper.defaultTimeoutMS")) * time.Millisecond,
		DefaultWaitTime:    time.Duration(v.GetInt("scraper.defaultWaitTimeMS")) * time.Millisecond,
		MaxConcurrentJobs:  v.GetInt("scraper.maxConcurrentJobs"),
		JobExpirationHours: v.GetInt("scraper.jobExpirationHours"),
	}

	// If BaseURL is not set, derive it from Port
	if cfg.BaseURL == "" {
		cfg.BaseURL = "http://localhost:" + cfg.Port
	}

	return cfg, nil
}
