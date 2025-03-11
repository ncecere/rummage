package models

import (
	"time"
)

// BatchScrapeRequest represents the request payload for the batch scrape endpoint
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

// WebhookConfig represents the webhook configuration for batch scrape
type WebhookConfig struct {
	URL      string                 `json:"url"`
	Headers  map[string]string      `json:"headers,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
	Events   []string               `json:"events,omitempty"`
}

// BatchScrapeResponse represents the response from the batch scrape endpoint
type BatchScrapeResponse struct {
	Success     bool     `json:"success"`
	ID          string   `json:"id"`
	URL         string   `json:"url,omitempty"`
	InvalidURLs []string `json:"invalidURLs,omitempty"`
}

// BatchScrapeStatusResponse represents the status response for a batch scrape job
type BatchScrapeStatusResponse struct {
	Status      string       `json:"status"`
	Total       int          `json:"total"`
	Completed   int          `json:"completed"`
	CreditsUsed int          `json:"creditsUsed"`
	ExpiresAt   time.Time    `json:"expiresAt"`
	Next        string       `json:"next,omitempty"`
	Data        []ScrapeData `json:"data,omitempty"`
}

// JobStatus represents the status of a batch scrape job
type JobStatus string

const (
	JobStatusPending    JobStatus = "pending"
	JobStatusProcessing JobStatus = "processing"
	JobStatusCompleted  JobStatus = "completed"
	JobStatusFailed     JobStatus = "failed"
)

// BatchJob represents a batch scrape job
type BatchJob struct {
	ID            string
	Status        JobStatus
	Request       BatchScrapeRequest
	Results       []ScrapeData
	Errors        []ScrapeError
	RobotsBlocked []string
	InvalidURLs   []string
	CreatedAt     time.Time
	UpdatedAt     time.Time
	ExpiresAt     time.Time
}
