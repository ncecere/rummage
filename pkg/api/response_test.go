package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRespondJSON(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		data       interface{}
		wantBody   string
	}{
		{
			name:       "Simple object",
			statusCode: http.StatusOK,
			data:       map[string]string{"message": "success"},
			wantBody:   `{"message":"success"}`,
		},
		{
			name:       "Array",
			statusCode: http.StatusOK,
			data:       []string{"item1", "item2"},
			wantBody:   `["item1","item2"]`,
		},
		{
			name:       "Error status",
			statusCode: http.StatusBadRequest,
			data:       map[string]string{"error": "invalid request"},
			wantBody:   `{"error":"invalid request"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a response recorder
			rr := httptest.NewRecorder()

			// Call the function
			respondJSON(rr, tt.statusCode, tt.data)

			// Check status code
			if status := rr.Code; status != tt.statusCode {
				t.Errorf("handler returned wrong status code: got %v want %v", status, tt.statusCode)
			}

			// Check content type
			contentType := rr.Header().Get("Content-Type")
			if contentType != "application/json" {
				t.Errorf("handler returned wrong content type: got %v want application/json", contentType)
			}

			// Check body
			// Normalize JSON for comparison (remove whitespace)
			var got, want interface{}
			if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
				t.Fatalf("Failed to unmarshal response body: %v", err)
			}
			if err := json.Unmarshal([]byte(tt.wantBody), &want); err != nil {
				t.Fatalf("Failed to unmarshal expected body: %v", err)
			}

			// Compare as JSON
			gotJSON, _ := json.Marshal(got)
			wantJSON, _ := json.Marshal(want)
			if string(gotJSON) != string(wantJSON) {
				t.Errorf("handler returned unexpected body: got %v want %v", string(gotJSON), string(wantJSON))
			}
		})
	}
}

func TestRespondError(t *testing.T) {
	// Create a response recorder
	rr := httptest.NewRecorder()

	// Call the function
	respondError(rr, http.StatusBadRequest, "Invalid request")

	// Check status code
	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
	}

	// Check content type
	contentType := rr.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("handler returned wrong content type: got %v want application/json", contentType)
	}

	// Check body
	var response APIResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response body: %v", err)
	}

	if response.Success {
		t.Errorf("Expected success to be false, got true")
	}

	if response.Error != "Invalid request" {
		t.Errorf("Expected error message 'Invalid request', got '%s'", response.Error)
	}
}

func TestRespondSuccess(t *testing.T) {
	// Create a response recorder
	rr := httptest.NewRecorder()

	// Test data
	data := map[string]string{"message": "success"}

	// Call the function
	respondSuccess(rr, data)

	// Check status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Check content type
	contentType := rr.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("handler returned wrong content type: got %v want application/json", contentType)
	}

	// Check body
	var response APIResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response body: %v", err)
	}

	if !response.Success {
		t.Errorf("Expected success to be true, got false")
	}

	// Check data
	responseData, ok := response.Data.(map[string]interface{})
	if !ok {
		t.Fatalf("Failed to cast response data to map")
	}

	if message, ok := responseData["message"]; !ok || message != "success" {
		t.Errorf("Expected message 'success', got '%v'", message)
	}
}
