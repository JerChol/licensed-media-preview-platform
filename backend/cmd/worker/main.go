package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/JerChol/licensed-media-preview-platform/internal/config"
	"github.com/JerChol/licensed-media-preview-platform/internal/models"
	"github.com/JerChol/licensed-media-preview-platform/internal/pdfpipeline"
	"github.com/JerChol/licensed-media-preview-platform/internal/queue"
	"github.com/JerChol/licensed-media-preview-platform/internal/storage"
	"github.com/JerChol/licensed-media-preview-platform/internal/store"
	"github.com/google/uuid"
)

func main() {
	ctx := context.Background()

	pgStore, err := store.NewPostgresStore(ctx, config.PostgresDSN())
	if err != nil {
		log.Fatalf("failed to connect to postgres: %v", err)
	}

	jobQueue := queue.NewRedisQueue("localhost:6379")

	log.Println("Worker started, waiting for jobs...")

	for {
		jobID, err := jobQueue.Dequeue(ctx)
		if err != nil {
			log.Printf("Dequeue error: %v", err)
			continue
		}
		log.Printf("Picked up jobs: %s", jobID)

		// Actual processing (PDF pipeline)
		processJob(ctx, pgStore, jobID)
	}
}

// processJob handles a single job. For now this just marks it as processing, then done - real PDF work comes in the next step
func processJob(ctx context.Context, s *store.PostgresStore, jobID string) {
	if err := s.UpdateJobStatus(ctx, jobID, models.JobStatusProcessing); err != nil {
		log.Printf("failed to mark job %s as processing: %v", jobID, err)
		return
	}

	job, ok, err := s.GetJob(ctx, jobID)
	if err != nil || !ok {
		log.Printf("failed to load job %s: %v", jobID, err)
		s.UpdateJobStatus(ctx, jobID, models.JobStatusFailed)
		return
	}

	localPath, err := pdfpipeline.DownloadToTemp(job.ResolvedURL)
	if err != nil {
		log.Printf("download failed for job %s: %v", jobID, err)
		s.UpdateJobStatus(ctx, jobID, models.JobStatusFailed)
		return
	}
	defer os.Remove(localPath)

	result, err := pdfpipeline.Extract(localPath)
	if err != nil {
		log.Printf("extract failed for job %s: %v", jobID, err)
		s.UpdateJobStatus(ctx, jobID, models.JobStatusFailed)
		return
	}

	snippetPath, err := storage.SaveTextSnippet(jobID, result.TextSnippet)
	if err != nil {
		log.Printf("failed to save snippet for job %s: %v", jobID, err)
		s.UpdateJobStatus(ctx, jobID, models.JobStatusFailed)
		return
	}

	log.Printf("job %s: extracted %d pages, snippet saved to %s", jobID, result.PageCount, snippetPath)

	artifact := models.Artifact{
		ID:           uuid.NewString(),
		JobID:        jobID,
		ArtifactType: "text_snippet",
		StoragePath:  snippetPath,
		CreatedAt:    time.Now(),
	}
	if err := s.SaveArtifact(ctx, artifact); err != nil {
		log.Printf("failed to save artifact record for job %s: %v", jobID, err)
		s.UpdateJobStatus(ctx, jobID, models.JobStatusFailed)
		return
	}

	if err := s.UpdateJobStatus(ctx, jobID, models.JobStatusDone); err != nil {
		log.Printf("failed to mark job %s as done: %v", jobID, err)
		return
	}

	log.Printf("job %s marked done", jobID)
}
