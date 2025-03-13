package storage

import (
	"errors"
	"testing"
	"time"

	"github.com/ncecere/rummage/pkg/model"
)

func TestCrawlJobOperations(t *testing.T) {
	// Create a mock Redis storage
	storage := NewMockRedisStorage()

	// Test data
	url := "https://example.com"
	req := model.CrawlRequest{
		URL:      url,
		MaxDepth: 2,
		Limit:    10,
	}

	// Create a crawl job
	jobID, err := storage.CreateCrawlJob(url, req)
	if err != nil {
		t.Fatalf("Failed to create crawl job: %v", err)
	}

	// Verify job ID
	if jobID == "" {
		t.Error("Expected non-empty job ID")
	}

	// Get the crawl job
	job, err := storage.GetCrawlJob(jobID)
	if err != nil {
		t.Fatalf("Failed to get crawl job: %v", err)
	}

	// Verify job data
	if job.Status != "pending" {
		t.Errorf("Expected status 'pending', got '%s'", job.Status)
	}
	if job.Total != 0 {
		t.Errorf("Expected total 0, got %d", job.Total)
	}
	if job.Completed != 0 {
		t.Errorf("Expected completed 0, got %d", job.Completed)
	}

	// Update the crawl job
	result := model.ScrapeResult{
		Markdown: "# Test",
		HTML:     "<h1>Test</h1>",
		Metadata: &model.ScrapeMetadata{
			Title:      "Test",
			SourceURL:  "https://example.com",
			StatusCode: 200,
		},
	}

	err = storage.UpdateCrawlJob(jobID, result)
	if err != nil {
		t.Fatalf("Failed to update crawl job: %v", err)
	}

	// Get the updated job
	job, err = storage.GetCrawlJob(jobID)
	if err != nil {
		t.Fatalf("Failed to get crawl job: %v", err)
	}

	// Verify job data
	if job.Status != "scraping" {
		t.Errorf("Expected status 'scraping', got '%s'", job.Status)
	}
	if job.Completed != 1 {
		t.Errorf("Expected completed 1, got %d", job.Completed)
	}
	if len(job.Data) != 1 {
		t.Errorf("Expected 1 result, got %d", len(job.Data))
	}

	// Complete the crawl job
	err = storage.CompleteCrawlJob(jobID)
	if err != nil {
		t.Fatalf("Failed to complete crawl job: %v", err)
	}

	// Get the updated job
	job, err = storage.GetCrawlJob(jobID)
	if err != nil {
		t.Fatalf("Failed to get crawl job: %v", err)
	}

	// Verify job data
	if job.Status != "completed" {
		t.Errorf("Expected status 'completed', got '%s'", job.Status)
	}
}

func TestCrawlErrorOperations(t *testing.T) {
	// Create a mock Redis storage
	storage := NewMockRedisStorage()

	// Test data
	jobID := "test-job-id"
	crawlError := model.CrawlError{
		ID:        "error-id",
		Timestamp: time.Now().Format(time.RFC3339),
		URL:       "https://example.com",
		Error:     "Test error",
	}

	// Store a crawl error
	err := storage.StoreCrawlError(jobID, crawlError)
	if err != nil {
		t.Fatalf("Failed to store crawl error: %v", err)
	}

	// Get crawl errors
	errors, err := storage.GetCrawlErrors(jobID)
	if err != nil {
		t.Fatalf("Failed to get crawl errors: %v", err)
	}

	// Verify errors
	if len(errors.Errors) != 1 {
		t.Errorf("Expected 1 error, got %d", len(errors.Errors))
	}
	if errors.Errors[0].ID != crawlError.ID {
		t.Errorf("Expected error ID '%s', got '%s'", crawlError.ID, errors.Errors[0].ID)
	}
	if errors.Errors[0].URL != crawlError.URL {
		t.Errorf("Expected error URL '%s', got '%s'", crawlError.URL, errors.Errors[0].URL)
	}
	if errors.Errors[0].Error != crawlError.Error {
		t.Errorf("Expected error message '%s', got '%s'", crawlError.Error, errors.Errors[0].Error)
	}
}

func TestRobotsBlockedOperations(t *testing.T) {
	// Create a mock Redis storage
	storage := NewMockRedisStorage()

	// Test data
	jobID := "test-job-id"
	url := "https://example.com/robots.txt"

	// Store a robots blocked URL
	err := storage.StoreRobotsBlocked(jobID, url)
	if err != nil {
		t.Fatalf("Failed to store robots blocked URL: %v", err)
	}

	// Get crawl errors
	errors, err := storage.GetCrawlErrors(jobID)
	if err != nil {
		t.Fatalf("Failed to get crawl errors: %v", err)
	}

	// Verify robots blocked
	if len(errors.RobotsBlocked) != 1 {
		t.Errorf("Expected 1 robots blocked URL, got %d", len(errors.RobotsBlocked))
	}
	if errors.RobotsBlocked[0] != url {
		t.Errorf("Expected robots blocked URL '%s', got '%s'", url, errors.RobotsBlocked[0])
	}
}

// Add these methods to the MockRedisStorage type in redis_test.go

// CreateCrawlJob creates a new crawl job and returns its ID.
func (m *MockRedisStorage) CreateCrawlJob(url string, req model.CrawlRequest) (string, error) {
	jobID := url

	m.jobs[jobID] = model.BatchScrapeStatus{
		Status:    "pending",
		Total:     0,
		Completed: 0,
		ExpiresAt: time.Now().Add(1 * time.Hour).Format(time.RFC3339),
	}

	return jobID, nil
}

// GetCrawlJob retrieves a crawl job by ID.
func (m *MockRedisStorage) GetCrawlJob(jobID string) (*model.CrawlStatus, error) {
	job, ok := m.jobs[jobID]
	if !ok {
		return nil, errors.New("job not found")
	}

	// Convert BatchScrapeStatus to CrawlStatus
	crawlJob := &model.CrawlStatus{
		Status:    job.Status,
		Total:     job.Total,
		Completed: job.Completed,
		ExpiresAt: job.ExpiresAt,
		Data:      job.Data,
	}

	return crawlJob, nil
}

// UpdateCrawlJob updates a crawl job with new results.
func (m *MockRedisStorage) UpdateCrawlJob(jobID string, result model.ScrapeResult) error {
	job, ok := m.jobs[jobID]
	if !ok {
		return errors.New("job not found")
	}

	job.Completed++
	job.Data = append(job.Data, result)

	if job.Status == "pending" {
		job.Status = "scraping"
	}

	m.jobs[jobID] = job

	return nil
}

// UpdateCrawlJobStatus updates the status of a crawl job.
func (m *MockRedisStorage) UpdateCrawlJobStatus(jobID string, status string, total int) error {
	job, ok := m.jobs[jobID]
	if !ok {
		return errors.New("job not found")
	}

	job.Status = status
	if total > 0 {
		job.Total = total
	}

	m.jobs[jobID] = job

	return nil
}

// CompleteCrawlJob marks a crawl job as completed.
func (m *MockRedisStorage) CompleteCrawlJob(jobID string) error {
	return m.UpdateCrawlJobStatus(jobID, "completed", 0)
}

// CancelCrawlJob marks a crawl job as cancelled.
func (m *MockRedisStorage) CancelCrawlJob(jobID string) error {
	return m.UpdateCrawlJobStatus(jobID, "cancelled", 0)
}

// StoreCrawlError stores an error that occurred during crawling.
func (m *MockRedisStorage) StoreCrawlError(jobID string, crawlError model.CrawlError) error {
	if _, ok := m.crawlErrors[jobID]; !ok {
		m.crawlErrors[jobID] = []model.CrawlError{}
	}
	m.crawlErrors[jobID] = append(m.crawlErrors[jobID], crawlError)
	return nil
}

// StoreRobotsBlocked stores a URL that was blocked by robots.txt.
func (m *MockRedisStorage) StoreRobotsBlocked(jobID string, url string) error {
	if _, ok := m.robotsBlocked[jobID]; !ok {
		m.robotsBlocked[jobID] = []string{}
	}
	m.robotsBlocked[jobID] = append(m.robotsBlocked[jobID], url)
	return nil
}

// GetCrawlErrors retrieves the errors for a crawl job.
func (m *MockRedisStorage) GetCrawlErrors(jobID string) (*model.CrawlErrorsResponse, error) {
	errors := m.crawlErrors[jobID]
	if errors == nil {
		errors = []model.CrawlError{}
	}

	blocked := m.robotsBlocked[jobID]
	if blocked == nil {
		blocked = []string{}
	}

	return &model.CrawlErrorsResponse{
		Errors:        errors,
		RobotsBlocked: blocked,
	}, nil
}
