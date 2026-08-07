package pdfpipeline

import (
	"fmt"
	"strings"

	"github.com/ledongthuc/pdf"
)

// ExtractResult holds what we pull out of a downloaded PDF
type ExtractResult struct {
	TextSnippet string
	PageCount   int
}

// Extract opens a local PDF file and pulls a short text snippet plus basic metadata
func Extract(localPath string) (ExtractResult, error) {
	f, r, err := pdf.Open(localPath)
	if err != nil {
		return ExtractResult{}, fmt.Errorf("failed to open pdf: %w", err)
	}
	defer f.Close()

	var builder strings.Builder
	totalPages := r.NumPage()

	for i := 1; i <= totalPages; i++ {
		page := r.Page(i)
		if page.V.IsNull() {
			continue
		}
		text, err := page.GetPlainText(nil)
		if err != nil {
			continue
		}
		builder.WriteString(text)

		//Stop early once we have enough for a snippet.
		if builder.Len() > 500 {
			break
		}
	}

	snippet := builder.String()
	if len(snippet) > 500 {
		snippet = snippet[:500]
	}

	return ExtractResult{
		TextSnippet: snippet,
		PageCount:   totalPages,
	}, nil

}
