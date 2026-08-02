package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/JerChol/licensed-media-preview-platform/internal/models"
	"github.com/JerChol/licensed-media-preview-platform/internal/resolver"
	"github.com/JerChol/licensed-media-preview-platform/internal/store"
	"github.com/google/uuid"
)

// Server bundles the dependencies our HTTP handlers need
type Server struct {
	Store   *store.MemoryStore
	Sources []models.Source
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
	s.Store.SaveJob(job)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(job)
}

// HandeGetJob handles GET /jobs/{id}.
func (s *Server) HandleGetJob(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	job, ok := s.Store.GetJob(id)
	if !ok {
		http.Error(w, `{"error":"job not found"}`, http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(job)
}
