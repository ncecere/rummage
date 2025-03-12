// Package model contains data structures used throughout the application.
package model

// ScrapeRequest represents a request to scrape a single URL.
type ScrapeRequest struct {
	URL             string            `json:"url"`
	Formats         []string          `json:"formats,omitempty"`
	OnlyMainContent bool              `json:"onlyMainContent,omitempty"`
	IncludeTags     []string          `json:"includeTags,omitempty"`
	ExcludeTags     []string          `json:"excludeTags,omitempty"`
	Headers         map[string]string `json:"headers,omitempty"`
	WaitFor         int               `json:"waitFor,omitempty"`
	Timeout         int               `json:"timeout,omitempty"`
}

// BatchScrapeRequest represents a request to scrape multiple URLs.
type BatchScrapeRequest struct {
	URLs              []string          `json:"urls"`
	Formats           []string          `json:"formats,omitempty"`
	OnlyMainContent   bool              `json:"onlyMainContent,omitempty"`
	IncludeTags       []string          `json:"includeTags,omitempty"`
	ExcludeTags       []string          `json:"excludeTags,omitempty"`
	Headers           map[string]string `json:"headers,omitempty"`
	WaitFor           int               `json:"waitFor,omitempty"`
	Timeout           int               `json:"timeout,omitempty"`
	IgnoreInvalidURLs bool              `json:"ignoreInvalidURLs,omitempty"`
	Webhook           *WebhookConfig    `json:"webhook,omitempty"`
}

// WebhookConfig represents webhook configuration for batch scraping.
type WebhookConfig struct {
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers,omitempty"`
}

// ScrapeResult represents the result of a scrape operation.
type ScrapeResult struct {
	Markdown string          `json:"markdown,omitempty"`
	HTML     string          `json:"html,omitempty"`
	RawHTML  string          `json:"rawHtml,omitempty"`
	Links    []string        `json:"links,omitempty"`
	Metadata *ScrapeMetadata `json:"metadata,omitempty"`
}

// ScrapeMetadata contains metadata about the scraped page.
type ScrapeMetadata struct {
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
	Language    string `json:"language,omitempty"`
	SourceURL   string `json:"sourceURL,omitempty"`
	StatusCode  int    `json:"statusCode,omitempty"`
}

// BatchScrapeResponse represents the response to a batch scrape request.
type BatchScrapeResponse struct {
	ID          string   `json:"id"`
	URL         string   `json:"url"`
	InvalidURLs []string `json:"invalidURLs,omitempty"`
}

// BatchScrapeStatus represents the status of a batch scrape job.
type BatchScrapeStatus struct {
	Status    string         `json:"status"`
	Total     int            `json:"total"`
	Completed int            `json:"completed"`
	ExpiresAt string         `json:"expiresAt"`
	Data      []ScrapeResult `json:"data,omitempty"`
}
