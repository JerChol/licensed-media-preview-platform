package main

import (
	"context"
	"log"
	"net/http"

	"github.com/JerChol/licensed-media-preview-platform/internal/api"
	"github.com/JerChol/licensed-media-preview-platform/internal/config"
	"github.com/JerChol/licensed-media-preview-platform/internal/queue"
	"github.com/JerChol/licensed-media-preview-platform/internal/storage"
	"github.com/JerChol/licensed-media-preview-platform/internal/store"
)

func main() {
	ctx := context.Background()

	redisCfg := config.LoadRedisConfig()
	jobQueue := queue.NewRedisQueue(redisCfg.Addr, redisCfg.Password, redisCfg.UseTLS)

	pgStore, err := store.NewPostgresStore(ctx, config.PostgresDSN())
	if err != nil {
		log.Fatalf("failed to connect to postgres: %v", err)
	}

	sources, err := pgStore.ListSources(ctx)
	if err != nil {
		log.Fatalf("Failed to load sources: %v", err)
	}

	s3Cfg := config.LoadS3Config()
	s3Store, err := storage.NewS3Storage(ctx, s3Cfg.BucketName, s3Cfg.Region)
	if err != nil {
		log.Fatalf("failed to set up s3 storage: %v", err)
	}

	log.Printf("loaded %d allowlisted sources", len(sources))

	server := &api.Server{
		Store:   pgStore,
		Sources: sources,
		Queue:   jobQueue,
		S3:      s3Store,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /jobs", server.HandleCreateJob)
	mux.HandleFunc("GET /jobs/{id}", server.HandleGetJob)
	mux.HandleFunc("GET /jobs/{id}/artifacts", server.HandleGetJobArtifacts)
	mux.HandleFunc("GET /artifacts/{artifactId}", server.HandleGetArtifactFile)

	log.Println("API server listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", withCORS(mux)))
}

// withCORS wraps a handler, adding headers that allow the vite dev server (a different origin) to call this API from the browser.
func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
