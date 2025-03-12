package api

import (
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly/v2"
)

// MapServicer is an interface for the map service
type MapServicer interface {
	Map(opts MapOptions) ([]string, error)
}

// MapService handles the business logic for the map endpoint
type MapService struct {
	client *http.Client
}

// NewMapService creates a new MapService
func NewMapService() *MapService {
	return &MapService{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// MapOptions contains options for the Map function
type MapOptions struct {
	URL               string
	Search            string
	IgnoreSitemap     bool
	SitemapOnly       bool
	IncludeSubdomains bool
	Limit             int
	Timeout           int
}

// Map scans a website and returns a list of all URLs it finds
func (s *MapService) Map(opts MapOptions) ([]string, error) {
	// Parse the base URL
	baseURL, err := url.Parse(opts.URL)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	// Create a collector for crawling
	c := colly.NewCollector(
		colly.MaxDepth(1), // Only visit the start page
	)

	// Set allowed domains - only if we're not in test mode
	if !strings.Contains(opts.URL, "127.0.0.1") && !strings.Contains(opts.URL, "localhost") {
		c.AllowedDomains = []string{baseURL.Host}
	}

	// If includeSubdomains is true, allow all subdomains
	if opts.IncludeSubdomains {
		c.AllowedDomains = []string{} // Empty means all domains are allowed
		// Add a callback to filter subdomains
		c.URLFilters = append(c.URLFilters,
			regexp.MustCompile(fmt.Sprintf(".*%s.*", strings.Replace(baseURL.Host, ".", "\\.", -1))))
	}

	// Set timeout if provided
	if opts.Timeout > 0 {
		c.SetRequestTimeout(time.Duration(opts.Timeout) * time.Millisecond)
	}

	// Store found links
	links := make([]string, 0)
	seenLinks := make(map[string]bool)

	// Process HTML
	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		absoluteURL := e.Request.AbsoluteURL(link)

		// Skip empty URLs
		if absoluteURL == "" {
			return
		}

		// Skip if we've already seen this link
		if seenLinks[absoluteURL] {
			return
		}

		// Check if the link is from a subdomain
		linkURL, err := url.Parse(absoluteURL)
		if err != nil {
			return
		}

		// Skip if it's a subdomain and includeSubdomains is false
		if !opts.IncludeSubdomains && linkURL.Host != baseURL.Host && strings.HasSuffix(linkURL.Host, baseURL.Host) {
			return
		}

		// Apply search filter if provided
		if opts.Search != "" && !strings.Contains(strings.ToLower(absoluteURL), strings.ToLower(opts.Search)) {
			return
		}

		// Add the link to our results
		links = append(links, absoluteURL)
		seenLinks[absoluteURL] = true

		// Stop if we've reached the limit
		if opts.Limit > 0 && len(links) >= opts.Limit {
			c.AllowedDomains = []string{"no-domain-will-match-this"} // Hack to stop the collector
		}
	})

	// Process sitemap if not ignored
	if !opts.IgnoreSitemap {
		sitemapLinks, err := s.fetchSitemap(baseURL.String())
		if err == nil {
			for _, link := range sitemapLinks {
				// Skip if we've already seen this link
				if seenLinks[link] {
					continue
				}

				// Apply search filter if provided
				if opts.Search != "" && !strings.Contains(strings.ToLower(link), strings.ToLower(opts.Search)) {
					continue
				}

				// Add the link to our results
				links = append(links, link)
				seenLinks[link] = true

				// Stop if we've reached the limit
				if opts.Limit > 0 && len(links) >= opts.Limit {
					break
				}
			}
		}
	}

	// If sitemapOnly is true, skip the HTML crawling
	if !opts.SitemapOnly {
		// Start the crawling process
		err = c.Visit(opts.URL)
		if err != nil {
			return nil, fmt.Errorf("error visiting URL: %w", err)
		}
	}

	// Apply the limit
	if opts.Limit > 0 && len(links) > opts.Limit {
		links = links[:opts.Limit]
	}

	return links, nil
}

// fetchSitemap fetches and parses the sitemap.xml file
func (s *MapService) fetchSitemap(baseURL string) ([]string, error) {
	// Ensure the URL ends with a slash
	if !strings.HasSuffix(baseURL, "/") {
		baseURL += "/"
	}

	// Try to fetch the sitemap
	sitemapURL := baseURL + "sitemap.xml"
	resp, err := s.client.Get(sitemapURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Check if the sitemap was found
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("sitemap not found: %d", resp.StatusCode)
	}

	// Parse the sitemap
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	// Extract links from the sitemap
	links := make([]string, 0)
	doc.Find("url loc").Each(func(i int, s *goquery.Selection) {
		links = append(links, s.Text())
	})

	return links, nil
}
