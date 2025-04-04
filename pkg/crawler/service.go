package crawler

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/ncecere/rummage/pkg/model"
	"github.com/ncecere/rummage/pkg/scraper"
)

// Service provides website crawling functionality.
type Service struct {
	client            *http.Client
	scraper           *scraper.Service
	baseURL           string
	updateJobFn       func(string, model.ScrapeResult) error
	updateJobStatusFn func(string, string, int) error
}

// ServiceOptions contains options for creating a crawler service.
type ServiceOptions struct {
	BaseURL           string
	UpdateJobFn       func(string, model.ScrapeResult) error
	UpdateJobStatusFn func(string, string, int) error
}

// NewService creates a new crawler service.
func NewService(opts ServiceOptions) *Service {
	return &Service{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		scraper:           scraper.NewService(),
		baseURL:           opts.BaseURL,
		updateJobFn:       opts.UpdateJobFn,
		updateJobStatusFn: opts.UpdateJobStatusFn,
	}
}

// Crawl initiates a crawl of the given URL and its subpages.
func (s *Service) Crawl(req model.CrawlRequest) (*model.CrawlResponse, string, error) {
	if req.URL == "" {
		return nil, "", errors.New("URL is required")
	}

	jobID := uuid.New().String()

	if req.MaxDepth <= 0 {
		req.MaxDepth = 10
	}
	if req.Limit <= 0 {
		req.Limit = 1000
	}
	if req.ScrapeOptions == nil {
		req.ScrapeOptions = &model.CrawlScrapeOptions{
			Formats: []string{"markdown"},
		}
	} else if len(req.ScrapeOptions.Formats) == 0 {
		req.ScrapeOptions.Formats = []string{"markdown"}
	}

	response := &model.CrawlResponse{
		Success: true,
		ID:      jobID,
		URL:     fmt.Sprintf("%s/v1/crawl/%s", s.baseURL, jobID),
	}

	return response, jobID, nil
}

// GetCrawlErrors returns the errors for a crawl job.
func (s *Service) GetCrawlErrors(jobID string) (*model.CrawlErrorsResponse, error) {
	return &model.CrawlErrorsResponse{
		Errors:        []model.CrawlError{},
		RobotsBlocked: []string{},
	}, nil
}

// CancelCrawl cancels a crawl job.
func (s *Service) CancelCrawl(jobID string) error {
	return nil
}
