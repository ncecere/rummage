package api

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/ncecere/rummage/pkg/models"
	"github.com/ncecere/rummage/pkg/service"
)

// BatchScrapeHandler handles batch scrape requests
type BatchScrapeHandler struct {
	batchService *service.BatchScraperService
}

// NewBatchScrapeHandler creates a new batch scrape handler
func NewBatchScrapeHandler(batchService *service.BatchScraperService) *BatchScrapeHandler {
	return &BatchScrapeHandler{
		batchService: batchService,
	}
}

// HandleBatchScrape handles POST requests to /v1/batch/scrape
func (h *BatchScrapeHandler) HandleBatchScrape(w http.ResponseWriter, r *http.Request) {
	// Parse the request body
	var req models.BatchScrapeRequest
	if !DecodeJSONBody(w, r, &req) {
		return
	}

	// Validate the request
	if len(req.URLs) == 0 {
		WriteErrorResponse(w, "No URLs provided", http.StatusBadRequest)
		return
	}

	// Set default formats if not provided
	if len(req.Formats) == 0 {
		req.Formats = []string{"markdown"}
	}

	// Process the batch scrape request
	resp, err := h.batchService.BatchScrape(r.Context(), req)
	if err != nil {
		WriteServiceError(w, err)
		return
	}

	// Return the response
	if err := WriteJSON(w, http.StatusOK, resp); err != nil {
		WriteServiceError(w, err)
	}
}

// HandleGetBatchScrapeStatus handles GET requests to /v1/batch/scrape/{id}
func (h *BatchScrapeHandler) HandleGetBatchScrapeStatus(w http.ResponseWriter, r *http.Request) {
	// Get the job ID from the URL
	vars := mux.Vars(r)
	jobID, ok := GetMuxVar(w, vars, "id")
	if !ok {
		return
	}

	// Get the job status
	resp, err := h.batchService.GetBatchScrapeStatus(r.Context(), jobID)
	if err != nil {
		WriteServiceError(w, err)
		return
	}

	// Return the response
	if err := WriteJSON(w, http.StatusOK, resp); err != nil {
		WriteServiceError(w, err)
	}
}

// HandleGetBatchScrapeErrors handles GET requests to /v1/batch/scrape/{id}/errors
func (h *BatchScrapeHandler) HandleGetBatchScrapeErrors(w http.ResponseWriter, r *http.Request) {
	// Get the job ID from the URL
	vars := mux.Vars(r)
	jobID, ok := GetMuxVar(w, vars, "id")
	if !ok {
		return
	}

	// Get the job errors
	resp, err := h.batchService.GetBatchScrapeErrors(r.Context(), jobID)
	if err != nil {
		WriteServiceError(w, err)
		return
	}

	// Return the response
	if err := WriteJSON(w, http.StatusOK, resp); err != nil {
		WriteServiceError(w, err)
	}
}
