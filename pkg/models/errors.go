package models

import (
	"time"
)

// ScrapeError represents an error that occurred during scraping
type ScrapeError struct {
	ID        string    `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	URL       string    `json:"url"`
	Error     string    `json:"error"`
}

// BatchScrapeErrorsResponse represents the response for the batch scrape errors endpoint
type BatchScrapeErrorsResponse struct {
	Errors        []ScrapeError `json:"errors"`
	RobotsBlocked []string      `json:"robotsBlocked"`
}
