package api

import (
	"net/http"
)

// handleHealth is a simple health check endpoint.
func (r *Router) handleHealth(w http.ResponseWriter, req *http.Request) {
	respondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
