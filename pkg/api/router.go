package api

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/ncecere/rummage/pkg/service"
	"github.com/ncecere/rummage/pkg/storage"
)

// RouterOptions contains options for configuring the router
type RouterOptions struct {
	BaseURL  string
	RedisURL string
}

// NewRouter creates and configures a new router
func NewRouter(opts RouterOptions) (*mux.Router, error) {
	r := mux.NewRouter()

	// Create services
	scraperService := service.NewScraperService()

	// Initialize Redis job store
	jobStore, err := storage.NewRedisJobStore(opts.RedisURL)
	if err != nil {
		return nil, err
	}

	// Create batch scraper service
	batchService := service.NewBatchScraperService(scraperService, jobStore, opts.BaseURL)

	// Create handlers
	scraperHandler := NewScraperHandler(scraperService)
	batchHandler := NewBatchScrapeHandler(batchService)

	// API routes
	r.HandleFunc("/v1/scrape", scraperHandler.HandleScrape).Methods("POST")
	r.HandleFunc("/v1/batch/scrape", batchHandler.HandleBatchScrape).Methods("POST")
	r.HandleFunc("/v1/batch/scrape/{id}", batchHandler.HandleGetBatchScrapeStatus).Methods("GET")
	r.HandleFunc("/v1/batch/scrape/{id}/errors", batchHandler.HandleGetBatchScrapeErrors).Methods("GET")

	// Add middleware
	r.Use(loggingMiddleware)

	return r, nil
}

// loggingMiddleware logs all requests
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Log the request
		// log.Printf("%s %s %s", r.RemoteAddr, r.Method, r.URL)

		// Call the next handler
		next.ServeHTTP(w, r)
	})
}
