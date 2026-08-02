package store

import (
	"sync"

	"github.com/JerChol/licensed-media-preview-platform/internal/models"
)

// MemoryStore holds jobs in memory only - data disappears when the process restarts. Good enough for local testing before we wire us Postgres.
type MemoryStore struct {
	mu   sync.Mutex
	jobs map[string]models.Job
}

// NewMemoryStore creates an empty store, ready to use.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		jobs: make(map[string]models.Job),
	}
}

// SaveJob adds or updates a job by its ID.
func (s *MemoryStore) SaveJob(job models.Job) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.jobs[job.ID] = job
}

// GetJob retrieves a job by ID. The second return value reports whether it was found, morroring Go's map lookup convention.
func (s *MemoryStore) GetJob(id string) (models.Job, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	job, ok := s.jobs[id]
	return job, ok
}
