# Rummage

A simplified Go implementation of the Firecrawl scraping API.

## Overview

Rummage is a web scraping API service built in Go that provides functionality similar to Firecrawl. It allows you to extract content from web pages in various formats including markdown, HTML, and more.

## Features

- **Scrape Endpoint**: Extract content from any URL
- **Map Endpoint**: Discover URLs from a starting point using sitemap.xml and HTML links
- **Crawl Endpoint**: Recursively crawl websites and scrape all accessible subpages
- **Batch Scraping**: Process multiple URLs asynchronously
- **Multiple Output Formats**:
  - `markdown`: Convert HTML to markdown (default)
  - `html`: Return processed HTML content
  - `rawHtml`: Return raw HTML content
  - `links`: Extract all links from the page
- **Content Filtering**: Extract only the main content or specific HTML tags
- **Asynchronous Processing**: Process batch jobs in the background
- **Redis Storage**: Store and retrieve batch job results
- **Flexible Configuration**: Configure via YAML files or environment variables

## Project Structure

The project follows a modular design with clear separation of concerns:

```
rummage/
├── cmd/                  # Application entry points
│   └── rummage/          # Main application
├── config/               # Configuration files
│   ├── config.yaml       # Default configuration
│   ├── minimal.yaml      # Minimal configuration example
│   └── full.yaml         # Full configuration example
├── pkg/                  # Reusable packages
│   ├── api/              # HTTP API handlers and router
│   ├── config/           # Configuration management
│   ├── crawler/          # Website crawling functionality
│   ├── model/            # Data models
│   ├── scraper/          # Web scraping functionality
│   ├── storage/          # Data persistence (Redis)
│   └── utils/            # Utility functions
├── Dockerfile            # Docker image definition
├── docker-compose.yml    # Docker Compose configuration
├── docker-compose.test.yml # Test environment configuration
└── Makefile              # Build and development tasks
```

## Getting Started

### Prerequisites

- Go 1.24 or higher
- Redis (for batch processing)

### Installation

```bash
# Clone the repository
git clone https://github.com/ncecere/rummage.git
cd rummage

# Install dependencies
go mod tidy

# Build the project
make build
```

### Running the Server

```bash
# Run directly
make run

# Or with Docker Compose
docker-compose up
```

By default, the server listens on port 8080. You can change this using configuration files or environment variables.

## Configuration

Rummage uses [Viper](https://github.com/spf13/viper) for configuration management, which provides flexibility in how you configure the application.

### Configuration Files

The application looks for a `config.yaml` file in the following locations:
- Current directory (`./`)
- Config directory (`./config/`)
- System config directory (`/etc/rummage/`)
- User home directory (`$HOME/.rummage/`)

### Configuration Options

#### Minimal Configuration Example

```yaml
# Minimal Rummage Configuration
server:
  port: 8080

redis:
  url: redis://localhost:6379
```

#### Full Configuration Example

```yaml
# Full Rummage Configuration

# Server configuration
server:
  # Port to listen on
  port: 8080
  # Base URL for API responses
  baseURL: http://localhost:8080

# Redis configuration
redis:
  # Redis connection URL (redis://host:port)
  url: redis://localhost:6379

# Scraper configuration
scraper:
  # Default timeout for scraping requests in milliseconds
  defaultTimeoutMS: 30000
  # Default wait time before scraping in milliseconds
  defaultWaitTimeMS: 1000
  # Maximum number of concurrent scraping jobs
  maxConcurrentJobs: 10
  # Hours until batch jobs expire
  jobExpirationHours: 24
```

### Environment Variables

All configuration options can also be set using environment variables with the `RUMMAGE_` prefix and using underscores to separate nested keys:

- `RUMMAGE_SERVER_PORT`: The port to listen on (default: `8080`)
- `RUMMAGE_SERVER_BASEURL`: The base URL of the API (default: `http://localhost:PORT`)
- `RUMMAGE_REDIS_URL`: The URL of the Redis server (default: `redis://localhost:6379`)
- `RUMMAGE_SCRAPER_DEFAULTTIMEOUTMS`: Default request timeout in milliseconds (default: `30000`)
- `RUMMAGE_SCRAPER_DEFAULTWAITTIMEMS`: Default wait time in milliseconds (default: `0`)
- `RUMMAGE_SCRAPER_MAXCONCURRENTJOBS`: Maximum number of concurrent batch jobs (default: `10`)
- `RUMMAGE_SCRAPER_JOBEXPIRATIONHOURS`: Hours until batch jobs expire (default: `24`)

Environment variables take precedence over configuration files.

## Development

The project includes several make targets to help with development:

```bash
# Run tests
make test

# Run tests with coverage
make coverage

# Format code
make fmt

# Run linter
make lint

# Clean build artifacts
make clean

# Show all available commands
make help
```

## API Usage

### Scrape Endpoint

```bash
curl --request POST \
  --url http://localhost:8080/v1/scrape \
  --header 'Content-Type: application/json' \
  --data '{
  "url": "https://example.com",
  "formats": ["markdown", "html", "links"],
  "onlyMainContent": true
}'
```

#### Request Parameters

- `url` (required): The URL to scrape
- `formats`: Array of output formats (default: `["markdown"]`)
- `onlyMainContent`: Extract only the main content of the page (default: `true`)
- `includeTags`: Array of HTML tags to include
- `excludeTags`: Array of HTML tags to exclude
- `headers`: Custom HTTP headers for the request
- `waitFor`: Time to wait in milliseconds before scraping
- `timeout`: Request timeout in milliseconds (default: 30000)

#### Response

```json
{
  "success": true,
  "data": {
    "markdown": "...",
    "html": "...",
    "links": ["..."],
    "metadata": {
      "title": "...",
      "description": "...",
      "language": "...",
      "sourceURL": "...",
      "statusCode": 200
    }
  }
}
```

### Map Endpoint

The Map endpoint discovers URLs from a starting point, using both sitemap.xml and HTML link discovery.

```bash
curl --request POST \
  --url http://localhost:8080/v1/map \
  --header 'Content-Type: application/json' \
  --data '{
  "url": "https://example.com",
  "search": "blog",
  "ignoreSitemap": false,
  "sitemapOnly": false,
  "includeSubdomains": false,
  "limit": 100
}'
```

#### Request Parameters

- `url` (required): Starting URL for URL discovery
- `search`: Optional search term to filter URLs
- `ignoreSitemap`: Skip sitemap.xml discovery and only use HTML links
- `sitemapOnly`: Only use sitemap.xml for discovery, ignore HTML links
- `includeSubdomains`: Include URLs from subdomains in results
- `limit`: Maximum number of URLs to return

#### Response

```json
{
  "success": true,
  "data": {
    "links": [
      "https://example.com/page1",
      "https://example.com/page2",
      "https://example.com/blog/post1"
    ]
  }
}
```

### Crawl Endpoint

The Crawl endpoint recursively crawls a website, discovering and scraping all accessible subpages.

```bash
curl --request POST \
  --url http://localhost:8080/v1/crawl \
  --header 'Content-Type: application/json' \
  --data '{
  "url": "https://example.com",
  "excludePaths": ["/admin", "/private"],
  "includePaths": ["/blog", "/docs"],
  "maxDepth": 3,
  "ignoreSitemap": false,
  "ignoreQueryParameters": true,
  "limit": 100,
  "allowBackwardLinks": false,
  "allowExternalLinks": false,
  "scrapeOptions": {
    "formats": ["markdown"],
    "onlyMainContent": true
  }
}'
```

#### Request Parameters

- `url` (required): The URL to crawl
- `excludePaths`: Array of URL paths to exclude from crawling
- `includePaths`: Only crawl these URL paths
- `maxDepth`: Maximum link depth to crawl (default: 10)
- `ignoreSitemap`: Skip sitemap.xml discovery (default: false)
- `ignoreQueryParameters`: Ignore query parameters when comparing URLs (default: false)
- `limit`: Maximum number of pages to crawl (default: 1000)
- `allowBackwardLinks`: Allow crawling links that point to parent directories (default: false)
- `allowExternalLinks`: Allow crawling links to external domains (default: false)
- `scrapeOptions`: Options for scraping each page (same as Scrape endpoint)

#### Response

```json
{
  "success": true,
  "id": "job-id",
  "url": "http://localhost:8080/v1/crawl/job-id"
}
```

### Get Crawl Status

```bash
curl --request GET \
  --url http://localhost:8080/v1/crawl/job-id
```

#### Response

```json
{
  "status": "scraping",
  "total": 36,
  "completed": 10,
  "expiresAt": "2025-03-11T10:36:14Z",
  "data": [
    {
      "markdown": "...",
      "html": "...",
      "links": ["..."],
      "metadata": {
        "title": "...",
        "description": "...",
        "language": "...",
        "sourceURL": "...",
        "statusCode": 200
      }
    }
  ]
}
```

### Cancel Crawl

```bash
curl --request DELETE \
  --url http://localhost:8080/v1/crawl/job-id
```

#### Response

```json
{
  "status": "cancelled"
}
```

### Get Crawl Errors

```bash
curl --request GET \
  --url http://localhost:8080/v1/crawl/job-id/errors
```

#### Response

```json
{
  "errors": [
    {
      "id": "error-id",
      "timestamp": "2025-03-11T10:36:14Z",
      "url": "https://example.com/broken-page",
      "error": "Failed to scrape URL: 404 Not Found"
    }
  ],
  "robotsBlocked": [
    "https://example.com/robots-blocked-page"
  ]
}
```

### Batch Scrape Endpoint

```bash
curl --request POST \
  --url http://localhost:8080/v1/batch/scrape \
  --header 'Content-Type: application/json' \
  --data '{
  "urls": ["https://example.com", "https://example.org"],
  "formats": ["markdown", "html", "links"],
  "onlyMainContent": true,
  "ignoreInvalidURLs": true
}'
```

#### Request Parameters

- `urls` (required): Array of URLs to scrape
- `formats`: Array of output formats (default: `["markdown"]`)
- `onlyMainContent`: Extract only the main content of the page (default: `true`)
- `includeTags`: Array of HTML tags to include
- `excludeTags`: Array of HTML tags to exclude
- `headers`: Custom HTTP headers for the request
- `waitFor`: Time to wait in milliseconds before scraping
- `timeout`: Request timeout in milliseconds (default: 30000)
- `ignoreInvalidURLs`: Whether to ignore invalid URLs (default: `false`)
- `webhook`: Webhook configuration for notifications

#### Response

```json
{
  "success": true,
  "id": "job-id",
  "url": "http://localhost:8080/v1/batch/scrape/job-id",
  "invalidURLs": ["invalid-url"]
}
```

### Get Batch Scrape Status

```bash
curl --request GET \
  --url http://localhost:8080/v1/batch/scrape/job-id
```

#### Response

```json
{
  "status": "completed",
  "total": 2,
  "completed": 2,
  "creditsUsed": 2,
  "expiresAt": "2025-03-11T10:36:14Z",
  "data": [
    {
      "markdown": "...",
      "html": "...",
      "links": ["..."],
      "metadata": {
        "title": "...",
        "description": "...",
        "language": "...",
        "sourceURL": "...",
        "statusCode": 200
      }
    },
    {
      "markdown": "...",
      "html": "...",
      "links": ["..."],
      "metadata": {
        "title": "...",
        "description": "...",
        "language": "...",
        "sourceURL": "...",
        "statusCode": 200
      }
    }
  ]
}
```

## Docker Support

The project includes Docker support for easy deployment:

```bash
# Build and run with Docker Compose
docker-compose up --build

# Run in detached mode
docker-compose up -d

# Stop containers
docker-compose down
```

The Docker configuration includes:
- Mounting the config directory for easy configuration
- Setting environment variables for customization
- Connecting to a Redis container for batch processing

## Testing

The project includes unit tests for all packages:

```bash
# Run all tests
make test

# Run tests with coverage
make coverage
```

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Acknowledgments

- Inspired by [Firecrawl](https://firecrawl.dev)
- Uses [Viper](https://github.com/spf13/viper) for configuration management
