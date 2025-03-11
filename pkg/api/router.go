package api

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/ncecere/rummage/pkg/service"
)

// NewRouter creates and configures a new router
func NewRouter() *mux.Router {
	r := mux.NewRouter()

	// Create services
	scraperService := service.NewScraperService()

	// Create handlers
	scraperHandler := NewScraperHandler(scraperService)

	// API routes
	r.HandleFunc("/v1/scrape", scraperHandler.HandleScrape).Methods("POST")

	// Add middleware
	r.Use(loggingMiddleware)

	return r
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
