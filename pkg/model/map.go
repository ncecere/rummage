// Package model contains data structures used throughout the application.
package model

// MapRequest represents a request to map a website's URLs.
type MapRequest struct {
	URL               string   `json:"url"`
	Search            string   `json:"search,omitempty"`
	IgnoreSitemap     bool     `json:"ignoreSitemap,omitempty"`
	SitemapOnly       bool     `json:"sitemapOnly,omitempty"`
	IncludeSubdomains bool     `json:"includeSubdomains,omitempty"`
	Limit             int      `json:"limit,omitempty"`
	Timeout           int      `json:"timeout,omitempty"`
	ExcludePaths      []string `json:"excludePaths,omitempty"`
	IncludePaths      []string `json:"includePaths,omitempty"`
}

// MapResponse represents the response to a map request.
type MapResponse struct {
	Success bool     `json:"success"`
	Links   []string `json:"links"`
}
