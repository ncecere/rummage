// Package scraper provides web scraping functionality.
package scraper

import (
	"errors"
	"net/http"
	"time"

	"github.com/ncecere/rummage/pkg/model"
	"github.com/ncecere/rummage/pkg/utils"
)

// Service provides web scraping functionality.
type Service struct {
	client *http.Client
}

// NewService creates a new scraper service.
func NewService() *Service {
	return &Service{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Scrape scrapes a single URL and returns the result.
func (s *Service) Scrape(req model.ScrapeRequest) (*model.ScrapeResult, error) {
	// Validate request
	if req.URL == "" {
		return nil, errors.New("URL is required")
	}

	// Set default formats if none provided
	if len(req.Formats) == 0 {
		req.Formats = []string{"markdown"}
	}

	// Set default timeout if not provided
	if req.Timeout <= 0 {
		req.Timeout = 30000 // 30 seconds
	}

	// Create a scraper for this request
	scraper := newScraper(s.client, req)

	// Perform the scrape
	return scraper.scrape()
}

// BatchScrape scrapes multiple URLs asynchronously.
func (s *Service) BatchScrape(req model.BatchScrapeRequest) ([]string, []string, error) {
	// Validate request
	if len(req.URLs) == 0 {
		return nil, nil, errors.New("at least one URL is required")
	}

	// Set default formats if none provided
	if len(req.Formats) == 0 {
		req.Formats = []string{"markdown"}
	}

	// Set default timeout if not provided
	if req.Timeout <= 0 {
		req.Timeout = 30000 // 30 seconds
	}

	// Validate URLs and separate valid from invalid
	validURLs := make([]string, 0, len(req.URLs))
	invalidURLs := make([]string, 0)

	for _, url := range req.URLs {
		if utils.IsValidURL(url) {
			validURLs = append(validURLs, url)
		} else {
			invalidURLs = append(invalidURLs, url)
		}
	}

	// If ignoreInvalidURLs is false and there are invalid URLs, return an error
	if !req.IgnoreInvalidURLs && len(invalidURLs) > 0 {
		return validURLs, invalidURLs, errors.New("invalid URLs detected")
	}

	// If no valid URLs, return an error
	if len(validURLs) == 0 {
		return nil, invalidURLs, errors.New("no valid URLs provided")
	}

	return validURLs, invalidURLs, nil
}

// ProcessBatchJob processes a batch job with the given URLs and options.
func (s *Service) ProcessBatchJob(jobID string, urls []string, req model.BatchScrapeRequest,
	resultCallback func(string, model.ScrapeResult) error) {

	// Process each URL
	for _, url := range urls {
		// Create a scrape request for this URL
		scrapeReq := model.ScrapeRequest{
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
		result, err := s.Scrape(scrapeReq)
		if err != nil {
			// Create an error result
			result = &model.ScrapeResult{
				Metadata: &model.ScrapeMetadata{
					SourceURL:  url,
					StatusCode: http.StatusInternalServerError,
				},
			}
		}

		// Call the result callback
		if resultCallback != nil {
			_ = resultCallback(jobID, *result)
		}
	}
}
