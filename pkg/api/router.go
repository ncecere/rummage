package api

import (
	"net/http"

	"github.com/gorilla/mux"
)

// RouterOptions contains configuration options for the router
type RouterOptions struct {
	BaseURL  string
	RedisURL string
}

// NewRouter creates and configures a new router for the API
func NewRouter(opts RouterOptions) (*mux.Router, error) {
	router := mux.NewRouter()

	// Create handlers
	scrapeHandler := NewScrapeHandler(opts)
	mapHandler := NewMapHandler(opts)

	// Register routes
	router.HandleFunc("/scrape", scrapeHandler.HandleScrape).Methods(http.MethodPost)
	router.HandleFunc("/map", mapHandler.HandleMap).Methods(http.MethodPost)

	return router, nil
}
