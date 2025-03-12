package api

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// MapHandler handles requests to the /map endpoint
type MapHandler struct {
	baseURL  string
	redisURL string
	service  MapServicer
}

// NewMapHandler creates a new MapHandler
func NewMapHandler(opts RouterOptions) *MapHandler {
	return &MapHandler{
		baseURL:  opts.BaseURL,
		redisURL: opts.RedisURL,
		service:  NewMapService(),
	}
}

// MapRequest represents the request body for the /map endpoint
type MapRequest struct {
	URL               string `json:"url"`
	Search            string `json:"search,omitempty"`
	IgnoreSitemap     bool   `json:"ignoreSitemap,omitempty"`
	SitemapOnly       bool   `json:"sitemapOnly,omitempty"`
	IncludeSubdomains bool   `json:"includeSubdomains,omitempty"`
	Limit             int    `json:"limit,omitempty"`
	Timeout           int    `json:"timeout,omitempty"`
}

// MapResponse represents the response body for the /map endpoint
type MapResponse struct {
	Success bool     `json:"success"`
	Links   []string `json:"links"`
}

// HandleMap handles POST requests to the /map endpoint
func (h *MapHandler) HandleMap(w http.ResponseWriter, r *http.Request) {
	var req MapRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if req.URL == "" {
		http.Error(w, "URL is required", http.StatusBadRequest)
		return
	}

	// Set default values if not provided
	if req.Limit <= 0 {
		req.Limit = 100 // Default limit is 100
	}
	if req.Limit > 5000 {
		req.Limit = 5000 // Maximum limit is 5000
	}

	// Call the map service to discover URLs
	links, err := h.service.Map(MapOptions{
		URL:               req.URL,
		Search:            req.Search,
		IgnoreSitemap:     req.IgnoreSitemap,
		SitemapOnly:       req.SitemapOnly,
		IncludeSubdomains: req.IncludeSubdomains,
		Limit:             req.Limit,
		Timeout:           req.Timeout,
	})

	if err != nil {
		http.Error(w, fmt.Sprintf("Error mapping URLs: %v", err), http.StatusInternalServerError)
		return
	}

	// Return the response
	resp := MapResponse{
		Success: true,
		Links:   links,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
