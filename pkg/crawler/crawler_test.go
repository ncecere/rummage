package crawler

import (
	"testing"

	"github.com/ncecere/rummage/pkg/model"
)

func TestCrawl(t *testing.T) {
	// Create a crawler service
	service := NewService(ServiceOptions{
		BaseURL:     "http://localhost:8080",
		UpdateJobFn: func(string, model.ScrapeResult) error { return nil },
	})

	// Test data
	req := model.CrawlRequest{
		URL:      "https://example.com",
		MaxDepth: 2,
		Limit:    10,
	}

	// Test crawl
	response, jobID, err := service.Crawl(req)
	if err != nil {
		t.Fatalf("Failed to create crawl job: %v", err)
	}

	// Verify response
	if response == nil {
		t.Error("Expected non-nil response")
	}
	if !response.Success {
		t.Error("Expected success to be true")
	}
	if response.ID == "" {
		t.Error("Expected non-empty job ID")
	}
	if response.URL == "" {
		t.Error("Expected non-empty URL")
	}
	if jobID == "" {
		t.Error("Expected non-empty job ID")
	}
}

func TestGetCrawlErrors(t *testing.T) {
	// Create a crawler service
	service := NewService(ServiceOptions{
		BaseURL:     "http://localhost:8080",
		UpdateJobFn: func(string, model.ScrapeResult) error { return nil },
	})

	// Test data
	jobID := "test-job-id"

	// Test get crawl errors
	errors, err := service.GetCrawlErrors(jobID)
	if err != nil {
		t.Fatalf("Failed to get crawl errors: %v", err)
	}

	// Verify errors
	if errors == nil {
		t.Error("Expected non-nil errors")
	}
	if len(errors.Errors) != 0 {
		t.Errorf("Expected 0 errors, got %d", len(errors.Errors))
	}
	if len(errors.RobotsBlocked) != 0 {
		t.Errorf("Expected 0 robots blocked, got %d", len(errors.RobotsBlocked))
	}
}

func TestCancelCrawl(t *testing.T) {
	// Create a crawler service
	service := NewService(ServiceOptions{
		BaseURL:     "http://localhost:8080",
		UpdateJobFn: func(string, model.ScrapeResult) error { return nil },
	})

	// Test data
	jobID := "test-job-id"

	// Test cancel crawl
	err := service.CancelCrawl(jobID)
	if err != nil {
		t.Fatalf("Failed to cancel crawl: %v", err)
	}
}
