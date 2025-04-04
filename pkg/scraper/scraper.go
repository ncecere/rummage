package scraper

import (
	"bytes"
	"fmt"
	"net/http"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly/v2"
	"github.com/ncecere/rummage/pkg/model"
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
	c := colly.NewCollector(
		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/109.0.0.0 Safari/537.36"),
	)

	c.SetRequestTimeout(time.Duration(s.request.Timeout) * time.Millisecond)

	if len(s.request.Headers) > 0 {
		c.OnRequest(func(r *colly.Request) {
			for key, value := range s.request.Headers {
				r.Headers.Set(key, value)
			}
		})
	}

	result := &model.ScrapeResult{
		Metadata: &model.ScrapeMetadata{
			SourceURL: s.request.URL,
		},
	}

	if s.request.WaitFor > 0 {
		c.OnRequest(func(r *colly.Request) {
			time.Sleep(time.Duration(s.request.WaitFor) * time.Millisecond)
		})
	}

	c.OnResponse(func(r *colly.Response) {
		result.Metadata.StatusCode = r.StatusCode

		doc, err := goquery.NewDocumentFromReader(bytes.NewReader(r.Body))
		if err != nil {
			return
		}

		result.Metadata.Title = doc.Find("title").Text()
		result.Metadata.Description = doc.Find("meta[name=description]").AttrOr("content", "")
		result.Metadata.Language = doc.Find("html").AttrOr("lang", "")

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

	err := c.Visit(s.request.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to scrape URL: %w", err)
	}

	return result, nil
}
