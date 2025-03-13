// Package model contains data structures used throughout the application.
package model

// CrawlRequest represents a request to crawl a website and its subpages.
type CrawlRequest struct {
	URL                   string              `json:"url"`
	ExcludePaths          []string            `json:"excludePaths,omitempty"`
	IncludePaths          []string            `json:"includePaths,omitempty"`
	MaxDepth              int                 `json:"maxDepth,omitempty"`
	MaxDiscoveryDepth     int                 `json:"maxDiscoveryDepth,omitempty"`
	IgnoreSitemap         bool                `json:"ignoreSitemap,omitempty"`
	IgnoreQueryParameters bool                `json:"ignoreQueryParameters,omitempty"`
	Limit                 int                 `json:"limit,omitempty"`
	AllowBackwardLinks    bool                `json:"allowBackwardLinks,omitempty"`
	AllowExternalLinks    bool                `json:"allowExternalLinks,omitempty"`
	Webhook               *WebhookConfig      `json:"webhook,omitempty"`
	ScrapeOptions         *CrawlScrapeOptions `json:"scrapeOptions,omitempty"`
}

// CrawlScrapeOptions represents options for scraping during a crawl.
type CrawlScrapeOptions struct {
	Formats             []string          `json:"formats,omitempty"`
	OnlyMainContent     bool              `json:"onlyMainContent,omitempty"`
	IncludeTags         []string          `json:"includeTags,omitempty"`
	ExcludeTags         []string          `json:"excludeTags,omitempty"`
	Headers             map[string]string `json:"headers,omitempty"`
	WaitFor             int               `json:"waitFor,omitempty"`
	Mobile              bool              `json:"mobile,omitempty"`
	SkipTlsVerification bool              `json:"skipTlsVerification,omitempty"`
	Timeout             int               `json:"timeout,omitempty"`
	JSONOptions         *JSONOptions      `json:"jsonOptions,omitempty"`
	Actions             []CrawlAction     `json:"actions,omitempty"`
	Location            *LocationOptions  `json:"location,omitempty"`
	RemoveBase64Images  bool              `json:"removeBase64Images,omitempty"`
	BlockAds            bool              `json:"blockAds,omitempty"`
	Proxy               string            `json:"proxy,omitempty"`
}

// JSONOptions represents options for JSON extraction.
type JSONOptions struct {
	Schema       map[string]interface{} `json:"schema,omitempty"`
	SystemPrompt string                 `json:"systemPrompt,omitempty"`
	Prompt       string                 `json:"prompt,omitempty"`
}

// CrawlAction represents an action to perform during crawling.
type CrawlAction struct {
	Type         string `json:"type"`
	Milliseconds int    `json:"milliseconds,omitempty"`
	Selector     string `json:"selector,omitempty"`
}

// LocationOptions represents location options for crawling.
type LocationOptions struct {
	Country   string   `json:"country,omitempty"`
	Languages []string `json:"languages,omitempty"`
}

// CrawlResponse represents the response to a crawl request.
type CrawlResponse struct {
	Success bool   `json:"success"`
	ID      string `json:"id"`
	URL     string `json:"url"`
}

// CrawlStatus represents the status of a crawl job.
type CrawlStatus struct {
	Status    string         `json:"status"`
	Total     int            `json:"total"`
	Completed int            `json:"completed"`
	ExpiresAt string         `json:"expiresAt"`
	Next      string         `json:"next,omitempty"`
	Data      []ScrapeResult `json:"data,omitempty"`
}

// CrawlError represents an error that occurred during crawling.
type CrawlError struct {
	ID        string `json:"id"`
	Timestamp string `json:"timestamp"`
	URL       string `json:"url"`
	Error     string `json:"error"`
}

// CrawlErrorsResponse represents the response to a crawl errors request.
type CrawlErrorsResponse struct {
	Errors        []CrawlError `json:"errors"`
	RobotsBlocked []string     `json:"robotsBlocked"`
}
