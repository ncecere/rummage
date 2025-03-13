package crawler

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/ncecere/rummage/pkg/model"
)

// Map discovers URLs on a website.
func (s *Service) Map(req model.MapRequest) (*model.MapResponse, error) {
	// Validate request
	if req.URL == "" {
		return nil, fmt.Errorf("URL is required")
	}

	// Set default values
	if req.Limit <= 0 {
		req.Limit = 5000
	}

	// Parse the base URL
	baseURL, err := url.Parse(req.URL)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	// Track discovered URLs and visited URLs
	discoveredURLs := make([]string, 0)
	visitedURLs := make(map[string]bool)
	var discoveredMutex sync.Mutex
	var visitedMutex sync.Mutex

	// Add the initial URL to the discovered URLs
	discoveredURLs = append(discoveredURLs, req.URL)
	visitedURLs[req.URL] = true

	// First, try to fetch the sitemap.xml if not ignored
	if !req.IgnoreSitemap {
		sitemapURL := fmt.Sprintf("%s://%s/sitemap.xml", baseURL.Scheme, baseURL.Host)
		sitemapResp, err := s.client.Get(sitemapURL)
		if err == nil && sitemapResp.StatusCode == http.StatusOK {
			defer sitemapResp.Body.Close()

			// Parse the sitemap XML
			type URLSet struct {
				URLs []struct {
					Loc string `xml:"loc"`
				} `xml:"url"`
			}

			var urlset URLSet
			decoder := xml.NewDecoder(sitemapResp.Body)
			if err := decoder.Decode(&urlset); err == nil {
				// Add all URLs from sitemap to discovered URLs
				discoveredMutex.Lock()
				for _, u := range urlset.URLs {
					if len(discoveredURLs) < req.Limit && shouldProcessURL(u.Loc, req.IncludePaths, req.ExcludePaths) {
						// Check if URL matches search term
						if req.Search == "" || strings.Contains(strings.ToLower(u.Loc), strings.ToLower(req.Search)) {
							// Check if we've already visited this URL
							visitedMutex.Lock()
							if !visitedURLs[u.Loc] {
								visitedURLs[u.Loc] = true
								discoveredURLs = append(discoveredURLs, u.Loc)
							}
							visitedMutex.Unlock()
						}
					}
				}
				discoveredMutex.Unlock()

				// If sitemapOnly is true, return the discovered URLs
				if req.SitemapOnly {
					return &model.MapResponse{
						Success: true,
						Links:   discoveredURLs,
					}, nil
				}
			}
		}
	}

	// Create a new collector with the specified options
	c := colly.NewCollector(
		colly.MaxDepth(1), // Only visit the initial page for mapping
		colly.Async(true),
		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/109.0.0.0 Safari/537.36"),
	)

	// Set concurrency limit
	err = c.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Parallelism: 5,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to set concurrency limit: %w", err)
	}

	// Set timeout
	timeout := 30000 // Default 30 seconds
	if req.Timeout > 0 {
		timeout = req.Timeout
	}
	c.SetRequestTimeout(time.Duration(timeout) * time.Millisecond)

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
		if !req.IncludeSubdomains && linkURL.Host != baseURL.Host {
			return
		}

		// Apply include/exclude path filters
		if !shouldProcessURL(linkURL.String(), req.IncludePaths, req.ExcludePaths) {
			return
		}

		// Check if URL matches search term
		if req.Search != "" && !strings.Contains(strings.ToLower(linkURL.String()), strings.ToLower(req.Search)) {
			return
		}

		// Normalize the URL
		normalizedURL := linkURL.String()

		// Check if we've already visited this URL
		visitedMutex.Lock()
		if visitedURLs[normalizedURL] {
			visitedMutex.Unlock()
			return
		}
		visitedURLs[normalizedURL] = true
		visitedMutex.Unlock()

		// Add to discovered URLs
		discoveredMutex.Lock()
		if len(discoveredURLs) < req.Limit {
			discoveredURLs = append(discoveredURLs, normalizedURL)
		}
		discoveredMutex.Unlock()
	})

	// Start crawling
	c.Visit(req.URL)

	// Wait for all requests to finish
	c.Wait()

	return &model.MapResponse{
		Success: true,
		Links:   discoveredURLs,
	}, nil
}
