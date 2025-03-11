package service

import (
	"fmt"

	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/PuerkitoBio/goquery"
	"github.com/ncecere/rummage/pkg/models"
	"github.com/ncecere/rummage/pkg/utils"
)

// extractLinks extracts all links from the document
func (s *ScraperService) extractLinks(doc *goquery.Document, response *models.ScrapeResponse) {
	links := []string{}
	doc.Find("a[href]").Each(func(_ int, s *goquery.Selection) {
		if href, exists := s.Attr("href"); exists {
			links = append(links, href)
		}
	})
	response.Data.Links = links
}

// extractHTML extracts HTML content based on request options
func (s *ScraperService) extractHTML(req models.ScrapeRequest, doc *goquery.Document, htmlContent string, response *models.ScrapeResponse) {
	// If onlyMainContent is true, try to extract main content
	if req.OnlyMainContent {
		mainContent := utils.ExtractMainContent(doc)
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

// extractMarkdown converts HTML to markdown based on request options
func (s *ScraperService) extractMarkdown(req models.ScrapeRequest, doc *goquery.Document, htmlContent string, response *models.ScrapeResponse) {
	// Configure the converter with options for cleaner output
	converter := md.NewConverter("", true, nil)

	// Remove unwanted elements before conversion
	doc.Find("style, script, noscript, iframe, nav, footer, header").Remove()

	var mdContent string
	var err error

	if req.OnlyMainContent {
		mainContent := utils.ExtractMainContent(doc)
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
		mdContent = utils.CleanMarkdown(mdContent)
		response.Data.Markdown = mdContent
	}
}
