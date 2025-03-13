package scraper

import (
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly/v2"
	"github.com/ncecere/rummage/pkg/model"
	"github.com/ncecere/rummage/pkg/utils"
)

// scraper handles the scraping of a single URL.
type scraper struct {
	client  *http.Client
	request model.ScrapeRequest
}

// newScraper creates a new scraper for the given request.
func newScraper(client *http.Client, req model.ScrapeRequest) *scraper {
	return &scraper{
		client:  client,
		request: req,
	}
}

// scrape performs the scraping operation and returns the result.
func (s *scraper) scrape() (*model.ScrapeResult, error) {
	// Create a new collector
	c := colly.NewCollector(
		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/109.0.0.0 Safari/537.36"),
	)

	// Set timeout
	c.SetRequestTimeout(time.Duration(s.request.Timeout) * time.Millisecond)

	// Set custom headers if provided
	if len(s.request.Headers) > 0 {
		c.OnRequest(func(r *colly.Request) {
			for key, value := range s.request.Headers {
				r.Headers.Set(key, value)
			}
		})
	}

	// Initialize result
	result := &model.ScrapeResult{
		Metadata: &model.ScrapeMetadata{
			SourceURL: s.request.URL,
		},
	}

	// Wait for JavaScript to load if specified
	if s.request.WaitFor > 0 {
		c.OnRequest(func(r *colly.Request) {
			time.Sleep(time.Duration(s.request.WaitFor) * time.Millisecond)
		})
	}

	// Process HTML
	c.OnResponse(func(r *colly.Response) {
		result.Metadata.StatusCode = r.StatusCode

		// Parse HTML
		doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(r.Body)))
		if err != nil {
			return
		}

		// Extract metadata
		result.Metadata.Title = doc.Find("title").Text()
		result.Metadata.Description = doc.Find("meta[name=description]").AttrOr("content", "")
		result.Metadata.Language = doc.Find("html").AttrOr("lang", "")

		// Process content based on requested formats
		for _, format := range s.request.Formats {
			switch format {
			case "markdown":
				result.Markdown = s.extractMarkdown(doc)
			case "html":
				result.HTML = s.extractHTML(doc)
			case "rawHtml":
				result.RawHTML = string(r.Body)
			case "links":
				result.Links = s.extractLinks(doc)
			}
		}
	})

	// Start scraping
	err := c.Visit(s.request.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to scrape URL: %w", err)
	}

	return result, nil
}

// extractMarkdown extracts markdown content from the document.
func (s *scraper) extractMarkdown(doc *goquery.Document) string {
	// Create a copy of the document to modify
	docCopy := cloneDocument(doc)

	// Apply content filters
	s.applyContentFilters(docCopy)

	// Convert HTML to markdown
	converter := md.NewConverter("", true, nil)

	// Add an After hook to remove line numbers from code blocks
	converter.After(func(markdown string) string {
		// Use a regular expression to find and remove line numbers in code blocks
		return removeLineNumbersFromCodeBlocks(markdown)
	})

	html, _ := docCopy.Html()
	markdown, _ := converter.ConvertString(html)

	return markdown
}

// removeLineNumbersFromCodeBlocks removes line numbers from code blocks in markdown.
// It looks for patterns like \n123\n124\n125 and removes them.
func removeLineNumbersFromCodeBlocks(markdown string) string {
	// Split the markdown by code block delimiters
	parts := strings.Split(markdown, "```")

	// Process each code block (odd indices are code blocks)
	for i := 1; i < len(parts); i += 2 {
		if i < len(parts) {
			// Find and remove line numbers in the code block
			lines := strings.Split(parts[i], "\n")
			if len(lines) > 1 {
				// Check if the first line is the language identifier
				startIdx := 1
				if !isLineNumber(lines[0]) {
					startIdx = 2 // Skip language identifier and first line
				}

				// Process each line
				cleanedLines := []string{lines[0]} // Keep language identifier
				for j := startIdx; j < len(lines); j++ {
					// Remove line numbers at the beginning of lines
					cleanedLines = append(cleanedLines, removeLineNumber(lines[j]))
				}

				// Rejoin the cleaned lines
				parts[i] = strings.Join(cleanedLines, "\n")
			}
		}
	}

	// Rejoin the markdown with code block delimiters
	return strings.Join(parts, "```")
}

// isLineNumber checks if a string is just a line number.
func isLineNumber(s string) bool {
	// Trim whitespace
	s = strings.TrimSpace(s)

	// Check if it's a number
	_, err := strconv.Atoi(s)
	return err == nil
}

// removeLineNumber removes line numbers from the beginning of a line.
func removeLineNumber(line string) string {
	// Match patterns like "123 " at the beginning of a line
	re := regexp.MustCompile(`^\s*\d+\s+`)
	return re.ReplaceAllString(line, "")
}

// extractHTML extracts processed HTML content from the document.
func (s *scraper) extractHTML(doc *goquery.Document) string {
	// Create a copy of the document to modify
	docCopy := cloneDocument(doc)

	// Apply content filters
	s.applyContentFilters(docCopy)

	// Get HTML
	html, err := docCopy.Html()
	if err != nil {
		return ""
	}

	return html
}

// extractLinks extracts all links from the document.
func (s *scraper) extractLinks(doc *goquery.Document) []string {
	links := make([]string, 0)
	baseURL, _ := url.Parse(s.request.URL)

	doc.Find("a[href]").Each(func(_ int, sel *goquery.Selection) {
		href, exists := sel.Attr("href")
		if !exists || href == "" || strings.HasPrefix(href, "#") {
			return
		}

		// Resolve relative URLs
		if utils.IsRelativeURL(href) {
			u, err := url.Parse(href)
			if err != nil {
				return
			}
			href = baseURL.ResolveReference(u).String()
		}

		// Only add valid URLs
		if utils.IsValidURL(href) {
			links = append(links, href)
		}
	})

	return links
}

// applyContentFilters applies content filters based on the request options.
func (s *scraper) applyContentFilters(doc *goquery.Document) {
	// Extract only main content if requested
	if s.request.OnlyMainContent {
		s.extractMainContent(doc)
	}

	// Include only specific tags if requested
	if len(s.request.IncludeTags) > 0 {
		s.includeOnlyTags(doc, s.request.IncludeTags)
	}

	// Exclude specific tags if requested
	if len(s.request.ExcludeTags) > 0 {
		s.excludeTags(doc, s.request.ExcludeTags)
	}
}

// extractMainContent attempts to extract the main content from the document.
func (s *scraper) extractMainContent(doc *goquery.Document) {
	// Remove common non-content elements
	doc.Find("header, nav, footer, aside, .sidebar, .nav, .menu, .advertisement, script, style, noscript").Remove()

	// Look for common content containers
	mainContent := doc.Find("main, article, .content, .post, .entry, #content, #main, #post")
	if mainContent.Length() > 0 {
		// Replace body with just the main content
		body := doc.Find("body")
		body.Empty()
		body.AppendSelection(mainContent)
	}
}

// includeOnlyTags keeps only the specified tags in the document.
func (s *scraper) includeOnlyTags(doc *goquery.Document, includeTags []string) {
	body := doc.Find("body")
	container := cloneDocument(doc).Find("body")
	container.Empty()

	// Create a selector for all tags to include
	selector := strings.Join(includeTags, ", ")
	body.Find(selector).Each(func(_ int, sel *goquery.Selection) {
		container.AppendSelection(sel)
	})

	// Replace body content with filtered content
	body.Empty()
	body.AppendSelection(container.Children())
}

// excludeTags removes the specified tags from the document.
func (s *scraper) excludeTags(doc *goquery.Document, excludeTags []string) {
	for _, tag := range excludeTags {
		doc.Find(tag).Remove()
	}
}

// cloneDocument is a helper function to clone a goquery document.
func cloneDocument(doc *goquery.Document) *goquery.Document {
	html, err := doc.Html()
	if err != nil {
		doc, _ := goquery.NewDocumentFromReader(strings.NewReader(""))
		return doc
	}

	newDoc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		doc, _ := goquery.NewDocumentFromReader(strings.NewReader(""))
		return doc
	}

	return newDoc
}
