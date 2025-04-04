package scraper

import (
	html2md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/PuerkitoBio/goquery"
)

// extractMarkdown extracts markdown content from the document.
func (s *scraper) extractMarkdown(doc *goquery.Document) string {
	docCopy := cloneDocument(doc)
	s.applyContentFilters(docCopy)

	converter := html2md.NewConverter("", true, nil)

	html, err := docCopy.Html()
	if err != nil {
		return ""
	}

	markdown, err := converter.ConvertString(html)
	if err != nil {
		return ""
	}

	return markdown
}

// extractHTML extracts processed HTML content from the document.
func (s *scraper) extractHTML(doc *goquery.Document) string {
	docCopy := cloneDocument(doc)
	s.applyContentFilters(docCopy)

	html, err := docCopy.Html()
	if err != nil {
		return ""
	}
	return html
}

// extractLinks extracts all links from the document.
func (s *scraper) extractLinks(doc *goquery.Document) []string {
	links := make([]string, 0)

	doc.Find("a[href]").Each(func(_ int, sel *goquery.Selection) {
		href, exists := sel.Attr("href")
		if !exists || href == "" || href[0] == '#' {
			return
		}
		links = append(links, href)
	})

	return links
}

// applyContentFilters applies content filters based on the request options.
func (s *scraper) applyContentFilters(doc *goquery.Document) {
	if s.request.OnlyMainContent {
		s.extractMainContent(doc)
	}
	if len(s.request.IncludeTags) > 0 {
		s.includeOnlyTags(doc, s.request.IncludeTags)
	}
	if len(s.request.ExcludeTags) > 0 {
		s.excludeTags(doc, s.request.ExcludeTags)
	}
}

// extractMainContent attempts to extract the main content from the document.
func (s *scraper) extractMainContent(doc *goquery.Document) {
	doc.Find("header, nav, footer, aside, .sidebar, .nav, .menu, .advertisement, script, style, noscript").Remove()

	mainContent := doc.Find("main, article, .content, .post, .entry, #content, #main, #post")
	if mainContent.Length() > 0 {
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

	selector := ""
	for i, tag := range includeTags {
		if i > 0 {
			selector += ", "
		}
		selector += tag
	}

	body.Find(selector).Each(func(_ int, sel *goquery.Selection) {
		container.AppendSelection(sel)
	})

	body.Empty()
	body.AppendSelection(container.Children())
}

// excludeTags removes the specified tags from the document.
func (s *scraper) excludeTags(doc *goquery.Document, excludeTags []string) {
	for _, tag := range excludeTags {
		doc.Find(tag).Remove()
	}
}
