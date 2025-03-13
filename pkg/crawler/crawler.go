// Package crawler provides website crawling functionality.
package crawler

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/gocolly/colly/v2"
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
	// Validate request
	if req.URL == "" {
		return nil, "", errors.New("URL is required")
	}

	// Generate a job ID
	jobID := uuid.New().String()

	// Set default values
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

	// Create response
	response := &model.CrawlResponse{
		Success: true,
		ID:      jobID,
		URL:     fmt.Sprintf("%s/v1/crawl/%s", s.baseURL, jobID),
	}

	return response, jobID, nil
}

// ProcessCrawlJob processes a crawl job in the background.
func (s *Service) ProcessCrawlJob(jobID string, req model.CrawlRequest) {
	// Parse the base URL
	baseURL, err := url.Parse(req.URL)
	if err != nil {
		return
	}

	// Create a new collector with the specified options
	c := colly.NewCollector(
		colly.MaxDepth(req.MaxDepth),
		colly.Async(true),
		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/109.0.0.0 Safari/537.36"),
	)

	// Set concurrency limit
	err = c.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Parallelism: 5,
	})
	if err != nil {
		return
	}

	// Track visited URLs to avoid duplicates
	visitedURLs := make(map[string]bool)
	var visitedMutex sync.Mutex

	// Track discovered URLs for processing
	discoveredURLs := make([]string, 0)
	var discoveredMutex sync.Mutex

	// Add the initial URL to the discovered URLs
	discoveredURLs = append(discoveredURLs, req.URL)

	// Update the job status to set the initial total count
	if s.updateJobStatusFn != nil {
		_ = s.updateJobStatusFn(jobID, "scraping", 1)
	}

	// Track errors
	errors := make([]model.CrawlError, 0)
	robotsBlocked := make([]string, 0)
	var errorsMutex sync.Mutex

	// Set timeout
	timeout := 30000 // Default 30 seconds
	if req.ScrapeOptions != nil && req.ScrapeOptions.Timeout > 0 {
		timeout = req.ScrapeOptions.Timeout
	}
	c.SetRequestTimeout(time.Duration(timeout) * time.Millisecond)

	// Handle robots.txt
	if !req.IgnoreSitemap {
		c.IgnoreRobotsTxt = false
	} else {
		c.IgnoreRobotsTxt = true
	}

	// Set custom headers if provided
	if req.ScrapeOptions != nil && len(req.ScrapeOptions.Headers) > 0 {
		c.OnRequest(func(r *colly.Request) {
			for key, value := range req.ScrapeOptions.Headers {
				r.Headers.Set(key, value)
			}
		})
	}

	// Handle on HTML callback
	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		// Extract the link
		link := e.Attr("href")
		if link == "" || strings.HasPrefix(link, "#") {
			return
		}

		// Parse the link
		linkURL, err := url.Parse(link)
		if err != nil {
			return
		}

		// Resolve relative URLs
		if linkURL.IsAbs() == false {
			linkURL = baseURL.ResolveReference(linkURL)
		}

		// Skip external links if not allowed
		if !req.AllowExternalLinks && linkURL.Host != baseURL.Host {
			return
		}

		// Skip backward links if not allowed
		if !req.AllowBackwardLinks && isBackwardLink(baseURL.Path, linkURL.Path) {
			return
		}

		// Apply include/exclude path filters
		if !shouldProcessURL(linkURL.String(), req.IncludePaths, req.ExcludePaths) {
			return
		}

		// Normalize the URL
		normalizedURL := linkURL.String()
		if req.IgnoreQueryParameters {
			linkURL.RawQuery = ""
			normalizedURL = linkURL.String()
		}

		// Check if we've already visited this URL
		visitedMutex.Lock()
		if visitedURLs[normalizedURL] {
			visitedMutex.Unlock()
			return
		}
		visitedMutex.Unlock()

		// Add to discovered URLs
		discoveredMutex.Lock()
		if len(discoveredURLs) < req.Limit {
			discoveredURLs = append(discoveredURLs, normalizedURL)
		}
		discoveredMutex.Unlock()

		// Visit the link
		if len(discoveredURLs) < req.Limit {
			c.Visit(normalizedURL)
		}
	})

	// Handle on response
	c.OnResponse(func(r *colly.Response) {
		// Mark URL as visited
		visitedMutex.Lock()
		visitedURLs[r.Request.URL.String()] = true
		visitedMutex.Unlock()

		// Create a scrape request for this URL
		scrapeReq := model.ScrapeRequest{
			URL: r.Request.URL.String(),
		}

		// Copy scrape options from crawl request
		if req.ScrapeOptions != nil {
			scrapeReq.Formats = req.ScrapeOptions.Formats
			scrapeReq.OnlyMainContent = req.ScrapeOptions.OnlyMainContent
			scrapeReq.IncludeTags = req.ScrapeOptions.IncludeTags
			scrapeReq.ExcludeTags = req.ScrapeOptions.ExcludeTags
			scrapeReq.Headers = req.ScrapeOptions.Headers
			scrapeReq.WaitFor = req.ScrapeOptions.WaitFor
			scrapeReq.Timeout = req.ScrapeOptions.Timeout
		}

		// Scrape the URL
		result, err := s.scraper.Scrape(scrapeReq)
		if err != nil {
			// Create an error result
			errorsMutex.Lock()
			errors = append(errors, model.CrawlError{
				ID:        uuid.New().String(),
				Timestamp: time.Now().Format(time.RFC3339),
				URL:       r.Request.URL.String(),
				Error:     err.Error(),
			})
			errorsMutex.Unlock()
			return
		}

		// Call the update job function
		if s.updateJobFn != nil {
			_ = s.updateJobFn(jobID, *result)
		}
	})

	// Handle on error
	c.OnError(func(r *colly.Response, err error) {
		errorsMutex.Lock()
		if strings.Contains(err.Error(), "blocked by robots.txt") {
			robotsBlocked = append(robotsBlocked, r.Request.URL.String())
		} else {
			errors = append(errors, model.CrawlError{
				ID:        uuid.New().String(),
				Timestamp: time.Now().Format(time.RFC3339),
				URL:       r.Request.URL.String(),
				Error:     err.Error(),
			})
		}
		errorsMutex.Unlock()
	})

	// Start crawling
	c.Visit(req.URL)

	// Wait for all requests to finish
	c.Wait()

	// Update job status to completed and set the total count
	if s.updateJobStatusFn != nil {
		// Update the job status to completed and set the total count
		_ = s.updateJobStatusFn(jobID, "completed", len(discoveredURLs))
	}

	// Store errors and robots blocked URLs
	// Note: In a real implementation, we would store these in Redis or another storage
}

// GetCrawlErrors returns the errors for a crawl job.
func (s *Service) GetCrawlErrors(jobID string) (*model.CrawlErrorsResponse, error) {
	// In a real implementation, we would retrieve the errors from storage
	// For now, we'll return an empty response
	return &model.CrawlErrorsResponse{
		Errors:        []model.CrawlError{},
		RobotsBlocked: []string{},
	}, nil
}

// CancelCrawl cancels a crawl job.
func (s *Service) CancelCrawl(jobID string) error {
	// In a real implementation, we would cancel the crawl job
	// For now, we'll just return nil
	return nil
}

// Helper functions

// isBackwardLink checks if a link points to a parent directory.
func isBackwardLink(basePath, linkPath string) bool {
	baseParts := strings.Split(strings.Trim(basePath, "/"), "/")
	linkParts := strings.Split(strings.Trim(linkPath, "/"), "/")

	// If the link has fewer parts than the base, it's a backward link
	if len(linkParts) < len(baseParts) {
		return true
	}

	// Check if the link is in a different branch of the directory tree
	for i := 0; i < len(baseParts); i++ {
		if i >= len(linkParts) || baseParts[i] != linkParts[i] {
			return true
		}
	}

	return false
}

// shouldProcessURL checks if a URL should be processed based on include/exclude paths.
func shouldProcessURL(urlStr string, includePaths, excludePaths []string) bool {
	// If include paths are specified, the URL must match one of them
	if len(includePaths) > 0 {
		matched := false
		for _, includePath := range includePaths {
			if strings.Contains(urlStr, includePath) {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}

	// If exclude paths are specified, the URL must not match any of them
	if len(excludePaths) > 0 {
		for _, excludePath := range excludePaths {
			if strings.Contains(urlStr, excludePath) {
				return false
			}
		}
	}

	return true
}
