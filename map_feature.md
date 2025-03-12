The /map endpoint is designed to quickly scan a website and return a list of all the URLs it finds. Essentially, it creates a “map” of the site's link structure without crawling every page in depth. The /map endpoint is focused on returning a list of URLs found on the website—it doesn't download or return files (like images, PDFs, or other media).

Key Points
Purpose:
It helps you understand the overall structure of a website by listing all accessible links. This can be especially useful for generating sitemaps, performing site audits, or setting up further targeted scraping jobs.

Usage:
You send a POST request to the endpoint with the URL you want to map. The response will be a JSON object containing an array of URLs that are contextually related to the input URL. 

Optional Parameters:

search: You can provide a search string to filter the returned links to those that contain specific text.
limit: Allows you to restrict the number of links returned (default is 100).
ignoreSitemap: When set to true (default), it ignores the website’s sitemap during mapping.
includeSubdomains: Set this to true if you want to include links from subdomains.

Use Cases
Sitemap Generation: Quickly obtain a full list of pages on a site.
Link Analysis: Identify internal and external links for SEO or content audits.
Targeted

Curl Example:

curl --request POST \
  --url https://api.firecrawl.dev/v1/map \
  --header 'Content-Type: application/json' \
  --data '{
  "url": "<string>",
  "search": "<string>",
  "ignoreSitemap": true,
  "sitemapOnly": false,
  "includeSubdomains": false,
  "limit": 5000,
  "timeout": 123
}'

output:

{
  "success": true,
  "links": [
    "<string>"
  ]
}