package main

import (
	"log"
	"net/http"

	"github.com/JerChol/licensed-media-preview-platform/internal/api"
	"github.com/JerChol/licensed-media-preview-platform/internal/models"
	"github.com/JerChol/licensed-media-preview-platform/internal/store"
)

func main() {
	// Hardcoded test source for now - real allowlist will come from Postgres later.
	sources := []models.Source{
		{
			SourceID:         "example-raw-github",
			DisplayName:      "Example Github Raw PDFs",
			LicenseType:      "CCO",
			AttributionText:  "Source: example repository",
			MatchHosts:       []string{"raw.githubusercontent.com"},
			PathPrefixes:     []string{"/example-org/"},
			FileExtensions:   []string{".pdf"},
			AllowedMediaKind: "pdf",
		},
	}
	server := &api.Server{
		Store:   store.NewMemoryStore(),
		Sources: sources,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /jobs", server.HandleCreateJob)
	mux.HandleFunc("GET /jobs/{id}", server.HandleGetJob)

	log.Println("API server listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
