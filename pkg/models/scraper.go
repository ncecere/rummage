package models

// ScrapeRequest represents the request payload for the scrape endpoint
type ScrapeRequest struct {
	URL             string            `json:"url"`
	Formats         []string          `json:"formats,omitempty"`
	OnlyMainContent bool              `json:"onlyMainContent,omitempty"`
	IncludeTags     []string          `json:"includeTags,omitempty"`
	ExcludeTags     []string          `json:"excludeTags,omitempty"`
	Headers         map[string]string `json:"headers,omitempty"`
	WaitFor         int               `json:"waitFor,omitempty"`
	Timeout         int               `json:"timeout,omitempty"`
}

// ScrapeResponse represents the response from the scrape endpoint
type ScrapeResponse struct {
	Success bool       `json:"success"`
	Data    ScrapeData `json:"data,omitempty"`
}

// ScrapeData represents the data returned from the scrape endpoint
type ScrapeData struct {
	Markdown string   `json:"markdown,omitempty"`
	HTML     string   `json:"html,omitempty"`
	RawHTML  string   `json:"rawHtml,omitempty"`
	Links    []string `json:"links,omitempty"`
	Metadata Metadata `json:"metadata,omitempty"`
	Warning  string   `json:"warning,omitempty"`
}

// Metadata represents the metadata of the scraped page
type Metadata struct {
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
	Language    string `json:"language,omitempty"`
	SourceURL   string `json:"sourceURL,omitempty"`
	StatusCode  int    `json:"statusCode,omitempty"`
	Error       string `json:"error,omitempty"`
}
