package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/ncecere/rummage/pkg/model"
)

const (
	// Key prefix for crawl jobs
	crawlJobKeyPrefix = "crawl:job:"
	// Key prefix for crawl errors
	crawlErrorsKeyPrefix = "crawl:errors:"
	// Key prefix for robots blocked URLs
	robotsBlockedKeyPrefix = "crawl:robots:"
)

// CreateCrawlJob creates a new crawl job and returns its ID.
func (s *RedisStorage) CreateCrawlJob(jobID string, req model.CrawlRequest) (string, error) {
	key := crawlJobKeyPrefix + jobID

	job := model.CrawlStatus{
		Status:    "pending",
		Total:     0, // Will be updated as URLs are discovered
		Completed: 0,
		ExpiresAt: time.Now().Add(s.jobExpirationTime).Format(time.RFC3339),
	}

	jobData, err := json.Marshal(job)
	if err != nil {
		return "", fmt.Errorf("failed to marshal job data: %w", err)
	}

	if err := s.client.Set(s.ctx, key, jobData, s.jobExpirationTime).Err(); err != nil {
		return "", fmt.Errorf("failed to store job in Redis: %w", err)
	}

	return jobID, nil
}

// GetCrawlJob retrieves a crawl job by ID.
func (s *RedisStorage) GetCrawlJob(jobID string) (*model.CrawlStatus, error) {
	key := crawlJobKeyPrefix + jobID

	jobData, err := s.client.Get(s.ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, fmt.Errorf("job not found: %s", jobID)
		}
		return nil, fmt.Errorf("failed to get job from Redis: %w", err)
	}

	var job model.CrawlStatus
	if err := json.Unmarshal([]byte(jobData), &job); err != nil {
		return nil, fmt.Errorf("failed to unmarshal job data: %w", err)
	}

	return &job, nil
}

// UpdateCrawlJob updates a crawl job with new results.
func (s *RedisStorage) UpdateCrawlJob(jobID string, result model.ScrapeResult) error {
	key := crawlJobKeyPrefix + jobID

	// Get current job data
	job, err := s.GetCrawlJob(jobID)
	if err != nil {
		return err
	}

	// Update job data
	job.Completed++
	job.Data = append(job.Data, result)

	// Update status if completed
	if job.Status == "pending" {
		job.Status = "scraping"
	}

	// Save updated job data
	jobData, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("failed to marshal updated job data: %w", err)
	}

	if err := s.client.Set(s.ctx, key, jobData, s.jobExpirationTime).Err(); err != nil {
		return fmt.Errorf("failed to update job in Redis: %w", err)
	}

	return nil
}

// UpdateCrawlJobStatus updates the status of a crawl job.
func (s *RedisStorage) UpdateCrawlJobStatus(jobID string, status string, total int) error {
	key := crawlJobKeyPrefix + jobID

	// Get current job data
	job, err := s.GetCrawlJob(jobID)
	if err != nil {
		return err
	}

	// Update job data
	job.Status = status
	if total > 0 {
		job.Total = total
	}

	// Save updated job data
	jobData, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("failed to marshal updated job data: %w", err)
	}

	if err := s.client.Set(s.ctx, key, jobData, s.jobExpirationTime).Err(); err != nil {
		return fmt.Errorf("failed to update job in Redis: %w", err)
	}

	return nil
}

// CompleteCrawlJob marks a crawl job as completed.
func (s *RedisStorage) CompleteCrawlJob(jobID string) error {
	return s.UpdateCrawlJobStatus(jobID, "completed", 0)
}

// CancelCrawlJob marks a crawl job as cancelled.
func (s *RedisStorage) CancelCrawlJob(jobID string) error {
	return s.UpdateCrawlJobStatus(jobID, "cancelled", 0)
}

// StoreCrawlError stores an error that occurred during crawling.
func (s *RedisStorage) StoreCrawlError(jobID string, crawlError model.CrawlError) error {
	key := crawlErrorsKeyPrefix + jobID

	// Get current errors
	var crawlErrors []model.CrawlError
	errorsData, err := s.client.Get(s.ctx, key).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		return fmt.Errorf("failed to get errors from Redis: %w", err)
	}

	if errorsData != "" {
		if err := json.Unmarshal([]byte(errorsData), &crawlErrors); err != nil {
			return fmt.Errorf("failed to unmarshal errors data: %w", err)
		}
	}

	// Add new error
	crawlErrors = append(crawlErrors, crawlError)

	// Save updated errors
	errorsDataBytes, err := json.Marshal(crawlErrors)
	if err != nil {
		return fmt.Errorf("failed to marshal errors data: %w", err)
	}

	if err := s.client.Set(s.ctx, key, errorsDataBytes, s.jobExpirationTime).Err(); err != nil {
		return fmt.Errorf("failed to store errors in Redis: %w", err)
	}

	return nil
}

// StoreRobotsBlocked stores a URL that was blocked by robots.txt.
func (s *RedisStorage) StoreRobotsBlocked(jobID string, url string) error {
	key := robotsBlockedKeyPrefix + jobID

	// Get current robots blocked URLs
	var robotsBlocked []string
	robotsData, err := s.client.Get(s.ctx, key).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		return fmt.Errorf("failed to get robots blocked URLs from Redis: %w", err)
	}

	if robotsData != "" {
		if err := json.Unmarshal([]byte(robotsData), &robotsBlocked); err != nil {
			return fmt.Errorf("failed to unmarshal robots blocked data: %w", err)
		}
	}

	// Add new URL
	robotsBlocked = append(robotsBlocked, url)

	// Save updated robots blocked URLs
	robotsDataBytes, err := json.Marshal(robotsBlocked)
	if err != nil {
		return fmt.Errorf("failed to marshal robots blocked data: %w", err)
	}

	if err := s.client.Set(s.ctx, key, robotsDataBytes, s.jobExpirationTime).Err(); err != nil {
		return fmt.Errorf("failed to store robots blocked URLs in Redis: %w", err)
	}

	return nil
}

// GetCrawlErrors retrieves the errors for a crawl job.
func (s *RedisStorage) GetCrawlErrors(jobID string) (*model.CrawlErrorsResponse, error) {
	errorsKey := crawlErrorsKeyPrefix + jobID
	robotsKey := robotsBlockedKeyPrefix + jobID

	// Get errors
	var crawlErrors []model.CrawlError
	errorsData, err := s.client.Get(s.ctx, errorsKey).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		return nil, fmt.Errorf("failed to get errors from Redis: %w", err)
	}

	if errorsData != "" {
		if err := json.Unmarshal([]byte(errorsData), &crawlErrors); err != nil {
			return nil, fmt.Errorf("failed to unmarshal errors data: %w", err)
		}
	}

	// Get robots blocked URLs
	var robotsBlocked []string
	robotsData, err := s.client.Get(s.ctx, robotsKey).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		return nil, fmt.Errorf("failed to get robots blocked URLs from Redis: %w", err)
	}

	if robotsData != "" {
		if err := json.Unmarshal([]byte(robotsData), &robotsBlocked); err != nil {
			return nil, fmt.Errorf("failed to unmarshal robots blocked data: %w", err)
		}
	}

	return &model.CrawlErrorsResponse{
		Errors:        crawlErrors,
		RobotsBlocked: robotsBlocked,
	}, nil
}
