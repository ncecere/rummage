package model

import (
	"encoding/json"
	"testing"
)

func TestScrapeRequestJSON(t *testing.T) {
	// Test marshaling and unmarshaling of ScrapeRequest
	req := ScrapeRequest{
		URL:             "https://example.com",
		Formats:         []string{"markdown", "html"},
		OnlyMainContent: true,
		IncludeTags:     []string{"article", "section"},
		ExcludeTags:     []string{"nav", "footer"},
		Headers: map[string]string{
			"User-Agent": "Test Agent",
		},
		WaitFor: 1000,
		Timeout: 30000,
	}

	// Marshal to JSON
	jsonData, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Failed to marshal ScrapeRequest: %v", err)
	}

	// Unmarshal back to struct
	var unmarshaledReq ScrapeRequest
	if err := json.Unmarshal(jsonData, &unmarshaledReq); err != nil {
		t.Fatalf("Failed to unmarshal ScrapeRequest: %v", err)
	}

	// Verify fields
	if unmarshaledReq.URL != req.URL {
		t.Errorf("URL mismatch: got %v, want %v", unmarshaledReq.URL, req.URL)
	}

	if len(unmarshaledReq.Formats) != len(req.Formats) {
		t.Errorf("Formats length mismatch: got %v, want %v", len(unmarshaledReq.Formats), len(req.Formats))
	} else {
		for i, format := range req.Formats {
			if unmarshaledReq.Formats[i] != format {
				t.Errorf("Format mismatch at index %d: got %v, want %v", i, unmarshaledReq.Formats[i], format)
			}
		}
	}

	if unmarshaledReq.OnlyMainContent != req.OnlyMainContent {
		t.Errorf("OnlyMainContent mismatch: got %v, want %v", unmarshaledReq.OnlyMainContent, req.OnlyMainContent)
	}

	if len(unmarshaledReq.Headers) != len(req.Headers) {
		t.Errorf("Headers length mismatch: got %v, want %v", len(unmarshaledReq.Headers), len(req.Headers))
	} else {
		for key, value := range req.Headers {
			if unmarshaledReq.Headers[key] != value {
				t.Errorf("Header mismatch for key %s: got %v, want %v", key, unmarshaledReq.Headers[key], value)
			}
		}
	}

	if unmarshaledReq.WaitFor != req.WaitFor {
		t.Errorf("WaitFor mismatch: got %v, want %v", unmarshaledReq.WaitFor, req.WaitFor)
	}

	if unmarshaledReq.Timeout != req.Timeout {
		t.Errorf("Timeout mismatch: got %v, want %v", unmarshaledReq.Timeout, req.Timeout)
	}
}

func TestBatchScrapeRequestJSON(t *testing.T) {
	// Test marshaling and unmarshaling of BatchScrapeRequest
	req := BatchScrapeRequest{
		URLs:              []string{"https://example.com", "https://example.org"},
		Formats:           []string{"markdown", "html"},
		OnlyMainContent:   true,
		IncludeTags:       []string{"article", "section"},
		ExcludeTags:       []string{"nav", "footer"},
		Headers:           map[string]string{"User-Agent": "Test Agent"},
		WaitFor:           1000,
		Timeout:           30000,
		IgnoreInvalidURLs: true,
		Webhook: &WebhookConfig{
			URL:     "https://webhook.example.com",
			Headers: map[string]string{"Authorization": "Bearer token"},
		},
	}

	// Marshal to JSON
	jsonData, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Failed to marshal BatchScrapeRequest: %v", err)
	}

	// Unmarshal back to struct
	var unmarshaledReq BatchScrapeRequest
	if err := json.Unmarshal(jsonData, &unmarshaledReq); err != nil {
		t.Fatalf("Failed to unmarshal BatchScrapeRequest: %v", err)
	}

	// Verify fields
	if len(unmarshaledReq.URLs) != len(req.URLs) {
		t.Errorf("URLs length mismatch: got %v, want %v", len(unmarshaledReq.URLs), len(req.URLs))
	} else {
		for i, url := range req.URLs {
			if unmarshaledReq.URLs[i] != url {
				t.Errorf("URL mismatch at index %d: got %v, want %v", i, unmarshaledReq.URLs[i], url)
			}
		}
	}

	if unmarshaledReq.IgnoreInvalidURLs != req.IgnoreInvalidURLs {
		t.Errorf("IgnoreInvalidURLs mismatch: got %v, want %v", unmarshaledReq.IgnoreInvalidURLs, req.IgnoreInvalidURLs)
	}

	if unmarshaledReq.Webhook == nil {
		t.Errorf("Webhook is nil, expected non-nil")
	} else {
		if unmarshaledReq.Webhook.URL != req.Webhook.URL {
			t.Errorf("Webhook URL mismatch: got %v, want %v", unmarshaledReq.Webhook.URL, req.Webhook.URL)
		}
	}
}

func TestScrapeResultJSON(t *testing.T) {
	// Test marshaling and unmarshaling of ScrapeResult
	result := ScrapeResult{
		Markdown: "# Example\n\nThis is an example.",
		HTML:     "<h1>Example</h1><p>This is an example.</p>",
		RawHTML:  "<!DOCTYPE html><html><head><title>Example</title></head><body><h1>Example</h1><p>This is an example.</p></body></html>",
		Links:    []string{"https://example.com/page1", "https://example.com/page2"},
		Metadata: &ScrapeMetadata{
			Title:       "Example",
			Description: "This is an example page",
			Language:    "en",
			SourceURL:   "https://example.com",
			StatusCode:  200,
		},
	}

	// Marshal to JSON
	jsonData, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("Failed to marshal ScrapeResult: %v", err)
	}

	// Unmarshal back to struct
	var unmarshaledResult ScrapeResult
	if err := json.Unmarshal(jsonData, &unmarshaledResult); err != nil {
		t.Fatalf("Failed to unmarshal ScrapeResult: %v", err)
	}

	// Verify fields
	if unmarshaledResult.Markdown != result.Markdown {
		t.Errorf("Markdown mismatch: got %v, want %v", unmarshaledResult.Markdown, result.Markdown)
	}

	if unmarshaledResult.HTML != result.HTML {
		t.Errorf("HTML mismatch: got %v, want %v", unmarshaledResult.HTML, result.HTML)
	}

	if unmarshaledResult.RawHTML != result.RawHTML {
		t.Errorf("RawHTML mismatch: got %v, want %v", unmarshaledResult.RawHTML, result.RawHTML)
	}

	if len(unmarshaledResult.Links) != len(result.Links) {
		t.Errorf("Links length mismatch: got %v, want %v", len(unmarshaledResult.Links), len(result.Links))
	}

	if unmarshaledResult.Metadata == nil {
		t.Errorf("Metadata is nil, expected non-nil")
	} else {
		if unmarshaledResult.Metadata.Title != result.Metadata.Title {
			t.Errorf("Metadata Title mismatch: got %v, want %v", unmarshaledResult.Metadata.Title, result.Metadata.Title)
		}
		if unmarshaledResult.Metadata.StatusCode != result.Metadata.StatusCode {
			t.Errorf("Metadata StatusCode mismatch: got %v, want %v", unmarshaledResult.Metadata.StatusCode, result.Metadata.StatusCode)
		}
	}
}
