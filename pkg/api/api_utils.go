package api

import (
	"encoding/json"
	"net/http"
)

// ErrorResponse represents a standard error response
type ErrorResponse struct {
	Error string `json:"error"`
}

// DecodeJSONBody decodes a JSON request body into the provided struct
func DecodeJSONBody(w http.ResponseWriter, r *http.Request, dst interface{}) bool {
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		WriteErrorResponse(w, "Invalid request body", http.StatusBadRequest)
		return false
	}
	return true
}

// WriteErrorResponse writes a standard error response
func WriteErrorResponse(w http.ResponseWriter, message string, status int) {
	resp := ErrorResponse{
		Error: message,
	}
	WriteJSON(w, status, resp)
}

// WriteServiceError writes an error response from a service error
func WriteServiceError(w http.ResponseWriter, err error) {
	WriteErrorResponse(w, err.Error(), http.StatusInternalServerError)
}

// GetMuxVar gets a variable from the mux.Vars map and validates it
func GetMuxVar(w http.ResponseWriter, vars map[string]string, name string) (string, bool) {
	value := vars[name]
	if value == "" {
		WriteErrorResponse(w, name+" is required", http.StatusBadRequest)
		return "", false
	}
	return value, true
}
