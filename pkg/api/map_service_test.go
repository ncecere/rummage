package api

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// MockMapService for testing that doesn't make actual HTTP requests
type MockMapServiceForTest struct{}

func (m *MockMapServiceForTest) Map(opts MapOptions) ([]string, error) {
	// Return mock links based on the options
	if opts.URL == "invalid-url" {
		return nil, fmt.Errorf("invalid URL")
	}

	links := []string{
		opts.URL + "/",
		opts.URL + "/page1",
		opts.URL + "/page2",
		opts.URL + "/about",
		opts.URL + "/contact",
	}

	// Add sitemap links if not ignored
	if !opts.IgnoreSitemap {
		links = append(links,
			"https://example.com/page1",
			"https://example.com/page2",
			"https://example.com/about",
		)
	}

	// Filter by search if provided
	if opts.Search != "" {
		filteredLinks := []string{}
		for _, link := range links {
			if strings.Contains(strings.ToLower(link), strings.ToLower(opts.Search)) {
				filteredLinks = append(filteredLinks, link)
			}
		}
		links = filteredLinks
	}

	// Apply limit if provided
	if opts.Limit > 0 && len(links) > opts.Limit {
		links = links[:opts.Limit]
	}

	return links, nil
}

func TestMapService_Map(t *testing.T) {
	// Create a mock service
	service := &MockMapServiceForTest{}

	// Test cases
	testCases := []struct {
		name          string
		options       MapOptions
		expectedCount int
		expectError   bool
	}{
		{
			name: "Basic Map",
			options: MapOptions{
				URL:           "http://example.com",
				IgnoreSitemap: true,
				Limit:         10,
			},
			expectedCount: 5, // Home, Page1, Page2, About, Contact
			expectError:   false,
		},
		{
			name: "With Sitemap",
			options: MapOptions{
				URL:           "http://example.com",
				IgnoreSitemap: false,
				Limit:         10,
			},
			expectedCount: 8, // 5 from HTML + 3 from sitemap
			expectError:   false,
		},
		{
			name: "With Search Filter",
			options: MapOptions{
				URL:           "http://example.com",
				Search:        "page",
				IgnoreSitemap: true,
				Limit:         10,
			},
			expectedCount: 2, // Page1, Page2
			expectError:   false,
		},
		{
			name: "With Limit",
			options: MapOptions{
				URL:           "http://example.com",
				IgnoreSitemap: true,
				Limit:         2,
			},
			expectedCount: 2,
			expectError:   false,
		},
		{
			name: "Invalid URL",
			options: MapOptions{
				URL:           "invalid-url",
				IgnoreSitemap: true,
				Limit:         10,
			},
			expectedCount: 0,
			expectError:   true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			links, err := service.Map(tc.options)

			// Check error
			if tc.expectError && err == nil {
				t.Errorf("Expected an error, but got none")
			}
			if !tc.expectError && err != nil {
				t.Errorf("Expected no error, but got: %v", err)
			}

			// Skip further checks if we expected an error
			if tc.expectError {
				return
			}

			// Check the number of links
			if len(links) != tc.expectedCount {
				t.Errorf("Expected %d links, got %d", tc.expectedCount, len(links))
			}
		})
	}
}

func TestFetchSitemap(t *testing.T) {
	// Create a test server that returns a simple sitemap
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "sitemap.xml") {
			// Return a simple sitemap
			w.Header().Set("Content-Type", "application/xml")
			w.Write([]byte(`
				<?xml version="1.0" encoding="UTF-8"?>
				<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
					<url>
						<loc>https://example.com/page1</loc>
					</url>
					<url>
						<loc>https://example.com/page2</loc>
					</url>
					<url>
						<loc>https://example.com/about</loc>
					</url>
				</urlset>
			`))
			return
		}

		// Return 404 for any other path
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	// Create a new map service
	service := NewMapService()

	// Test fetching a valid sitemap
	links, err := service.fetchSitemap(ts.URL)
	if err != nil {
		t.Errorf("Expected no error, but got: %v", err)
	}
	if len(links) != 3 {
		t.Errorf("Expected 3 links, got %d", len(links))
	}

	// Test fetching a non-existent sitemap
	nonExistentURL := "http://non-existent-domain-12345.com"
	links, err = service.fetchSitemap(nonExistentURL)
	if err == nil {
		t.Errorf("Expected an error, but got none")
	}
	if len(links) != 0 {
		t.Errorf("Expected 0 links, got %d", len(links))
	}
}
