package scraper

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly/v2"
)

// ScrapeRequest represents the request payload for the scrape endpoint
type ScrapeRequest struct {
	URL                 string                   `json:"url"`
	Formats             []string                 `json:"formats,omitempty"`
	OnlyMainContent     bool                     `json:"onlyMainContent,omitempty"`
	IncludeTags         []string                 `json:"includeTags,omitempty"`
	ExcludeTags         []string                 `json:"excludeTags,omitempty"`
	Headers             map[string]string        `json:"headers,omitempty"`
	WaitFor             int                      `json:"waitFor,omitempty"`
	Mobile              bool                     `json:"mobile,omitempty"`
	SkipTlsVerification bool                     `json:"skipTlsVerification,omitempty"`
	Timeout             int                      `json:"timeout,omitempty"`
	JSONOptions         JSONOptions              `json:"jsonOptions,omitempty"`
	Actions             []map[string]interface{} `json:"actions,omitempty"`
	Location            Location                 `json:"location,omitempty"`
	RemoveBase64Images  bool                     `json:"removeBase64Images,omitempty"`
	BlockAds            bool                     `json:"blockAds,omitempty"`
	Proxy               string                   `json:"proxy,omitempty"`
}

// JSONOptions represents the options for JSON extraction
type JSONOptions struct {
	Schema       map[string]interface{} `json:"schema,omitempty"`
	SystemPrompt string                 `json:"systemPrompt,omitempty"`
	Prompt       string                 `json:"prompt,omitempty"`
}

// Location represents the location settings for the request
type Location struct {
	Country   string   `json:"country,omitempty"`
	Languages []string `json:"languages,omitempty"`
}

// ScrapeResponse represents the response from the scrape endpoint
type ScrapeResponse struct {
	Success bool       `json:"success"`
	Data    ScrapeData `json:"data,omitempty"`
}

// ScrapeData represents the data returned from the scrape endpoint
type ScrapeData struct {
	Markdown      string                 `json:"markdown,omitempty"`
	HTML          string                 `json:"html,omitempty"`
	RawHTML       string                 `json:"rawHtml,omitempty"`
	Screenshot    string                 `json:"screenshot,omitempty"`
	Links         []string               `json:"links,omitempty"`
	Actions       *ActionsResult         `json:"actions,omitempty"`
	Metadata      Metadata               `json:"metadata,omitempty"`
	LLMExtraction map[string]interface{} `json:"llm_extraction,omitempty"`
	Warning       string                 `json:"warning,omitempty"`
}

// ActionsResult represents the result of actions performed during scraping
type ActionsResult struct {
	Screenshots []string `json:"screenshots,omitempty"`
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

// HandleScrape handles the scrape endpoint
func HandleScrape(w http.ResponseWriter, r *http.Request) {
	// Parse the request body
	var req ScrapeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request payload: %v", err), http.StatusBadRequest)
		return
	}

	// Validate the URL
	if req.URL == "" {
		http.Error(w, "URL is required", http.StatusBadRequest)
		return
	}

	// Set default formats if not provided
	if len(req.Formats) == 0 {
		req.Formats = []string{"markdown"}
	}

	// Set default timeout if not provided
	if req.Timeout == 0 {
		req.Timeout = 30000
	}

	// Create a new collector
	c := colly.NewCollector(
		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36"),
		colly.MaxDepth(1),
	)

	// Set timeout
	c.SetRequestTimeout(time.Duration(req.Timeout) * time.Millisecond)

	// Initialize the response
	response := ScrapeResponse{
		Success: true,
		Data: ScrapeData{
			Metadata: Metadata{
				SourceURL: req.URL,
			},
		},
	}

	// Handle HTML response
	var htmlContent string
	c.OnResponse(func(r *colly.Response) {
		htmlContent = string(r.Body)
		response.Data.Metadata.StatusCode = r.StatusCode

		// Parse HTML with goquery
		doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
		if err != nil {
			response.Data.Metadata.Error = fmt.Sprintf("Failed to parse HTML: %v", err)
			return
		}

		// Extract metadata
		response.Data.Metadata.Title = doc.Find("title").Text()
		response.Data.Metadata.Description = getMetaContent(doc, "description")
		response.Data.Metadata.Language = doc.Find("html").AttrOr("lang", "")

		// Extract links if requested
		if contains(req.Formats, "links") {
			links := []string{}
			doc.Find("a[href]").Each(func(_ int, s *goquery.Selection) {
				if href, exists := s.Attr("href"); exists {
					links = append(links, href)
				}
			})
			response.Data.Links = links
		}

		// Include HTML if requested
		if contains(req.Formats, "html") {
			// If onlyMainContent is true, try to extract main content
			if req.OnlyMainContent {
				mainContent := extractMainContent(doc)
				if mainContent != "" {
					// Create a simplified HTML document with just the main content
					response.Data.HTML = fmt.Sprintf("<!DOCTYPE html><html>\n\n<body>\n%s\n\n</body></html>", mainContent)
				} else {
					// If no main content found, extract just the body
					bodyHTML, err := doc.Find("body").Html()
					if err == nil && bodyHTML != "" {
						response.Data.HTML = fmt.Sprintf("<!DOCTYPE html><html>\n\n<body>\n%s\n</body></html>", bodyHTML)
					} else {
						response.Data.HTML = htmlContent
					}
				}
			} else {
				// Even when not requesting only main content, provide a cleaner HTML
				bodyHTML, err := doc.Find("body").Html()
				if err == nil && bodyHTML != "" {
					response.Data.HTML = fmt.Sprintf("<!DOCTYPE html><html>\n\n<body>\n%s\n</body></html>", bodyHTML)
				} else {
					response.Data.HTML = htmlContent
				}
			}
		}

		// Include raw HTML if requested
		if contains(req.Formats, "rawHtml") {
			response.Data.RawHTML = htmlContent
		}

		// Handle screenshot format (not implemented yet)
		if contains(req.Formats, "screenshot") {
			response.Data.Warning = "Screenshot format is not implemented in this version"
		}

		// Handle screenshot@fullPage format (not implemented yet)
		if contains(req.Formats, "screenshot@fullPage") {
			response.Data.Warning = "Screenshot@fullPage format is not implemented in this version"
		}

		// Handle JSON format (not implemented yet)
		if contains(req.Formats, "json") {
			if req.JSONOptions.Schema != nil || req.JSONOptions.Prompt != "" || req.JSONOptions.SystemPrompt != "" {
				response.Data.Warning = "JSON extraction with LLM is not implemented in this version"
			}
		}

		// Convert to markdown if requested
		if contains(req.Formats, "markdown") {
			// Configure the converter with options for cleaner output
			converter := md.NewConverter("", true, nil)

			// Remove unwanted elements before conversion
			doc.Find("style, script, noscript, iframe, nav, footer, header").Remove()

			var mdContent string
			var err error

			if req.OnlyMainContent {
				mainContent := extractMainContent(doc)
				if mainContent != "" {
					// For main content, convert directly
					mdContent, err = converter.ConvertString(mainContent)
				} else {
					// If no main content found, try to extract just the body content
					bodyHTML, bodyErr := doc.Find("body").Html()
					if bodyErr == nil && bodyHTML != "" {
						mdContent, err = converter.ConvertString(bodyHTML)
					} else {
						mdContent, err = converter.ConvertString(htmlContent)
					}
				}
			} else {
				// Even when not requesting only main content, try to focus on body
				bodyHTML, bodyErr := doc.Find("body").Html()
				if bodyErr == nil && bodyHTML != "" {
					mdContent, err = converter.ConvertString(bodyHTML)
				} else {
					mdContent, err = converter.ConvertString(htmlContent)
				}
			}

			if err != nil {
				response.Data.Warning = fmt.Sprintf("Failed to convert HTML to markdown: %v", err)
			} else {
				// Clean up the markdown output
				mdContent = cleanMarkdown(mdContent)
				response.Data.Markdown = mdContent
			}
		}
	})

	// Handle errors
	c.OnError(func(r *colly.Response, err error) {
		response.Success = false
		response.Data.Metadata.StatusCode = r.StatusCode
		response.Data.Metadata.Error = err.Error()
	})

	// Start scraping
	err := c.Visit(req.URL)
	if err != nil {
		response.Success = false
		response.Data.Metadata.Error = err.Error()
	}

	// Return the response with proper JSON formatting
	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	encoder.SetEscapeHTML(false) // Don't escape HTML entities in the output
	encoder.Encode(response)
}

// Helper function to check if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// Helper function to get meta content
func getMetaContent(doc *goquery.Document, name string) string {
	content := ""
	doc.Find("meta").Each(func(_ int, s *goquery.Selection) {
		if n, _ := s.Attr("name"); n == name {
			content, _ = s.Attr("content")
		} else if p, _ := s.Attr("property"); p == "og:"+name {
			content, _ = s.Attr("content")
		}
	})
	return content
}

// Helper function to extract main content
func extractMainContent(doc *goquery.Document) string {
	// Try to find main content using common selectors
	selectors := []string{"main", "article", "#content", ".content", "#main", ".main"}

	for _, selector := range selectors {
		if selection := doc.Find(selector).First(); selection.Length() > 0 {
			html, err := selection.Html()
			if err == nil && html != "" {
				return html
			}
		}
	}

	return ""
}

// Helper function to clean up markdown output
func cleanMarkdown(input string) string {
	// Remove extra newlines
	re := regexp.MustCompile(`\n{3,}`)
	cleaned := re.ReplaceAllString(input, "\n\n")

	// Remove title if it's duplicated in the content
	lines := strings.Split(cleaned, "\n")
	if len(lines) > 2 {
		if strings.HasPrefix(lines[2], "# ") && strings.TrimSpace(lines[0]) == strings.TrimSpace(strings.TrimPrefix(lines[2], "# ")) {
			lines = lines[2:]
			cleaned = strings.Join(lines, "\n")
		}
	}

	// Trim leading/trailing whitespace
	cleaned = strings.TrimSpace(cleaned)

	return cleaned
}
