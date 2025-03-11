package service

import (
	"context"
	"fmt"
	"net/url"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/ncecere/rummage/pkg/models"
	"github.com/ncecere/rummage/pkg/storage"
)

// BatchScraperService handles batch scraping operations
type BatchScraperService struct {
	scraper  *ScraperService
	jobStore storage.JobStore
	baseURL  string
}

// NewBatchScraperService creates a new batch scraper service
func NewBatchScraperService(scraper *ScraperService, jobStore storage.JobStore, baseURL string) *BatchScraperService {
	return &BatchScraperService{
		scraper:  scraper,
		jobStore: jobStore,
		baseURL:  baseURL,
	}
}

// BatchScrape processes a batch scrape request
func (s *BatchScraperService) BatchScrape(ctx context.Context, req models.BatchScrapeRequest) (*models.BatchScrapeResponse, error) {
	// Generate a unique ID for the job
	jobID := uuid.New().String()

	// Create a new job
	job := models.BatchJob{
		ID:      jobID,
		Status:  models.JobStatusPending,
		Request: req,
		Results: make([]models.ScrapeData, 0),
	}

	// Validate URLs and collect invalid ones if ignoreInvalidURLs is true
	validURLs := make([]string, 0, len(req.URLs))
	invalidURLs := make([]string, 0)

	for _, rawURL := range req.URLs {
		_, err := url.ParseRequestURI(rawURL)
		if err != nil {
			if req.IgnoreInvalidURLs {
				invalidURLs = append(invalidURLs, rawURL)
				continue
			}
			return nil, fmt.Errorf("invalid URL: %s", rawURL)
		}
		validURLs = append(validURLs, rawURL)
	}

	// Update the job with invalid URLs if any
	if len(invalidURLs) > 0 {
		job.InvalidURLs = invalidURLs
	}

	// Store the job
	if err := s.jobStore.CreateJob(ctx, job); err != nil {
		return nil, fmt.Errorf("failed to create job: %w", err)
	}

	// Start processing the job in a goroutine
	go s.processJob(context.Background(), jobID, validURLs)

	// Return the response
	return &models.BatchScrapeResponse{
		Success:     true,
		ID:          jobID,
		URL:         fmt.Sprintf("%s/v1/batch/scrape/%s", s.baseURL, jobID),
		InvalidURLs: invalidURLs,
	}, nil
}

// GetBatchScrapeStatus retrieves the status of a batch scrape job
func (s *BatchScraperService) GetBatchScrapeStatus(ctx context.Context, jobID string) (*models.BatchScrapeStatusResponse, error) {
	// Get the job from the store
	job, err := s.jobStore.GetJob(ctx, jobID)
	if err != nil {
		return nil, fmt.Errorf("failed to get job: %w", err)
	}

	// Convert job status to response
	status := string(job.Status)

	// Calculate completed count
	completed := len(job.Results)

	// Calculate total count (valid URLs + invalid URLs)
	total := len(job.Request.URLs)

	// For simplicity, we're using 1 credit per URL
	creditsUsed := completed

	return &models.BatchScrapeStatusResponse{
		Status:      status,
		Total:       total,
		Completed:   completed,
		CreditsUsed: creditsUsed,
		ExpiresAt:   job.ExpiresAt,
		Data:        job.Results,
	}, nil
}

// GetBatchScrapeErrors retrieves the errors for a batch scrape job
func (s *BatchScraperService) GetBatchScrapeErrors(ctx context.Context, jobID string) (*models.BatchScrapeErrorsResponse, error) {
	// Get the job from the store
	job, err := s.jobStore.GetJob(ctx, jobID)
	if err != nil {
		return nil, fmt.Errorf("failed to get job: %w", err)
	}

	// Return the errors
	return &models.BatchScrapeErrorsResponse{
		Errors:        job.Errors,
		RobotsBlocked: job.RobotsBlocked,
	}, nil
}

// processURL processes a single URL and returns either a successful ScrapeData or a ScrapeError
// It also handles checking for robots.txt blocked URLs
func (s *BatchScraperService) processURL(ctx context.Context, url string, req models.BatchScrapeRequest) (models.ScrapeData, *models.ScrapeError, bool) {
	// Create a scrape request for this URL
	scrapeReq := models.ScrapeRequest{
		URL:             url,
		Formats:         req.Formats,
		OnlyMainContent: req.OnlyMainContent,
		IncludeTags:     req.IncludeTags,
		ExcludeTags:     req.ExcludeTags,
		Headers:         req.Headers,
		WaitFor:         req.WaitFor,
		Timeout:         req.Timeout,
	}

	// Scrape the URL
	result, err := s.scraper.Scrape(scrapeReq)
	if err != nil {
		// Check if it's a robots.txt blocked error
		if err.Error() == "blocked by robots.txt" {
			return models.ScrapeData{}, nil, true
		}

		// Create a scrape error
		scrapeError := &models.ScrapeError{
			ID:        uuid.New().String(),
			Timestamp: time.Now(),
			URL:       url,
			Error:     err.Error(),
		}
		return models.ScrapeData{}, scrapeError, false
	}

	// Return the successful result
	return result.Data, nil, false
}

// processURLWithSemaphore wraps processURL with semaphore-based concurrency control
func (s *BatchScraperService) processURLWithSemaphore(
	ctx context.Context,
	url string,
	req models.BatchScrapeRequest,
	resultChan chan<- models.ScrapeData,
	errorChan chan<- models.ScrapeError,
	robotsBlockedChan chan<- string,
	semaphore chan struct{},
	wg *sync.WaitGroup,
) {
	defer wg.Done()

	// Acquire semaphore
	semaphore <- struct{}{}
	defer func() { <-semaphore }()

	// Process the URL
	data, scrapeErr, isRobotsBlocked := s.processURL(ctx, url, req)

	// Send the result to the appropriate channel
	if isRobotsBlocked {
		robotsBlockedChan <- url
	} else if scrapeErr != nil {
		errorChan <- *scrapeErr
	} else {
		resultChan <- data
	}
}

// collectResults collects results from channels and updates the job
func (s *BatchScraperService) collectResults(
	resultChan <-chan models.ScrapeData,
	errorChan <-chan models.ScrapeError,
	robotsBlockedChan <-chan string,
	job *models.BatchJob,
) {
	// Collect results
	results := make([]models.ScrapeData, 0)
	for result := range resultChan {
		results = append(results, result)
	}

	// Collect errors
	scrapeErrors := make([]models.ScrapeError, 0)
	for scrapeErr := range errorChan {
		scrapeErrors = append(scrapeErrors, scrapeErr)
	}

	// Collect robots blocked URLs
	robotsBlocked := make([]string, 0)
	for url := range robotsBlockedChan {
		robotsBlocked = append(robotsBlocked, url)
	}

	// Update job with results and errors
	job.Results = results
	job.Errors = scrapeErrors
	job.RobotsBlocked = robotsBlocked
	job.UpdatedAt = time.Now()

	// Update job status
	if len(results) == 0 && (len(scrapeErrors) > 0 || len(robotsBlocked) > 0) {
		// All URLs failed
		job.Status = models.JobStatusFailed
	} else {
		// At least some URLs succeeded
		job.Status = models.JobStatusCompleted
	}
}

// processJob processes a batch scrape job
func (s *BatchScraperService) processJob(ctx context.Context, jobID string, urls []string) {
	// Get the job from the store
	job, err := s.jobStore.GetJob(ctx, jobID)
	if err != nil {
		// Log the error
		fmt.Printf("Failed to get job %s: %v\n", jobID, err)
		return
	}

	// Update job status to processing
	job.Status = models.JobStatusProcessing
	if err := s.jobStore.UpdateJob(ctx, *job); err != nil {
		// Log the error
		fmt.Printf("Failed to update job %s: %v\n", jobID, err)
		return
	}

	// Set up channels for collecting results
	var wg sync.WaitGroup
	resultChan := make(chan models.ScrapeData, len(urls))
	errorChan := make(chan models.ScrapeError, len(urls))
	robotsBlockedChan := make(chan string, len(urls))

	// Limit concurrency to avoid overwhelming the system
	semaphore := make(chan struct{}, 5)

	// Process each URL concurrently with controlled parallelism
	for _, url := range urls {
		wg.Add(1)
		go s.processURLWithSemaphore(
			ctx,
			url,
			job.Request,
			resultChan,
			errorChan,
			robotsBlockedChan,
			semaphore,
			&wg,
		)
	}

	// Wait for all goroutines to finish
	wg.Wait()
	close(resultChan)
	close(errorChan)
	close(robotsBlockedChan)

	// Collect results and update the job
	s.collectResults(resultChan, errorChan, robotsBlockedChan, job)

	// Update the job in the store
	if err := s.jobStore.UpdateJob(ctx, *job); err != nil {
		// Log the error
		fmt.Printf("Failed to update job %s with results: %v\n", jobID, err)
	}

	// TODO: Handle webhook if configured
	if job.Request.Webhook != nil {
		// Implement webhook notification
	}
}
