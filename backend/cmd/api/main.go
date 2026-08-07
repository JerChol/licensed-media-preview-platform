package main

import (
	"context"
	"log"
	"net/http"

	"github.com/JerChol/licensed-media-preview-platform/internal/api"
	"github.com/JerChol/licensed-media-preview-platform/internal/config"
	"github.com/JerChol/licensed-media-preview-platform/internal/queue"
	"github.com/JerChol/licensed-media-preview-platform/internal/store"
)

func main() {
	ctx := context.Background()
	jobQueue := queue.NewRedisQueue("localhost:6379")

	pgStore, err := store.NewPostgresStore(ctx, config.PostgresDSN())
	if err != nil {
		log.Fatalf("failed to connect to postgres: %v", err)
	}

	sources, err := pgStore.ListSources(ctx)
	if err != nil {
		log.Fatalf("Failed to load sources: %v", err)
	}
	log.Printf("loaded %d allowlisted sources", len(sources))

	server := &api.Server{
		Store:   pgStore,
		Sources: sources,
		Queue:   jobQueue,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /jobs", server.HandleCreateJob)
	mux.HandleFunc("GET /jobs/{id}", server.HandleGetJob)
	mux.HandleFunc("GET /jobs/{id}/artifacts", server.HandleGetJobArtifacts)

	log.Println("API server listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
