package api

import (
	"encoding/json"
	"net/http"
)

// APIResponse represents a standard API response structure.
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// respondJSON sends a JSON response with the given status code and data.
func respondJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	// Create a custom encoder that doesn't escape HTML
	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)

	if err := encoder.Encode(data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// respondError sends a JSON error response with the given status code and error message.
func respondError(w http.ResponseWriter, statusCode int, message string) {
	respondJSON(w, statusCode, APIResponse{
		Success: false,
		Error:   message,
	})
}

// respondSuccess sends a JSON success response with the given data.
func respondSuccess(w http.ResponseWriter, data interface{}) {
	respondJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    data,
	})
}
