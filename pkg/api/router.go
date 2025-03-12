// Package api provides HTTP API functionality for the Rummage service.
package api

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/ncecere/rummage/pkg/model"
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

	// Create router instance
	r := &Router{
		Router:  mux.NewRouter(),
		scraper: scraperService,
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
}

// handleHealth is a simple health check endpoint.
func (r *Router) handleHealth(w http.ResponseWriter, req *http.Request) {
	respondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// handleScrape handles requests to scrape a single URL.
func (r *Router) handleScrape(w http.ResponseWriter, req *http.Request) {
	var scrapeReq model.ScrapeRequest
	if err := json.NewDecoder(req.Body).Decode(&scrapeReq); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body: "+err.Error())
		return
	}

	// Validate URL
	if scrapeReq.URL == "" {
		respondError(w, http.StatusBadRequest, "URL is required")
		return
	}

	// Perform scrape
	result, err := r.scraper.Scrape(scrapeReq)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to scrape URL: "+err.Error())
		return
	}

	// Return result
	respondSuccess(w, result)
}

// handleBatchScrape handles requests to scrape multiple URLs.
func (r *Router) handleBatchScrape(w http.ResponseWriter, req *http.Request) {
	var batchReq model.BatchScrapeRequest
	if err := json.NewDecoder(req.Body).Decode(&batchReq); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body: "+err.Error())
		return
	}

	// Validate URLs
	validURLs, invalidURLs, err := r.scraper.BatchScrape(batchReq)
	if err != nil && !batchReq.IgnoreInvalidURLs {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Create batch job
	jobID, err := r.storage.CreateBatchJob(validURLs, invalidURLs)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to create batch job: "+err.Error())
		return
	}

	// Start processing in background
	go r.scraper.ProcessBatchJob(jobID, validURLs, batchReq, r.storage.UpdateBatchJob)

	// Return job ID and status URL
	respondSuccess(w, model.BatchScrapeResponse{
		ID:          jobID,
		URL:         r.baseURL + "/v1/batch/scrape/" + jobID,
		InvalidURLs: invalidURLs,
	})
}

// handleGetBatchStatus handles requests to get the status of a batch job.
func (r *Router) handleGetBatchStatus(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	jobID := vars["id"]

	if jobID == "" {
		respondError(w, http.StatusBadRequest, "Job ID is required")
		return
	}

	// Get job status
	status, err := r.storage.GetBatchJob(jobID)
	if err != nil {
		respondError(w, http.StatusNotFound, "Job not found: "+err.Error())
		return
	}

	// Return status
	respondSuccess(w, status)
}
