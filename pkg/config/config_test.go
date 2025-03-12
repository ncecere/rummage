package config

import (
	"os"
	"testing"
	"time"
)

func TestLoadConfig(t *testing.T) {
	// Save original environment variables to restore later
	origPort := os.Getenv("RUMMAGE_SERVER_PORT")
	origBaseURL := os.Getenv("RUMMAGE_SERVER_BASEURL")
	origRedisURL := os.Getenv("RUMMAGE_REDIS_URL")
	origDefaultTimeout := os.Getenv("RUMMAGE_SCRAPER_DEFAULTTIMEOUTMS")
	origDefaultWaitTime := os.Getenv("RUMMAGE_SCRAPER_DEFAULTWAITTIMEMS")
	origMaxConcurrentJobs := os.Getenv("RUMMAGE_SCRAPER_MAXCONCURRENTJOBS")
	origJobExpirationHours := os.Getenv("RUMMAGE_SCRAPER_JOBEXPIRATIONHOURS")

	// Restore environment variables after the test
	defer func() {
		os.Setenv("RUMMAGE_SERVER_PORT", origPort)
		os.Setenv("RUMMAGE_SERVER_BASEURL", origBaseURL)
		os.Setenv("RUMMAGE_REDIS_URL", origRedisURL)
		os.Setenv("RUMMAGE_SCRAPER_DEFAULTTIMEOUTMS", origDefaultTimeout)
		os.Setenv("RUMMAGE_SCRAPER_DEFAULTWAITTIMEMS", origDefaultWaitTime)
		os.Setenv("RUMMAGE_SCRAPER_MAXCONCURRENTJOBS", origMaxConcurrentJobs)
		os.Setenv("RUMMAGE_SCRAPER_JOBEXPIRATIONHOURS", origJobExpirationHours)
	}()

	// Test with default values
	t.Run("Default values", func(t *testing.T) {
		// Clear environment variables
		os.Unsetenv("RUMMAGE_SERVER_PORT")
		os.Unsetenv("RUMMAGE_SERVER_BASEURL")
		os.Unsetenv("RUMMAGE_REDIS_URL")
		os.Unsetenv("RUMMAGE_SCRAPER_DEFAULTTIMEOUTMS")
		os.Unsetenv("RUMMAGE_SCRAPER_DEFAULTWAITTIMEMS")
		os.Unsetenv("RUMMAGE_SCRAPER_MAXCONCURRENTJOBS")
		os.Unsetenv("RUMMAGE_SCRAPER_JOBEXPIRATIONHOURS")

		cfg, err := LoadConfig()
		if err != nil {
			t.Fatalf("Failed to load config: %v", err)
		}

		// Check default values
		if cfg.Port != "8080" {
			t.Errorf("Expected default Port to be '8080', got '%s'", cfg.Port)
		}
		if cfg.BaseURL != "http://localhost:8080" {
			t.Errorf("Expected default BaseURL to be 'http://localhost:8080', got '%s'", cfg.BaseURL)
		}
		if cfg.RedisURL != "redis://localhost:6379" {
			t.Errorf("Expected default RedisURL to be 'redis://localhost:6379', got '%s'", cfg.RedisURL)
		}
		if cfg.DefaultTimeout != 30000*time.Millisecond {
			t.Errorf("Expected default DefaultTimeout to be 30000ms, got '%v'", cfg.DefaultTimeout)
		}
		if cfg.DefaultWaitTime != 0 {
			t.Errorf("Expected default DefaultWaitTime to be 0, got '%v'", cfg.DefaultWaitTime)
		}
		if cfg.MaxConcurrentJobs != 10 {
			t.Errorf("Expected default MaxConcurrentJobs to be 10, got '%d'", cfg.MaxConcurrentJobs)
		}
		if cfg.JobExpirationHours != 24 {
			t.Errorf("Expected default JobExpirationHours to be 24, got '%d'", cfg.JobExpirationHours)
		}
	})

	// Test with custom values
	t.Run("Custom values", func(t *testing.T) {
		// Set environment variables
		os.Setenv("RUMMAGE_SERVER_PORT", "3000")
		os.Setenv("RUMMAGE_SERVER_BASEURL", "https://example.com")
		os.Setenv("RUMMAGE_REDIS_URL", "redis://redis:6379")
		os.Setenv("RUMMAGE_SCRAPER_DEFAULTTIMEOUTMS", "5000")
		os.Setenv("RUMMAGE_SCRAPER_DEFAULTWAITTIMEMS", "1000")
		os.Setenv("RUMMAGE_SCRAPER_MAXCONCURRENTJOBS", "5")
		os.Setenv("RUMMAGE_SCRAPER_JOBEXPIRATIONHOURS", "48")

		cfg, err := LoadConfig()
		if err != nil {
			t.Fatalf("Failed to load config: %v", err)
		}

		// Check custom values
		if cfg.Port != "3000" {
			t.Errorf("Expected Port to be '3000', got '%s'", cfg.Port)
		}
		if cfg.BaseURL != "https://example.com" {
			t.Errorf("Expected BaseURL to be 'https://example.com', got '%s'", cfg.BaseURL)
		}
		if cfg.RedisURL != "redis://redis:6379" {
			t.Errorf("Expected RedisURL to be 'redis://redis:6379', got '%s'", cfg.RedisURL)
		}
		if cfg.DefaultTimeout != 5000*time.Millisecond {
			t.Errorf("Expected DefaultTimeout to be 5000ms, got '%v'", cfg.DefaultTimeout)
		}
		if cfg.DefaultWaitTime != 1000*time.Millisecond {
			t.Errorf("Expected DefaultWaitTime to be 1000ms, got '%v'", cfg.DefaultWaitTime)
		}
		if cfg.MaxConcurrentJobs != 5 {
			t.Errorf("Expected MaxConcurrentJobs to be 5, got '%d'", cfg.MaxConcurrentJobs)
		}
		if cfg.JobExpirationHours != 48 {
			t.Errorf("Expected JobExpirationHours to be 48, got '%d'", cfg.JobExpirationHours)
		}
	})

	// Test with invalid values
	t.Run("Invalid values", func(t *testing.T) {
		// Set environment variables with invalid values
		os.Setenv("RUMMAGE_SCRAPER_DEFAULTTIMEOUTMS", "invalid")
		os.Setenv("RUMMAGE_SCRAPER_MAXCONCURRENTJOBS", "invalid")

		cfg, err := LoadConfig()
		if err != nil {
			t.Fatalf("Failed to load config: %v", err)
		}

		// Check that default values are used for invalid inputs
		if cfg.DefaultTimeout != 30000*time.Millisecond {
			t.Errorf("Expected DefaultTimeout to fall back to default 30000ms for invalid input, got '%v'", cfg.DefaultTimeout)
		}
		if cfg.MaxConcurrentJobs != 10 {
			t.Errorf("Expected MaxConcurrentJobs to fall back to default 10 for invalid input, got '%d'", cfg.MaxConcurrentJobs)
		}
	})
}
