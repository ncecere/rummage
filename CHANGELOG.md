# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [v0.2.0] - 2025-03-22

### Fixed
- Fixed map endpoint not processing sitemaps when `ignoreSitemap` is false
- Added support for XML namespaces in sitemap parsing
- Added support for sitemap index files
- Added support for gzipped sitemaps
- Added support for finding sitemaps in robots.txt
- Added support for non-standard sitemap formats (plain text lists of URLs)
- Added support for sitemaps at non-standard locations (e.g., `/sitemap` without the .xml extension)
- Added support for sitemaps in the path of the URL
- Improved error handling and logging for sitemap processing

## [v0.1.0] - 2025-03-13

### Added
- Initial release of Rummage
- Scrape endpoint for extracting content from URLs
- Map endpoint for discovering URLs from a starting point
- Crawl endpoint for recursively crawling websites
- Batch scraping functionality for processing multiple URLs asynchronously
- Multiple output formats (markdown, HTML, raw HTML, links)
- Content filtering options
- Redis storage for batch job results
- Docker support
- GitHub Actions workflow for automated releases
- Pre-built binaries for Linux, macOS, and Windows
