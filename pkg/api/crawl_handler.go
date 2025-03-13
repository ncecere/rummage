package api

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/ncecere/rummage/pkg/model"
)

// handleCrawl handles requests to crawl a website and its subpages.
func (r *Router) handleCrawl(w http.ResponseWriter, req *http.Request) {
	var crawlReq model.CrawlRequest
	if err := json.NewDecoder(req.Body).Decode(&crawlReq); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body: "+err.Error())
		return
	}

	// Validate URL
	if crawlReq.URL == "" {
		respondError(w, http.StatusBadRequest, "URL is required")
		return
	}

	// Create crawl job
	response, jobID, err := r.crawler.Crawl(crawlReq)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to create crawl job: "+err.Error())
		return
	}

	// Store job in Redis
	_, err = r.storage.CreateCrawlJob(jobID, crawlReq)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to store crawl job: "+err.Error())
		return
	}

	// Start processing in background
	go r.crawler.ProcessCrawlJob(jobID, crawlReq)

	// Return job ID and status URL
	respondSuccess(w, response)
}

// handleGetCrawlStatus handles requests to get the status of a crawl job.
func (r *Router) handleGetCrawlStatus(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	jobID := vars["id"]

	if jobID == "" {
		respondError(w, http.StatusBadRequest, "Job ID is required")
		return
	}

	// Get job status
	status, err := r.storage.GetCrawlJob(jobID)
	if err != nil {
		respondError(w, http.StatusNotFound, "Job not found: "+err.Error())
		return
	}

	// Return status
	respondSuccess(w, status)
}

// handleCancelCrawl handles requests to cancel a crawl job.
func (r *Router) handleCancelCrawl(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	jobID := vars["id"]

	if jobID == "" {
		respondError(w, http.StatusBadRequest, "Job ID is required")
		return
	}

	// Cancel job
	err := r.storage.CancelCrawlJob(jobID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to cancel job: "+err.Error())
		return
	}

	// Return status
	respondSuccess(w, map[string]string{"status": "cancelled"})
}

// handleGetCrawlErrors handles requests to get the errors for a crawl job.
func (r *Router) handleGetCrawlErrors(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	jobID := vars["id"]

	if jobID == "" {
		respondError(w, http.StatusBadRequest, "Job ID is required")
		return
	}

	// Get errors
	errors, err := r.storage.GetCrawlErrors(jobID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get errors: "+err.Error())
		return
	}

	// Return errors
	respondSuccess(w, errors)
}
