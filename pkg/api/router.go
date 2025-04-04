// Package api provides HTTP API functionality for the Rummage service.
package api

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/ncecere/rummage/pkg/crawler"
	"github.com/ncecere/rummage/pkg/scraper"
	"github.com/ncecere/rummage/pkg/storage"
)

// RouterOptions contains configuration options for the API router.
type RouterOptions struct {
	BaseURL  string
	RedisURL string
}

// Router represents the API router with its dependencies.
type Router struct {
	*mux.Router
	scraper *scraper.Service
	crawler *crawler.Service
	storage *storage.RedisStorage
	baseURL string
}

// NewRouter creates and configures a new API router.
func NewRouter(opts RouterOptions) (*mux.Router, error) {
	// Initialize storage
	redisStorage, err := storage.NewRedisStorage(opts.RedisURL)
	if err != nil {
		return nil, err
	}

	// Initialize scraper service
	scraperService := scraper.NewService()

	// Initialize crawler service
	crawlerService := crawler.NewService(crawler.ServiceOptions{
		BaseURL:           opts.BaseURL,
		UpdateJobFn:       redisStorage.UpdateCrawlJob,
		UpdateJobStatusFn: redisStorage.UpdateCrawlJobStatus,
	})

	// Create router instance
	r := &Router{
		Router:  mux.NewRouter(),
		scraper: scraperService,
		crawler: crawlerService,
		storage: redisStorage,
		baseURL: opts.BaseURL,
	}

	// Register routes
	r.registerRoutes()

	return r.Router, nil
}

// registerRoutes sets up all API routes.
func (r *Router) registerRoutes() {
	// API version prefix
	api := r.PathPrefix("/v1").Subrouter()

	// Health check endpoint
	api.HandleFunc("/health", r.handleHealth).Methods(http.MethodGet)

	// Scrape endpoints
	api.HandleFunc("/scrape", r.handleScrape).Methods(http.MethodPost)
	api.HandleFunc("/batch/scrape", r.handleBatchScrape).Methods(http.MethodPost)
	api.HandleFunc("/batch/scrape/{id}", r.handleGetBatchStatus).Methods(http.MethodGet)

	// Crawl endpoints
	api.HandleFunc("/crawl", r.handleCrawl).Methods(http.MethodPost)
	api.HandleFunc("/crawl/{id}", r.handleGetCrawlStatus).Methods(http.MethodGet)
	api.HandleFunc("/crawl/{id}", r.handleCancelCrawl).Methods(http.MethodDelete)
	api.HandleFunc("/crawl/{id}/errors", r.handleGetCrawlErrors).Methods(http.MethodGet)

	// Map endpoints
	api.HandleFunc("/map", r.handleMap).Methods(http.MethodPost)
}
