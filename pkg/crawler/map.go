package crawler

import (
	"compress/gzip"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/ncecere/rummage/pkg/model"
)

// XML structures for sitemap parsing
type URLSet struct {
	XMLName xml.Name `xml:"urlset"`
	Xmlns   string   `xml:"xmlns,attr"`
	URLs    []URL    `xml:"url"`
}

type URL struct {
	Loc        string `xml:"loc"`
	LastMod    string `xml:"lastmod,omitempty"`
	ChangeFreq string `xml:"changefreq,omitempty"`
	Priority   string `xml:"priority,omitempty"`
}

type SitemapIndex struct {
	XMLName  xml.Name  `xml:"sitemapindex"`
	Xmlns    string    `xml:"xmlns,attr"`
	Sitemaps []Sitemap `xml:"sitemap"`
}

type Sitemap struct {
	Loc     string `xml:"loc"`
	LastMod string `xml:"lastmod,omitempty"`
}

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
		// Try to find sitemap URLs
		sitemapURLs := []string{
			fmt.Sprintf("%s://%s/sitemap.xml", baseURL.Scheme, baseURL.Host),
			fmt.Sprintf("%s://%s/sitemap_index.xml", baseURL.Scheme, baseURL.Host),
			fmt.Sprintf("%s://%s/sitemap", baseURL.Scheme, baseURL.Host),
		}

		// Also check for sitemaps in the path
		if baseURL.Path != "" && baseURL.Path != "/" {
			// Try with the base path
			basePath := strings.TrimSuffix(baseURL.Path, "/")
			sitemapURLs = append(sitemapURLs,
				fmt.Sprintf("%s://%s%s/sitemap.xml", baseURL.Scheme, baseURL.Host, basePath),
				fmt.Sprintf("%s://%s%s/sitemap", baseURL.Scheme, baseURL.Host, basePath))
		}

		// Try to find sitemap in robots.txt
		robotsTxtURL := fmt.Sprintf("%s://%s/robots.txt", baseURL.Scheme, baseURL.Host)
		robotsTxtResp, err := s.client.Get(robotsTxtURL)
		if err == nil && robotsTxtResp.StatusCode == http.StatusOK {
			defer robotsTxtResp.Body.Close()

			// Read robots.txt content
			robotsTxtContent, err := io.ReadAll(robotsTxtResp.Body)
			if err == nil {
				// Look for Sitemap: entries
				re := regexp.MustCompile(`(?i)Sitemap:\s*(.+)`)
				matches := re.FindAllStringSubmatch(string(robotsTxtContent), -1)
				for _, match := range matches {
					if len(match) > 1 {
						sitemapURLs = append(sitemapURLs, strings.TrimSpace(match[1]))
					}
				}
			}
		}

		// Process all potential sitemap URLs
		for _, sitemapURL := range sitemapURLs {
			// Skip if we've already reached the limit
			if len(discoveredURLs) >= req.Limit {
				break
			}

			sitemapResp, err := s.client.Get(sitemapURL)
			if err != nil || sitemapResp.StatusCode != http.StatusOK {
				continue
			}
			defer sitemapResp.Body.Close()

			// Check if the response is gzipped
			var reader io.Reader = sitemapResp.Body
			if strings.HasSuffix(sitemapURL, ".gz") || sitemapResp.Header.Get("Content-Encoding") == "gzip" {
				gzReader, err := gzip.NewReader(sitemapResp.Body)
				if err != nil {
					continue
				}
				defer gzReader.Close()
				reader = gzReader
			}

			// Try to parse as sitemap index first
			var sitemapIndex SitemapIndex
			indexData, err := io.ReadAll(reader)
			if err != nil {
				continue
			}

			// Try to parse as sitemap index
			if err := xml.Unmarshal(indexData, &sitemapIndex); err == nil && len(sitemapIndex.Sitemaps) > 0 {
				// Process each sitemap in the index
				for _, sitemap := range sitemapIndex.Sitemaps {
					// Skip if we've already reached the limit
					if len(discoveredURLs) >= req.Limit {
						break
					}

					// Process the individual sitemap
					s.processSitemap(sitemap.Loc, req, &discoveredURLs, visitedURLs, &discoveredMutex, &visitedMutex)
				}
			} else {
				// Try to parse as regular sitemap
				var urlset URLSet
				if err := xml.Unmarshal(indexData, &urlset); err == nil && len(urlset.URLs) > 0 {
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
				} else {
					// Try to handle non-standard sitemap formats
					// Some sitemaps might just be a plain list of URLs
					contentStr := string(indexData)

					// Check if it's a plain text list of URLs (one per line)
					lines := strings.Split(contentStr, "\n")
					discoveredMutex.Lock()
					for _, line := range lines {
						line = strings.TrimSpace(line)
						if line == "" || strings.HasPrefix(line, "#") {
							continue // Skip empty lines and comments
						}

						// Check if it looks like a URL
						if strings.HasPrefix(line, "http://") || strings.HasPrefix(line, "https://") {
							if len(discoveredURLs) < req.Limit && shouldProcessURL(line, req.IncludePaths, req.ExcludePaths) {
								// Check if URL matches search term
								if req.Search == "" || strings.Contains(strings.ToLower(line), strings.ToLower(req.Search)) {
									// Check if we've already visited this URL
									visitedMutex.Lock()
									if !visitedURLs[line] {
										visitedURLs[line] = true
										discoveredURLs = append(discoveredURLs, line)
									}
									visitedMutex.Unlock()
								}
							}
						}
					}
					discoveredMutex.Unlock()
				}
			}
		}

		// If sitemapOnly is true, return the discovered URLs
		if req.SitemapOnly {
			return &model.MapResponse{
				Success: true,
				Links:   discoveredURLs,
			}, nil
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

// processSitemap fetches and processes a sitemap URL, adding discovered URLs to the results
func (s *Service) processSitemap(sitemapURL string, req model.MapRequest, discoveredURLs *[]string, visitedURLs map[string]bool, discoveredMutex, visitedMutex *sync.Mutex) {
	// Fetch the sitemap
	sitemapResp, err := s.client.Get(sitemapURL)
	if err != nil || sitemapResp.StatusCode != http.StatusOK {
		return
	}
	defer sitemapResp.Body.Close()

	// Check if the response is gzipped
	var reader io.Reader = sitemapResp.Body
	if strings.HasSuffix(sitemapURL, ".gz") || sitemapResp.Header.Get("Content-Encoding") == "gzip" {
		gzReader, err := gzip.NewReader(sitemapResp.Body)
		if err != nil {
			return
		}
		defer gzReader.Close()
		reader = gzReader
	}

	// Read the sitemap content
	sitemapData, err := io.ReadAll(reader)
	if err != nil {
		return
	}

	// Try to parse as sitemap index first
	var sitemapIndex SitemapIndex
	if err := xml.Unmarshal(sitemapData, &sitemapIndex); err == nil && len(sitemapIndex.Sitemaps) > 0 {
		// Process each sitemap in the index (recursively)
		for _, sitemap := range sitemapIndex.Sitemaps {
			// Skip if we've already reached the limit
			discoveredMutex.Lock()
			if len(*discoveredURLs) >= req.Limit {
				discoveredMutex.Unlock()
				return
			}
			discoveredMutex.Unlock()

			// Process the individual sitemap
			s.processSitemap(sitemap.Loc, req, discoveredURLs, visitedURLs, discoveredMutex, visitedMutex)
		}
	} else {
		// Try to parse as regular sitemap
		var urlset URLSet
		if err := xml.Unmarshal(sitemapData, &urlset); err == nil && len(urlset.URLs) > 0 {
			// Add all URLs from sitemap to discovered URLs
			discoveredMutex.Lock()
			for _, u := range urlset.URLs {
				if len(*discoveredURLs) < req.Limit && shouldProcessURL(u.Loc, req.IncludePaths, req.ExcludePaths) {
					// Check if URL matches search term
					if req.Search == "" || strings.Contains(strings.ToLower(u.Loc), strings.ToLower(req.Search)) {
						// Check if we've already visited this URL
						visitedMutex.Lock()
						if !visitedURLs[u.Loc] {
							visitedURLs[u.Loc] = true
							*discoveredURLs = append(*discoveredURLs, u.Loc)
						}
						visitedMutex.Unlock()
					}
				}
			}
			discoveredMutex.Unlock()
		} else {
			// Try to handle non-standard sitemap formats
			// Some sitemaps might just be a plain list of URLs
			contentStr := string(sitemapData)

			// Check if it's a plain text list of URLs (one per line)
			lines := strings.Split(contentStr, "\n")
			discoveredMutex.Lock()
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if line == "" || strings.HasPrefix(line, "#") {
					continue // Skip empty lines and comments
				}

				// Check if it looks like a URL
				if strings.HasPrefix(line, "http://") || strings.HasPrefix(line, "https://") {
					if len(*discoveredURLs) < req.Limit && shouldProcessURL(line, req.IncludePaths, req.ExcludePaths) {
						// Check if URL matches search term
						if req.Search == "" || strings.Contains(strings.ToLower(line), strings.ToLower(req.Search)) {
							// Check if we've already visited this URL
							visitedMutex.Lock()
							if !visitedURLs[line] {
								visitedURLs[line] = true
								*discoveredURLs = append(*discoveredURLs, line)
							}
							visitedMutex.Unlock()
						}
					}
				}
			}
			discoveredMutex.Unlock()
		}
	}
}
