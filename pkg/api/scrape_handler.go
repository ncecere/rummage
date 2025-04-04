package api

import (
	"encoding/json"
	"net/http"

	"github.com/ncecere/rummage/pkg/model"
)

// ScrapeHandler handles requests to the /scrape endpoint
type ScrapeHandler struct {
	baseURL  string
	redisURL string
}

// NewScrapeHandler creates a new ScrapeHandler
func NewScrapeHandler(opts RouterOptions) *ScrapeHandler {
	return &ScrapeHandler{
		baseURL:  opts.BaseURL,
		redisURL: opts.RedisURL,
	}
}

// handleScrape handles requests to scrape a single URL.
func (r *Router) handleScrape(w http.ResponseWriter, req *http.Request) {
	var scrapeReq model.ScrapeRequest
	if err := json.NewDecoder(req.Body).Decode(&scrapeReq); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body: "+err.Error())
		return
	}

	// Validate URL
	if scrapeReq.URL == "" {
		respondError(w, http.StatusBadRequest, "URL is required")
		return
	}

	// Perform scrape
	result, err := r.scraper.Scrape(scrapeReq)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to scrape URL: "+err.Error())
		return
	}

	// Return result
	respondSuccess(w, result)
}
