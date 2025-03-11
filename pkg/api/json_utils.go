package api

import (
	"bytes"
	"encoding/json"
	"net/http"
)

// WriteJSON is a utility function to write JSON responses with proper formatting
// and without HTML escaping. This ensures consistent JSON output across all handlers.
func WriteJSON(w http.ResponseWriter, status int, data interface{}) error {
	// Set content type and status code
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	// Create a buffer to hold the JSON output
	buffer := &bytes.Buffer{}

	// Create an encoder that doesn't escape HTML and has proper indentation
	encoder := json.NewEncoder(buffer)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")

	// Encode the data to the buffer
	if err := encoder.Encode(data); err != nil {
		return err
	}

	// Get the JSON bytes
	jsonBytes := buffer.Bytes()

	// Check if the first character is a newline (which can happen with some encoders)
	// and ensure the response starts with an opening brace
	if len(jsonBytes) > 0 && jsonBytes[0] == '\n' {
		// Replace the newline with an opening brace
		jsonBytes[0] = '{'
	} else if len(jsonBytes) > 0 && jsonBytes[0] != '{' {
		// If it's not a newline but also not an opening brace, prepend one
		jsonBytes = append([]byte{'{'}, jsonBytes...)
	}

	// Write the JSON bytes to the response
	_, err := w.Write(jsonBytes)
	return err
}
