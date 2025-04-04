package scraper

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// cloneDocument creates a deep copy of a goquery document.
func cloneDocument(doc *goquery.Document) *goquery.Document {
	html, err := doc.Html()
	if err != nil {
		newDoc, _ := goquery.NewDocumentFromReader(strings.NewReader(""))
		return newDoc
	}

	newDoc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		newDoc, _ := goquery.NewDocumentFromReader(strings.NewReader(""))
		return newDoc
	}

	return newDoc
}
