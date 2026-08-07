package pdfpipeline

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

// DownloadToTemp fetches the file at url and writes it to a temp file on disk, returning the local path.
func DownloadToTemp(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("http get failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status %d fetching %s", resp.StatusCode, url)
	}

	tmpFile, err := os.CreateTemp("", "lmpp-*.pdf")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file. %w", err)
	}
	defer tmpFile.Close()

	if _, err := io.Copy(tmpFile, resp.Body); err != nil {
		return "", fmt.Errorf("failed to write temp file: %w", err)
	}

	return tmpFile.Name(), nil
}
