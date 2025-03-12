package storage

import (
	"errors"
	"testing"
	"time"

	"github.com/ncecere/rummage/pkg/model"
)

// MockRedisStorage is a mock implementation of the RedisStorage for testing
type MockRedisStorage struct {
	jobs map[string]model.BatchScrapeStatus
}

// NewMockRedisStorage creates a new mock Redis storage instance
func NewMockRedisStorage() *MockRedisStorage {
	return &MockRedisStorage{
		jobs: make(map[string]model.BatchScrapeStatus),
	}
}

// CreateBatchJob creates a new batch job and returns its ID
func (m *MockRedisStorage) CreateBatchJob(urls []string, invalidURLs []string) (string, error) {
	jobID := "mock-job-id"

	m.jobs[jobID] = model.BatchScrapeStatus{
		Status:    "pending",
		Total:     len(urls),
		Completed: 0,
		ExpiresAt: time.Now().Add(1 * time.Hour).Format(time.RFC3339),
	}

	return jobID, nil
}

// GetBatchJob retrieves a batch job by ID
func (m *MockRedisStorage) GetBatchJob(jobID string) (*model.BatchScrapeStatus, error) {
	job, ok := m.jobs[jobID]
	if !ok {
		return nil, errors.New("job not found")
	}

	return &job, nil
}

// UpdateBatchJob updates a batch job with new results
func (m *MockRedisStorage) UpdateBatchJob(jobID string, result model.ScrapeResult) error {
	job, ok := m.jobs[jobID]
	if !ok {
		return errors.New("job not found")
	}

	job.Completed++
	job.Data = append(job.Data, result)

	if job.Completed >= job.Total {
		job.Status = "completed"
	}

	m.jobs[jobID] = job

	return nil
}

// Close closes the Redis connection
func (m *MockRedisStorage) Close() error {
	return nil
}

func TestMockRedisStorage_CreateBatchJob(t *testing.T) {
	// Create a mock Redis storage
	storage := NewMockRedisStorage()

	// Test data
	urls := []string{"https://example.com", "https://example.org"}
	invalidURLs := []string{"invalid-url"}

	// Create a batch job
	jobID, err := storage.CreateBatchJob(urls, invalidURLs)
	if err != nil {
		t.Fatalf("Failed to create batch job: %v", err)
	}

	// Verify job ID
	if jobID == "" {
		t.Error("Expected non-empty job ID")
	}

	// Verify job was created
	job, err := storage.GetBatchJob(jobID)
	if err != nil {
		t.Fatalf("Failed to get batch job: %v", err)
	}

	// Verify job data
	if job.Status != "pending" {
		t.Errorf("Expected status 'pending', got '%s'", job.Status)
	}
	if job.Total != len(urls) {
		t.Errorf("Expected total %d, got %d", len(urls), job.Total)
	}
}

func TestMockRedisStorage_GetBatchJob(t *testing.T) {
	// Create a mock Redis storage
	storage := NewMockRedisStorage()

	// Create a batch job
	urls := []string{"https://example.com", "https://example.org"}
	jobID, err := storage.CreateBatchJob(urls, nil)
	if err != nil {
		t.Fatalf("Failed to create batch job: %v", err)
	}

	// Get the batch job
	job, err := storage.GetBatchJob(jobID)
	if err != nil {
		t.Fatalf("Failed to get batch job: %v", err)
	}

	// Verify job data
	if job.Status != "pending" {
		t.Errorf("Expected status 'pending', got '%s'", job.Status)
	}
	if job.Total != len(urls) {
		t.Errorf("Expected total %d, got %d", len(urls), job.Total)
	}
	if job.Completed != 0 {
		t.Errorf("Expected completed 0, got %d", job.Completed)
	}

	// Test non-existent job
	_, err = storage.GetBatchJob("non-existent-job")
	if err == nil {
		t.Error("Expected error for non-existent job, got nil")
	}
}

func TestMockRedisStorage_UpdateBatchJob(t *testing.T) {
	// Create a mock Redis storage
	storage := NewMockRedisStorage()

	// Create a batch job
	urls := []string{"https://example.com", "https://example.org"}
	jobID, err := storage.CreateBatchJob(urls, nil)
	if err != nil {
		t.Fatalf("Failed to create batch job: %v", err)
	}

	// Update the batch job
	result := model.ScrapeResult{
		Markdown: "# Test",
		HTML:     "<h1>Test</h1>",
		Metadata: &model.ScrapeMetadata{
			Title:      "Test",
			SourceURL:  "https://example.com",
			StatusCode: 200,
		},
	}

	err = storage.UpdateBatchJob(jobID, result)
	if err != nil {
		t.Fatalf("Failed to update batch job: %v", err)
	}

	// Get the updated job
	job, err := storage.GetBatchJob(jobID)
	if err != nil {
		t.Fatalf("Failed to get batch job: %v", err)
	}

	// Verify job data
	if job.Completed != 1 {
		t.Errorf("Expected completed 1, got %d", job.Completed)
	}
	if len(job.Data) != 1 {
		t.Errorf("Expected 1 result, got %d", len(job.Data))
	}

	// Update again to complete the job
	err = storage.UpdateBatchJob(jobID, result)
	if err != nil {
		t.Fatalf("Failed to update batch job: %v", err)
	}

	// Get the updated job
	job, err = storage.GetBatchJob(jobID)
	if err != nil {
		t.Fatalf("Failed to get batch job: %v", err)
	}

	// Verify job data
	if job.Completed != 2 {
		t.Errorf("Expected completed 2, got %d", job.Completed)
	}
	if job.Status != "completed" {
		t.Errorf("Expected status 'completed', got '%s'", job.Status)
	}
}

func TestMockRedisStorage_Close(t *testing.T) {
	// Create a mock Redis storage
	storage := NewMockRedisStorage()

	// Close the connection
	err := storage.Close()
	if err != nil {
		t.Fatalf("Failed to close mock connection: %v", err)
	}
}
