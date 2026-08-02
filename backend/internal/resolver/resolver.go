package resolver

import (
	"net/url"
	"strings"

	"github.com/JerChol/licensed-media-preview-platform/internal/models"
)

// ErrorCode Identifies why a resolution attempt failed, so the API layer can map it to a clear message for the UI.

type ErrorCode string

const (
	ErrUnknownSource      ErrorCode = "unknown_source"
	ErrAmbiguousSource    ErrorCode = "ambiguous_source"
	ErrDisallowedFileType ErrorCode = "disallowed_file_type"
	ErrBadInput           ErrorCode = "bad_input"
)

// Resolver Error is returned when a submitten input can't be safely resolved to exactly one allowlisted source
type ResolveError struct {
	Code    ErrorCode
	Message string
}

// Error implements Go's built-in 'error' interface, so a *ResolvedError can be used anywhere an 'error'  is expected.
func (e *ResolveError) Error() string {
	return e.Message
}

// ResolveResult is what we return when resolution succeeds.
type ResolveResult struct {
	SourceID        string
	ResolvedFileURL string
	AttributionText string
	LicenseType     string
}

// Resolve checks a submitted URL against every allowlisted source and succeeds only if exactly one source matches.
func Resolve(inputUrl string, sources []models.Source) (ResolveResult, *ResolveError) {
	parsed, err := url.Parse(inputUrl)
	if err != nil || parsed.Host == "" {
		return ResolveResult{}, &ResolveError{
			Code:    ErrBadInput,
			Message: "could not parse URL",
		}
	}

	host := strings.ToLower(parsed.Host)
	path := parsed.Path
	ext := strings.ToLower(getExtension(path))

	var matches []models.Source

	for _, source := range sources {
		if !hostMatches(source.MatchHosts, host) {
			continue
		}
		if len(source.PathPrefixes) > 0 && !hasAnyPrefix(path, source.PathPrefixes) {
			continue
		}
		if !containsIgnoreCase(source.FileExtensions, ext) {
			continue
		}
		matches = append(matches, source)
	}

	if len(matches) == 0 {
		return ResolveResult{}, &ResolveError{
			Code:    ErrUnknownSource,
			Message: "no allowlisted source matches this URL",
		}
	}
	if len(matches) > 1 {
		return ResolveResult{}, &ResolveError{
			Code:    ErrAmbiguousSource,
			Message: "more than one source matches; URL is ambiguous",
		}
	}

	chosen := matches[0]
	return ResolveResult{
		SourceID:        chosen.SourceID,
		ResolvedFileURL: inputUrl,
		AttributionText: chosen.AttributionText,
		LicenseType:     chosen.LicenseType,
	}, nil

}

// hostMatches checks whether host exactly matches any entry in allowed.
func hostMatches(allowed []string, host string) bool {
	for _, h := range allowed {
		if strings.ToLower(h) == host {
			return true
		}
	}
	return false
}

// hasAnyPrefix checks whether path starts with any of the given prefixes.
func hasAnyPrefix(path string, prefixes []string) bool {
	for _, p := range prefixes {
		if strings.HasPrefix(path, p) {
			return true
		}
	}
	return false
}

// containsIgnoreCase checks whether ext is present in list, case-insensitively.
func containsIgnoreCase(list []string, ext string) bool {
	for _, item := range list {
		if strings.ToLower(item) == ext {
			return true
		}
	}
	return false
}

// getExtension returns the file extension (including the dot) from a path, or "" if there isn't one.
func getExtension(path string) string {
	idx := strings.LastIndex(path, ".")
	if idx == -1 {
		return ""
	}
	return path[idx:]
}
