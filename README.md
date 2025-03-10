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
  - `json`: Extract structured data (coming soon)
  - `screenshot`: Capture screenshot of the page (coming soon)
  - `screenshot@fullPage`: Capture full page screenshot (coming soon)

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

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Acknowledgments

- Inspired by [Firecrawl](https://firecrawl.dev)
