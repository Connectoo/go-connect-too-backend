// Package storage is a placeholder layer for future profile photo and KYC
// document uploads.
//
// It currently validates and returns URL-based file references only. The
// Service type is intentionally small so a future implementation backed by
// S3, Firebase Storage, or local uploads can swap in without changing call
// sites in employee or KYC modules.
package storage

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
)

// ErrInvalidURL is returned when the given file URL is empty or malformed.
var ErrInvalidURL = errors.New("invalid file url")

// FileStorage describes the storage surface used by other modules.
//
// Real upload, signing, and deletion methods will be added when a concrete
// storage provider is wired in.
type FileStorage interface {
	ValidateURL(raw string) (FileReference, error)
}

// Service is the placeholder storage implementation backed by URL validation.
type Service struct{}

// NewService creates a placeholder storage service.
func NewService() *Service {
	return &Service{}
}

// ValidateURL trims and validates a file URL, returning a FileReference on
// success.
func (s *Service) ValidateURL(raw string) (FileReference, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return FileReference{}, fmt.Errorf("%w: url is required", ErrInvalidURL)
	}

	parsed, err := url.Parse(trimmed)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return FileReference{}, fmt.Errorf("%w: url must be absolute", ErrInvalidURL)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return FileReference{}, fmt.Errorf("%w: url must use http or https", ErrInvalidURL)
	}

	return FileReference{URL: trimmed}, nil
}
