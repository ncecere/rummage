// Package utils provides utility functions used across the application.
package utils

import (
	"net/url"
	"regexp"
	"strings"
)

// IsValidURL checks if a URL is valid.
func IsValidURL(rawURL string) bool {
	// Basic URL validation
	u, err := url.Parse(rawURL)
	if err != nil {
		return false
	}

	// Check if URL has a scheme and host
	return u.Scheme != "" && u.Host != ""
}

// NormalizeURL normalizes a URL by removing trailing slashes, fragments, etc.
func NormalizeURL(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}

	// Check if it's a valid URL with a host
	if u.Host == "" && !strings.Contains(rawURL, ".") {
		return rawURL // Return original for invalid URLs
	}

	// Remove fragment
	u.Fragment = ""

	// Ensure scheme is present for valid URLs
	if u.Scheme == "" {
		u.Scheme = "http"
	}

	// Remove trailing slash if present
	u.Path = strings.TrimSuffix(u.Path, "/")

	return u.String()
}

// ExtractDomain extracts the domain from a URL.
func ExtractDomain(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}

	return u.Hostname()
}

// IsRelativeURL checks if a URL is relative.
func IsRelativeURL(rawURL string) bool {
	return !strings.HasPrefix(rawURL, "http://") && !strings.HasPrefix(rawURL, "https://")
}

// IsValidEmail checks if an email address is valid.
func IsValidEmail(email string) bool {
	// Simple email validation regex
	pattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	re := regexp.MustCompile(pattern)
	return re.MatchString(email)
}
