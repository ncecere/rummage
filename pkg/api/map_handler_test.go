package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// MockMapService is a mock implementation of the MapServicer interface for testing
type MockMapService struct{}

// Map is a mock implementation of the Map method
func (m *MockMapService) Map(opts MapOptions) ([]string, error) {
	// Return some mock links
	return []string{
		"https://example.com/page1",
		"https://example.com/page2",
		"https://example.com/about",
	}, nil
}

// NewMockMapService creates a new MockMapService
func NewMockMapService() MapServicer {
	return &MockMapService{}
}

func TestHandleMap(t *testing.T) {
	// Create a new map handler with a mock service
	handler := &MapHandler{
		baseURL:  "http://localhost:8080",
		redisURL: "localhost:6379",
		service:  NewMockMapService(),
	}

	// Create a test request
	reqBody := MapRequest{
		URL:               "https://example.com",
		Search:            "test",
		IgnoreSitemap:     true,
		SitemapOnly:       false,
		IncludeSubdomains: false,
		Limit:             10,
	}
	reqJSON, err := json.Marshal(reqBody)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

	// Create a test HTTP request
	req, err := http.NewRequest("POST", "/map", bytes.NewBuffer(reqJSON))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Create a response recorder
	rr := httptest.NewRecorder()

	// Call the handler
	handler.HandleMap(rr, req)

	// Check the status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Check the response body
	var resp MapResponse
	err = json.Unmarshal(rr.Body.Bytes(), &resp)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// Verify the response
	if !resp.Success {
		t.Errorf("Expected success to be true, got false")
	}

	// Since we're using a real service in this test, we can't predict the exact links
	// but we can check that we got some links back
	if len(resp.Links) == 0 {
		t.Errorf("Expected some links, got none")
	}
}

func TestHandleMapValidation(t *testing.T) {
	// Create a new map handler with a mock service
	handler := &MapHandler{
		baseURL:  "http://localhost:8080",
		redisURL: "localhost:6379",
		service:  NewMockMapService(),
	}

	// Test cases
	testCases := []struct {
		name           string
		request        MapRequest
		expectedStatus int
	}{
		{
			name:           "Empty URL",
			request:        MapRequest{URL: ""},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Valid URL",
			request:        MapRequest{URL: "https://example.com"},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Limit Exceeds Maximum",
			request:        MapRequest{URL: "https://example.com", Limit: 10000},
			expectedStatus: http.StatusOK, // Should still work, but limit will be capped
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			reqJSON, err := json.Marshal(tc.request)
			if err != nil {
				t.Fatalf("Failed to marshal request: %v", err)
			}

			req, err := http.NewRequest("POST", "/map", bytes.NewBuffer(reqJSON))
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			handler.HandleMap(rr, req)

			if status := rr.Code; status != tc.expectedStatus {
				t.Errorf("Handler returned wrong status code: got %v want %v", status, tc.expectedStatus)
			}
		})
	}
}
