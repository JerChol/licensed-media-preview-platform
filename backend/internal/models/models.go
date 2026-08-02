package models

import "time"

//Source represents an allowlisted place we're permitted to pull
//media from. Each source defines the rules a submitted URL must match before we'll process it.

type Source struct {
	SourceID         string   `json:"source_id"`
	DisplayName      string   `json:"display_name"`
	LicenseType      string   `json:"license_type"`
	AttributionText  string   `json:"attribution_text"`
	MatchHosts       []string `json:"match_hosts"`
	PathPrefixes     []string `json:"path_prefixes"`
	FileExtensions   []string `json:"file_extensions"`
	AllowedMediaKind string   `json:"allowed_media_kind"`
}

// JobStatus is the set of states a Job can be in as it moves through the pipeline
type JobStatus string

const (
	JobStatusQueued     JobStatus = "queued"
	JobStatusProcessing JobStatus = "processing"
	JobStatusDone                 = "done"
	JobStatusFailed               = "failed"
)

// Job represents a single preview-generation request submitted by a user. It tracks the input they gave us, what we resolved it to, and its current progress through the pipeline.
type Job struct {
	ID               string    `json:"id"`
	SourceURL        string    `json:"source_url"`
	ResolvedURL      string    `json:"resolved_url"`
	ResolvedSourceID string    `json:"resolved_source_id"`
	Status           JobStatus `json:"status"`
	Progress         int       `json:"progress"`
	ErrorMessage     string    `json:"error_message,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// Artifact is a single derived output produced from a Job - a thumbnail, a text snippet, or metadata. We never store the original file itself, only these derived pieces.
type Artifact struct {
	ID           string    `json:"id"`
	JobID        string    `json:"job_id"`
	ArtifactType string    `json:"artifact_type"`
	StoragePath  string    `json:"storage_path"`
	CreatedAt    time.Time `json:"created_at"`
}
