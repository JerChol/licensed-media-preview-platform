package storage

import (
	"fmt"
	"os"
	"path/filepath"
)

const baseDir = "data/artifacts"

// SaveTextSnippet writes a text snippet to disk under the given job ID and returns the path it was saved to.
func SaveTextSnippet(jobID string, content string) (string, error) {
	dir := filepath.Join(baseDir, jobID)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("failed to create artifact dir: %w", err)
	}

	path := filepath.Join(dir, "snippet.txt")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return "", fmt.Errorf("failed to write snippet: %w", err)
	}

	return path, nil
}
