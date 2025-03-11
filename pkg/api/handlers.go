package api

import (
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
	if !DecodeJSONBody(w, r, &req) {
		return
	}

	// Validate the URL
	if req.URL == "" {
		WriteErrorResponse(w, "URL is required", http.StatusBadRequest)
		return
	}

	// Set default formats if not provided
	if len(req.Formats) == 0 {
		req.Formats = []string{"markdown"}
	}

	// Call the service
	response, err := h.service.Scrape(req)
	if err != nil {
		WriteServiceError(w, err)
		return
	}

	// Return the response
	if err := WriteJSON(w, http.StatusOK, response); err != nil {
		WriteServiceError(w, err)
	}
}
