package storage

import (
	"context"

	"github.com/ncecere/rummage/pkg/models"
)

// JobStore defines the interface for job storage
type JobStore interface {
	// CreateJob creates a new batch job
	CreateJob(ctx context.Context, job models.BatchJob) error

	// GetJob retrieves a job by ID
	GetJob(ctx context.Context, id string) (*models.BatchJob, error)

	// UpdateJob updates an existing job
	UpdateJob(ctx context.Context, job models.BatchJob) error

	// ListJobs returns a list of all jobs
	ListJobs(ctx context.Context) ([]models.BatchJob, error)

	// Close closes the job store
	Close() error
}
