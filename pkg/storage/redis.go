package storage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/ncecere/rummage/pkg/models"
)

const (
	// Default expiration time for jobs (24 hours)
	defaultJobExpiration = 24 * time.Hour

	// Key prefixes for Redis
	jobKeyPrefix = "job:"
	jobListKey   = "jobs"
)

// RedisJobStore implements job storage using Redis
type RedisJobStore struct {
	client *redis.Client
}

// NewRedisJobStore creates a new Redis-based job store
func NewRedisJobStore(addr string) (*RedisJobStore, error) {
	client := redis.NewClient(&redis.Options{
		Addr: addr,
	})

	// Test the connection
	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisJobStore{
		client: client,
	}, nil
}

// CreateJob creates a new batch job and stores it in Redis
func (s *RedisJobStore) CreateJob(ctx context.Context, job models.BatchJob) error {
	// Set job creation time
	job.CreatedAt = time.Now()
	job.UpdatedAt = job.CreatedAt
	job.ExpiresAt = job.CreatedAt.Add(defaultJobExpiration)

	// Serialize the job
	jobData, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("failed to marshal job: %w", err)
	}

	// Store the job in Redis
	jobKey := fmt.Sprintf("%s%s", jobKeyPrefix, job.ID)

	// Use a pipeline to execute multiple commands atomically
	pipe := s.client.Pipeline()

	// Store the job data with expiration
	pipe.Set(ctx, jobKey, jobData, defaultJobExpiration)

	// Add the job ID to the list of jobs
	pipe.LPush(ctx, jobListKey, job.ID)

	// Execute the pipeline
	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to store job in Redis: %w", err)
	}

	return nil
}

// GetJob retrieves a job from Redis by ID
func (s *RedisJobStore) GetJob(ctx context.Context, id string) (*models.BatchJob, error) {
	jobKey := fmt.Sprintf("%s%s", jobKeyPrefix, id)

	// Get the job data from Redis
	jobData, err := s.client.Get(ctx, jobKey).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, errors.New("job not found")
		}
		return nil, fmt.Errorf("failed to get job from Redis: %w", err)
	}

	// Deserialize the job
	var job models.BatchJob
	if err := json.Unmarshal(jobData, &job); err != nil {
		return nil, fmt.Errorf("failed to unmarshal job: %w", err)
	}

	return &job, nil
}

// UpdateJob updates an existing job in Redis
func (s *RedisJobStore) UpdateJob(ctx context.Context, job models.BatchJob) error {
	// Update the job's update time
	job.UpdatedAt = time.Now()

	// Serialize the job
	jobData, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("failed to marshal job: %w", err)
	}

	// Store the job in Redis
	jobKey := fmt.Sprintf("%s%s", jobKeyPrefix, job.ID)

	// Calculate the remaining TTL
	ttl := job.ExpiresAt.Sub(time.Now())
	if ttl <= 0 {
		ttl = defaultJobExpiration
	}

	// Update the job data with the same expiration
	err = s.client.Set(ctx, jobKey, jobData, ttl).Err()
	if err != nil {
		return fmt.Errorf("failed to update job in Redis: %w", err)
	}

	return nil
}

// ListJobs returns a list of all jobs
func (s *RedisJobStore) ListJobs(ctx context.Context) ([]models.BatchJob, error) {
	// Get all job IDs from the list
	jobIDs, err := s.client.LRange(ctx, jobListKey, 0, -1).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to list jobs from Redis: %w", err)
	}

	jobs := make([]models.BatchJob, 0, len(jobIDs))

	// Get each job by ID
	for _, id := range jobIDs {
		job, err := s.GetJob(ctx, id)
		if err != nil {
			// Skip jobs that can't be retrieved
			continue
		}
		jobs = append(jobs, *job)
	}

	return jobs, nil
}

// Close closes the Redis connection
func (s *RedisJobStore) Close() error {
	return s.client.Close()
}
