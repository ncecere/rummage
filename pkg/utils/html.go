package utils

import (
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// Contains checks if a slice contains a string
func Contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// GetMetaContent extracts meta content from HTML document
func GetMetaContent(doc *goquery.Document, name string) string {
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

// ExtractMainContent finds the main content section of an HTML document
func ExtractMainContent(doc *goquery.Document) string {
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

// CleanMarkdown improves markdown output by removing redundancies and formatting issues
func CleanMarkdown(input string) string {
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
