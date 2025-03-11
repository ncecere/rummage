package api

import (
	"encoding/json"
	"net/http"

	"github.com/ncecere/rummage/pkg/models"
	"github.com/ncecere/rummage/pkg/service"
)

// ScraperHandler handles HTTP requests for scraping
type ScraperHandler struct {
	service *service.ScraperService
}

// NewScraperHandler creates a new scraper handler
func NewScraperHandler(service *service.ScraperService) *ScraperHandler {
	return &ScraperHandler{
		service: service,
	}
}

// HandleScrape handles the scrape endpoint
func (h *ScraperHandler) HandleScrape(w http.ResponseWriter, r *http.Request) {
	// Parse the request body
	var req models.ScrapeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Validate the URL
	if req.URL == "" {
		http.Error(w, "URL is required", http.StatusBadRequest)
		return
	}

	// Set default formats if not provided
	if len(req.Formats) == 0 {
		req.Formats = []string{"markdown"}
	}

	// Call the service
	response, err := h.service.Scrape(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return the response using the utility function
	if err := WriteJSON(w, http.StatusOK, response); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
