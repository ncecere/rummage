package api

import (
	"encoding/json"
	"net/http"

	"github.com/ncecere/rummage/pkg/model"
)

// handleMap handles requests to map a website's URLs.
func (r *Router) handleMap(w http.ResponseWriter, req *http.Request) {
	var mapReq model.MapRequest
	if err := json.NewDecoder(req.Body).Decode(&mapReq); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body: "+err.Error())
		return
	}

	// Validate URL
	if mapReq.URL == "" {
		respondError(w, http.StatusBadRequest, "URL is required")
		return
	}

	// Perform map operation
	result, err := r.crawler.Map(mapReq)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to map website: "+err.Error())
		return
	}

	// Return result
	respondSuccess(w, result)
}
