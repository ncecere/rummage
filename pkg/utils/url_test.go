package utils

import (
	"testing"
)

func TestIsValidURL(t *testing.T) {
	tests := []struct {
		name string
		url  string
		want bool
	}{
		{
			name: "Valid HTTP URL",
			url:  "http://example.com",
			want: true,
		},
		{
			name: "Valid HTTPS URL",
			url:  "https://example.com/path?query=value",
			want: true,
		},
		{
			name: "Valid URL with port",
			url:  "https://example.com:8080/path",
			want: true,
		},
		{
			name: "Invalid URL - missing scheme",
			url:  "example.com",
			want: false,
		},
		{
			name: "Invalid URL - empty string",
			url:  "",
			want: false,
		},
		{
			name: "Invalid URL - malformed",
			url:  "http://",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsValidURL(tt.url); got != tt.want {
				t.Errorf("IsValidURL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNormalizeURL(t *testing.T) {
	tests := []struct {
		name string
		url  string
		want string
	}{
		{
			name: "URL with trailing slash",
			url:  "https://example.com/path/",
			want: "https://example.com/path",
		},
		{
			name: "URL with fragment",
			url:  "https://example.com/path#fragment",
			want: "https://example.com/path",
		},
		{
			name: "URL without scheme",
			url:  "example.com",
			want: "http://example.com",
		},
		{
			name: "URL with query parameters",
			url:  "https://example.com/path?query=value",
			want: "https://example.com/path?query=value",
		},
		{
			name: "Invalid URL",
			url:  "invalid-url",
			want: "invalid-url",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NormalizeURL(tt.url); got != tt.want {
				t.Errorf("NormalizeURL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExtractDomain(t *testing.T) {
	tests := []struct {
		name string
		url  string
		want string
	}{
		{
			name: "Simple domain",
			url:  "https://example.com",
			want: "example.com",
		},
		{
			name: "Subdomain",
			url:  "https://sub.example.com/path",
			want: "sub.example.com",
		},
		{
			name: "Domain with port",
			url:  "https://example.com:8080",
			want: "example.com",
		},
		{
			name: "Invalid URL",
			url:  "invalid-url",
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ExtractDomain(tt.url); got != tt.want {
				t.Errorf("ExtractDomain() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsRelativeURL(t *testing.T) {
	tests := []struct {
		name string
		url  string
		want bool
	}{
		{
			name: "Absolute URL with HTTP",
			url:  "http://example.com",
			want: false,
		},
		{
			name: "Absolute URL with HTTPS",
			url:  "https://example.com",
			want: false,
		},
		{
			name: "Relative URL - path only",
			url:  "/path/to/resource",
			want: true,
		},
		{
			name: "Relative URL - query parameters",
			url:  "path?query=value",
			want: true,
		},
		{
			name: "Relative URL - fragment",
			url:  "#section",
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsRelativeURL(tt.url); got != tt.want {
				t.Errorf("IsRelativeURL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsValidEmail(t *testing.T) {
	tests := []struct {
		name  string
		email string
		want  bool
	}{
		{
			name:  "Valid email",
			email: "user@example.com",
			want:  true,
		},
		{
			name:  "Valid email with subdomain",
			email: "user@sub.example.com",
			want:  true,
		},
		{
			name:  "Valid email with plus",
			email: "user+tag@example.com",
			want:  true,
		},
		{
			name:  "Valid email with dots",
			email: "first.last@example.com",
			want:  true,
		},
		{
			name:  "Invalid email - missing @",
			email: "userexample.com",
			want:  false,
		},
		{
			name:  "Invalid email - missing domain",
			email: "user@",
			want:  false,
		},
		{
			name:  "Invalid email - missing username",
			email: "@example.com",
			want:  false,
		},
		{
			name:  "Invalid email - invalid TLD",
			email: "user@example.c",
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsValidEmail(tt.email); got != tt.want {
				t.Errorf("IsValidEmail() = %v, want %v", got, tt.want)
			}
		})
	}
}
