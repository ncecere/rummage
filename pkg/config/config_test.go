package config

import (
	"os"
	"testing"
	"time"
)

func TestLoadConfig(t *testing.T) {
	// Save original environment variables to restore later
	origPort := os.Getenv("PORT")
	origBaseURL := os.Getenv("BASE_URL")
	origRedisURL := os.Getenv("REDIS_URL")
	origDefaultTimeout := os.Getenv("DEFAULT_TIMEOUT_MS")
	origDefaultWaitTime := os.Getenv("DEFAULT_WAIT_TIME_MS")
	origMaxConcurrentJobs := os.Getenv("MAX_CONCURRENT_JOBS")
	origJobExpirationHours := os.Getenv("JOB_EXPIRATION_HOURS")

	// Restore environment variables after the test
	defer func() {
		os.Setenv("PORT", origPort)
		os.Setenv("BASE_URL", origBaseURL)
		os.Setenv("REDIS_URL", origRedisURL)
		os.Setenv("DEFAULT_TIMEOUT_MS", origDefaultTimeout)
		os.Setenv("DEFAULT_WAIT_TIME_MS", origDefaultWaitTime)
		os.Setenv("MAX_CONCURRENT_JOBS", origMaxConcurrentJobs)
		os.Setenv("JOB_EXPIRATION_HOURS", origJobExpirationHours)
	}()

	// Test with default values
	t.Run("Default values", func(t *testing.T) {
		// Clear environment variables
		os.Unsetenv("PORT")
		os.Unsetenv("BASE_URL")
		os.Unsetenv("REDIS_URL")
		os.Unsetenv("DEFAULT_TIMEOUT_MS")
		os.Unsetenv("DEFAULT_WAIT_TIME_MS")
		os.Unsetenv("MAX_CONCURRENT_JOBS")
		os.Unsetenv("JOB_EXPIRATION_HOURS")

		cfg := LoadConfig()

		// Check default values
		if cfg.Port != "8080" {
			t.Errorf("Expected default Port to be '8080', got '%s'", cfg.Port)
		}
		if cfg.BaseURL != "http://localhost:8080" {
			t.Errorf("Expected default BaseURL to be 'http://localhost:8080', got '%s'", cfg.BaseURL)
		}
		if cfg.RedisURL != "localhost:6379" {
			t.Errorf("Expected default RedisURL to be 'localhost:6379', got '%s'", cfg.RedisURL)
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
		os.Setenv("PORT", "3000")
		os.Setenv("BASE_URL", "https://example.com")
		os.Setenv("REDIS_URL", "redis:6379")
		os.Setenv("DEFAULT_TIMEOUT_MS", "5000")
		os.Setenv("DEFAULT_WAIT_TIME_MS", "1000")
		os.Setenv("MAX_CONCURRENT_JOBS", "5")
		os.Setenv("JOB_EXPIRATION_HOURS", "48")

		cfg := LoadConfig()

		// Check custom values
		if cfg.Port != "3000" {
			t.Errorf("Expected Port to be '3000', got '%s'", cfg.Port)
		}
		if cfg.BaseURL != "https://example.com" {
			t.Errorf("Expected BaseURL to be 'https://example.com', got '%s'", cfg.BaseURL)
		}
		if cfg.RedisURL != "redis:6379" {
			t.Errorf("Expected RedisURL to be 'redis:6379', got '%s'", cfg.RedisURL)
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
		os.Setenv("DEFAULT_TIMEOUT_MS", "invalid")
		os.Setenv("MAX_CONCURRENT_JOBS", "invalid")

		cfg := LoadConfig()

		// Check that default values are used for invalid inputs
		if cfg.DefaultTimeout != 30000*time.Millisecond {
			t.Errorf("Expected DefaultTimeout to fall back to default 30000ms for invalid input, got '%v'", cfg.DefaultTimeout)
		}
		if cfg.MaxConcurrentJobs != 10 {
			t.Errorf("Expected MaxConcurrentJobs to fall back to default 10 for invalid input, got '%d'", cfg.MaxConcurrentJobs)
		}
	})
}

func TestGetEnvAsBool(t *testing.T) {
	// Save original environment variable to restore later
	origValue := os.Getenv("TEST_BOOL")
	defer os.Setenv("TEST_BOOL", origValue)

	// Test with true value
	os.Setenv("TEST_BOOL", "true")
	if !getEnvAsBool("TEST_BOOL", false) {
		t.Errorf("Expected getEnvAsBool to return true for 'true' value")
	}

	// Test with false value
	os.Setenv("TEST_BOOL", "false")
	if getEnvAsBool("TEST_BOOL", true) {
		t.Errorf("Expected getEnvAsBool to return false for 'false' value")
	}

	// Test with invalid value
	os.Setenv("TEST_BOOL", "invalid")
	if !getEnvAsBool("TEST_BOOL", true) {
		t.Errorf("Expected getEnvAsBool to return default value (true) for invalid input")
	}

	// Test with missing value
	os.Unsetenv("TEST_BOOL")
	if !getEnvAsBool("TEST_BOOL", true) {
		t.Errorf("Expected getEnvAsBool to return default value (true) for missing env var")
	}
}
