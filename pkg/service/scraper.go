package service

import (
	"fmt"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly/v2"
	"github.com/ncecere/rummage/pkg/models"
	"github.com/ncecere/rummage/pkg/utils"
)

// ScraperService handles web scraping operations
type ScraperService struct{}

// NewScraperService creates a new scraper service
func NewScraperService() *ScraperService {
	return &ScraperService{}
}

// Scrape performs the web scraping operation based on the request
func (s *ScraperService) Scrape(req models.ScrapeRequest) (models.ScrapeResponse, error) {
	// Initialize the response
	response := models.ScrapeResponse{
		Success: true,
		Data: models.ScrapeData{
			Metadata: models.Metadata{
				SourceURL: req.URL,
			},
		},
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
		response.Data.Metadata.Description = utils.GetMetaContent(doc, "description")
		response.Data.Metadata.Language = doc.Find("html").AttrOr("lang", "")

		// Process the document based on requested formats
		s.processFormats(req, doc, htmlContent, &response)
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
		return response, err
	}

	return response, nil
}

// processFormats handles the different output formats requested
func (s *ScraperService) processFormats(req models.ScrapeRequest, doc *goquery.Document, htmlContent string, response *models.ScrapeResponse) {
	// Extract links if requested
	if utils.Contains(req.Formats, "links") {
		s.extractLinks(doc, response)
	}

	// Include HTML if requested
	if utils.Contains(req.Formats, "html") {
		s.extractHTML(req, doc, htmlContent, response)
	}

	// Include raw HTML if requested
	if utils.Contains(req.Formats, "rawHtml") {
		response.Data.RawHTML = htmlContent
	}

	// Convert to markdown if requested
	if utils.Contains(req.Formats, "markdown") {
		s.extractMarkdown(req, doc, htmlContent, response)
	}
}
