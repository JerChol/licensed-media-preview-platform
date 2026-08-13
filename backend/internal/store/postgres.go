package store

import (
	"context"

	"github.com/JerChol/licensed-media-preview-platform/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresStore persists jobs in a real Postgres database.
type PostgresStore struct {
	pool *pgxpool.Pool
}

// NewPostgresStore opens a connection pool against the given DNS
func NewPostgresStore(ctx context.Context, dsn string) (*PostgresStore, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, err
	}
	return &PostgresStore{pool: pool}, nil
}

// SaveJob inserts a new job row. We'll add update logic once the worker needs to change job status later.
func (s *PostgresStore) SaveJob(ctx context.Context, job models.Job) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO jobs (id, source_url, resolved_url, resolved_source_id, status, progress, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, job.ID, job.SourceURL, job.ResolvedURL, job.ResolvedSourceID, job.Status, job.Progress, job.CreatedAt, job.UpdatedAt)
	return err
}

// SaveArtifact inserts a record of a derived artifact for a job
func (s *PostgresStore) SaveArtifact(ctx context.Context, artifact models.Artifact) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO artifacts (id, job_id, artifact_type, storage_path, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`, artifact.ID, artifact.JobID, artifact.ArtifactType, artifact.StoragePath, artifact.CreatedAt)
	return err
}

// ListArtifacts returns every artifact belonging to a job.
func (s *PostgresStore) ListArtifacts(ctx context.Context, jobID string) ([]models.Artifact, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, job_id, artifact_type, storage_path, created_at
		FROM artifacts
		WHERE job_id = $1
	`, jobID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var artifacts []models.Artifact
	for rows.Next() {
		var a models.Artifact
		if err := rows.Scan(&a.ID, &a.JobID, &a.ArtifactType, &a.StoragePath, &a.CreatedAt); err != nil {
			return nil, err
		}
		artifacts = append(artifacts, a)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return artifacts, nil
}

// GetJob fetches a single job by ID
func (s *PostgresStore) GetJob(ctx context.Context, id string) (models.Job, bool, error) {
	var job models.Job
	row := s.pool.QueryRow(ctx, `
		SELECT id, source_url, resolved_url, resolved_source_id, status, progress, created_at, updated_at FROM jobs WHERE id = $1
	`, id)

	err := row.Scan(
		&job.ID, &job.SourceURL, &job.ResolvedURL, &job.ResolvedSourceID, &job.Status, &job.Progress, &job.CreatedAt, &job.UpdatedAt,
	)
	if err != nil {
		return models.Job{}, false, err
	}
	return job, true, nil
}

// ListSources returns every allowlisted source from the database.
func (s *PostgresStore) ListSources(ctx context.Context) ([]models.Source, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT source_id, display_name, license_type, attribution_text, match_hosts, path_prefixes, file_extensions, allowed_media_kind FROM sources
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sources []models.Source
	for rows.Next() {
		var src models.Source
		err := rows.Scan(
			&src.SourceID, &src.DisplayName, &src.LicenseType, &src.AttributionText, &src.MatchHosts, &src.PathPrefixes, &src.FileExtensions, &src.AllowedMediaKind,
		)
		if err != nil {
			return nil, err
		}
		sources = append(sources, src)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return sources, nil
}

// UpdateJobStatus updates a job's status (and optionally an error message) and bumps its updated_at timestamp.
func (s *PostgresStore) UpdateJobStatus(ctx context.Context, id string, status models.JobStatus) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE  jobs
		SET status = $1, updated_at = now()
		WHERE id = $2
	`, status, id)
	return err
}

// GetArtifactStoragePath returns the storage key for a songle artifact by ID.
func (s *PostgresStore) GetArtifactStoragePath(ctx context.Context, artifactID string) (string, error) {
	var path string
	err := s.pool.QueryRow(ctx, `
		SELECT storage_path FROM artifacts WHERE id = $1
	`, artifactID).Scan(&path)
	return path, err
}
