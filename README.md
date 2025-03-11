# Rummage

A simplified Go implementation of the Firecrawl scraping API.

## Overview

Rummage is a web scraping API service built in Go that provides functionality similar to Firecrawl. It allows you to extract content from web pages in various formats including markdown, HTML, and more.

## Features

- **Scrape Endpoint**: Extract content from any URL
- **Multiple Output Formats**:
  - `markdown`: Convert HTML to markdown (default)
  - `html`: Return processed HTML content
  - `rawHtml`: Return raw HTML content
  - `links`: Extract all links from the page

## Getting Started

### Prerequisites

- Go 1.16 or higher

### Installation

```bash
# Clone the repository
git clone https://github.com/ncecere/rummage.git
cd rummage

# Install dependencies
go mod tidy

# Build the project
go build -o rummage ./cmd/rummage
```

### Running the Server

```bash
# Run the server
./rummage
```

By default, the server listens on port 8080. You can change this by setting the `PORT` environment variable.

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

## Using Redis for Job Storage

The batch scrape endpoint uses Redis to store job information. By default, the application will try to connect to Redis at `localhost:6379`. You can customize the Redis URL by setting the `REDIS_URL` environment variable.

For testing purposes, you can use the included Docker Compose file:

```bash
docker-compose -f docker-compose.test.yml up -d
```

This will start a Redis instance that the application can use for job storage.

## Environment Variables

- `PORT`: The port to listen on (default: `8080`)
- `REDIS_URL`: The URL of the Redis server (default: `localhost:6379`)
- `BASE_URL`: The base URL of the API (default: `http://localhost:PORT`)

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Acknowledgments

- Inspired by [Firecrawl](https://firecrawl.dev)
