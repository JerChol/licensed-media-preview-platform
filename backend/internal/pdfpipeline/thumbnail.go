package pdfpipeline

import (
	"fmt"
	"os/exec"
)

// GenerateThumbnail renders the first page of a PDF to a JPEG file.
// outputPrefix is a path without extension; pdftoppm appends "-1.jpg".
func GenerateThumbnail(localPath string, outputPrefix string) (string, error) {
	cmd := exec.Command("pdftoppm", "-jpeg", "-f", "1", "-l", "1", "-r", "72", localPath, outputPrefix)
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("pdftoppm failed: %w", err)
	}
	return outputPrefix + "-1.jpg", nil
}
