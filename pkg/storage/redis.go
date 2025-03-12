// Package storage provides data persistence functionality.
package storage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/ncecere/rummage/pkg/config"
	"github.com/ncecere/rummage/pkg/model"
)

const (
	// Key prefix for batch jobs
	batchJobKeyPrefix = "batch:job:"
)

// StorageOptions contains configuration options for the Redis storage.
type StorageOptions struct {
	RedisURL          string
	JobExpirationTime time.Duration
}

// RedisStorage handles Redis operations for the application.
type RedisStorage struct {
	client            *redis.Client
	ctx               context.Context
	jobExpirationTime time.Duration
}

// NewRedisStorage creates a new Redis storage instance.
func NewRedisStorage(redisURL string) (*RedisStorage, error) {
	// Load config for default values
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	return NewRedisStorageWithOptions(StorageOptions{
		RedisURL:          redisURL,
		JobExpirationTime: time.Duration(cfg.JobExpirationHours) * time.Hour,
	})
}

// NewRedisStorageWithOptions creates a new Redis storage instance with custom options.
func NewRedisStorageWithOptions(opts StorageOptions) (*RedisStorage, error) {
	redisOpts, err := redis.ParseURL(opts.RedisURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Redis URL: %w", err)
	}

	client := redis.NewClient(redisOpts)
	ctx := context.Background()

	// Test connection
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisStorage{
		client:            client,
		ctx:               ctx,
		jobExpirationTime: opts.JobExpirationTime,
	}, nil
}

// CreateBatchJob creates a new batch job and returns its ID.
func (s *RedisStorage) CreateBatchJob(urls []string, invalidURLs []string) (string, error) {
	jobID := uuid.New().String()
	key := batchJobKeyPrefix + jobID

	job := model.BatchScrapeStatus{
		Status:    "pending",
		Total:     len(urls),
		Completed: 0,
		ExpiresAt: time.Now().Add(s.jobExpirationTime).Format(time.RFC3339),
	}

	// Store invalid URLs if any
	if len(invalidURLs) > 0 {
		job.Status = "partial"
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

// GetBatchJob retrieves a batch job by ID.
func (s *RedisStorage) GetBatchJob(jobID string) (*model.BatchScrapeStatus, error) {
	key := batchJobKeyPrefix + jobID

	jobData, err := s.client.Get(s.ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, fmt.Errorf("job not found: %s", jobID)
		}
		return nil, fmt.Errorf("failed to get job from Redis: %w", err)
	}

	var job model.BatchScrapeStatus
	if err := json.Unmarshal([]byte(jobData), &job); err != nil {
		return nil, fmt.Errorf("failed to unmarshal job data: %w", err)
	}

	return &job, nil
}

// UpdateBatchJob updates a batch job with new results.
func (s *RedisStorage) UpdateBatchJob(jobID string, result model.ScrapeResult) error {
	key := batchJobKeyPrefix + jobID

	// Get current job data
	job, err := s.GetBatchJob(jobID)
	if err != nil {
		return err
	}

	// Update job data
	job.Completed++
	job.Data = append(job.Data, result)

	// Update status if completed
	if job.Completed >= job.Total {
		job.Status = "completed"
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

// Close closes the Redis connection.
func (s *RedisStorage) Close() error {
	return s.client.Close()
}
