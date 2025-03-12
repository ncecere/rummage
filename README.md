# Rummage

A simplified Go implementation of the Firecrawl scraping API.

## Overview

Rummage is a web scraping API service built in Go that provides functionality similar to Firecrawl. It allows you to extract content from web pages in various formats including markdown, HTML, and more.

## Features

- **Scrape Endpoint**: Extract content from any URL
- **Batch Scraping**: Process multiple URLs asynchronously
- **Multiple Output Formats**:
  - `markdown`: Convert HTML to markdown (default)
  - `html`: Return processed HTML content
  - `rawHtml`: Return raw HTML content
  - `links`: Extract all links from the page
- **Content Filtering**: Extract only the main content or specific HTML tags
- **Asynchronous Processing**: Process batch jobs in the background
- **Redis Storage**: Store and retrieve batch job results

## Project Structure

The project follows a modular design with clear separation of concerns:

```
rummage/
├── cmd/                  # Application entry points
│   └── rummage/          # Main application
├── pkg/                  # Reusable packages
│   ├── api/              # HTTP API handlers and router
│   ├── config/           # Configuration management
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

By default, the server listens on port 8080. You can change this by setting the `PORT` environment variable.

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

## Environment Variables

- `PORT`: The port to listen on (default: `8080`)
- `REDIS_URL`: The URL of the Redis server (default: `localhost:6379`)
- `BASE_URL`: The base URL of the API (default: `http://localhost:PORT`)
- `DEFAULT_TIMEOUT_MS`: Default request timeout in milliseconds (default: `30000`)
- `DEFAULT_WAIT_TIME_MS`: Default wait time in milliseconds (default: `0`)
- `MAX_CONCURRENT_JOBS`: Maximum number of concurrent batch jobs (default: `10`)
- `JOB_EXPIRATION_HOURS`: Hours until batch jobs expire (default: `24`)

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
