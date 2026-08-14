package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
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

	s3Cfg := config.LoadS3Config()
	s3Store, err := storage.NewS3Storage(ctx, s3Cfg.BucketName, s3Cfg.Region)
	if err != nil {
		log.Fatalf("failed to set up s3 storage: %v", err)
	}

	redisCfg := config.LoadRedisConfig()
	jobQueue := queue.NewRedisQueue(redisCfg.Addr, redisCfg.Password, redisCfg.UseTLS)

	go startHealthServer()

	log.Println("Worker started, waiting for jobs...")

	for {
		jobID, err := jobQueue.Dequeue(ctx)
		if err != nil {
			log.Printf("Dequeue error: %v", err)
			continue
		}
		log.Printf("Picked up jobs: %s", jobID)

		// Actual processing (PDF pipeline)
		processJob(ctx, pgStore, s3Store, jobID)
	}
}

// processJob handles a single job. For now this just marks it as processing, then done - real PDF work comes in the next step
func processJob(ctx context.Context, s *store.PostgresStore, s3s *storage.S3Storage, jobID string) {
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

	snippetKey := fmt.Sprintf("jobs/%s/snippet.txt", jobID)
	if err := s3s.UploadBytes(ctx, snippetKey, []byte(result.TextSnippet), "text/plain"); err != nil {
		log.Printf("failed to upload snippet for job %s: %v", jobID, err)
		s.UpdateJobStatus(ctx, jobID, models.JobStatusFailed)
		return
	}

	log.Printf("job %s: extracted %d pages, snippet saved to %s", jobID, result.PageCount, snippetKey)

	artifact := models.Artifact{
		ID:           uuid.NewString(),
		JobID:        jobID,
		ArtifactType: "text_snippet",
		StoragePath:  snippetKey,
		CreatedAt:    time.Now(),
	}
	if err := s.SaveArtifact(ctx, artifact); err != nil {
		log.Printf("failed to save artifact record for job %s: %v", jobID, err)
		s.UpdateJobStatus(ctx, jobID, models.JobStatusFailed)
		return
	}

	thumbLocalPrefix := filepath.Join(os.TempDir(), jobID+"-thumbnail")
	thumbLocalPath, err := pdfpipeline.GenerateThumbnail(localPath, thumbLocalPrefix)
	if err != nil {
		log.Printf("thumbnail generation failed for job %s: %v", jobID, err)
		// Not fatal - we still have the text snippet, so don't fail the whole job
	} else {
		defer os.Remove(thumbLocalPath)

		thumbKey := fmt.Sprintf("jobs/%s/thumbnail.jpg", jobID)
		if err := s3s.UploadFile(ctx, thumbLocalPath, thumbKey, "image/jpeg"); err != nil {
			log.Printf("failed to upload thumbnail for job %s: %v", jobID, err)
		} else {
			thumbArtifact := models.Artifact{
				ID:           uuid.NewString(),
				JobID:        jobID,
				ArtifactType: "thumbnail",
				StoragePath:  thumbKey,
				CreatedAt:    time.Now(),
			}
			if err := s.SaveArtifact(ctx, thumbArtifact); err != nil {
				log.Printf("failed to save thumbnail artifact for job %s: %v", jobID, err)
			}
		}
	}

	if err := s.UpdateJobStatus(ctx, jobID, models.JobStatusDone); err != nil {
		log.Printf("failed to mark job %s as done: %v", jobID, err)
		return
	}

	log.Printf("job %s marked done", jobID)
}

// startHealthServer runs a minimal HTTP server so Render(hosting provider) treats this as a valid web service. It has no real functionality - it exists purely to satisfy Render's health checks and give an external uptime pinger something to hit.
func startHealthServer() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("worker is running"))
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Health check server listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
