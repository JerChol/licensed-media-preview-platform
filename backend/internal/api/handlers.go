package api

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/JerChol/licensed-media-preview-platform/internal/models"
	"github.com/JerChol/licensed-media-preview-platform/internal/queue"
	"github.com/JerChol/licensed-media-preview-platform/internal/resolver"
	"github.com/JerChol/licensed-media-preview-platform/internal/store"
	"github.com/google/uuid"
)

// Server bundles the dependencies our HTTP handlers need
type Server struct {
	Store   *store.PostgresStore
	Sources []models.Source
	Queue   *queue.RedisQueue
}

// createJobRequest is the shape we expect in the POST / jobs body
type createJobRequest struct {
	URL string `json:"url"`
}

// HandleCreateJob handles POST /jobs
func (s *Server) HandleCreateJob(w http.ResponseWriter, r *http.Request) {
	var req createJobRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	result, resolveErr := resolver.Resolve(req.URL, s.Sources)
	if resolveErr != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(w).Encode(map[string]string{
			"error_code": string(resolveErr.Code),
			"message":    resolveErr.Message,
		})
		return
	}

	job := models.Job{
		ID:               uuid.NewString(),
		SourceURL:        req.URL,
		ResolvedURL:      result.ResolvedFileURL,
		ResolvedSourceID: result.SourceID,
		Status:           models.JobStatusQueued,
		Progress:         0,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}
	if err := s.Store.SaveJob(r.Context(), job); err != nil {
		log.Printf("SaveJob failed: %v", err)
		http.Error(w, `"error":"failed to save job"`, http.StatusInternalServerError)
		return
	}
	if err := s.Queue.Enqueue(r.Context(), job.ID); err != nil {
		log.Printf("Enqueue failed: %v", err)
		http.Error(w, `{"error":"failed to enqueue job"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(job)
}

// HandeGetJob handles GET /jobs/{id}.
func (s *Server) HandleGetJob(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	job, ok, err := s.Store.GetJob(r.Context(), id)
	if err != nil {
		http.Error(w, `{"error":"failed to fetch job"}`, http.StatusInternalServerError)
		return
	}
	if !ok {
		http.Error(w, `{"error":"job not found"}`, http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(job)
}

// HandleGetJobArtifacts handles GET /jobs/{id}/artifacts.
func (s *Server) HandleGetJobArtifacts(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	// Confirm the job itself exists first, so we return 404 rather than an empty list for a job ID that was never valid.
	_, ok, err := s.Store.GetJob(r.Context(), id)
	if err != nil {
		http.Error(w, `{"error":"failed to fetch job"}`, http.StatusInternalServerError)
		return
	}
	if !ok {
		http.Error(w, `{"error":"job not found"}`, http.StatusNotFound)
		return
	}

	artifacts, err := s.Store.ListArtifacts(r.Context(), id)
	if err != nil {
		http.Error(w, `{"error":"failed to fetch artifacts"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(artifacts)
}
