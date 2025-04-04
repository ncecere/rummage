package api

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/ncecere/rummage/pkg/model"
)

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
