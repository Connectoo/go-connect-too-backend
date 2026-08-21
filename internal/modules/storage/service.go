package storage

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// ObjectStore generates presigned upload URLs and public object references.
type ObjectStore interface {
	PresignPut(ctx context.Context, objectKey, contentType string, expires time.Duration) (string, error)
	PublicURL(objectKey string) string
}

// FileStore persists uploaded file metadata.
type FileStore interface {
	Create(ctx context.Context, file *UploadedFile) (*UploadedFile, error)
	GetByID(ctx context.Context, id uuid.UUID) (*UploadedFile, error)
	SoftDelete(ctx context.Context, userID, fileID uuid.UUID, at time.Time) error
}

// Service handles uploaded file business logic.
type Service struct {
	store   FileStore
	objects ObjectStore
	now     func() time.Time
}

// NewService creates a storage service.
func NewService(store FileStore, objects ObjectStore) *Service {
	return &Service{
		store:   store,
		objects: objects,
		now:     func() time.Time { return time.Now().UTC() },
	}
}

const presignTTL = 15 * time.Minute

var allowedPurposes = map[string]struct{}{
	PurposeKYCIDProof:      {},
	PurposeKYCAddressProof: {},
	PurposeProfilePhoto:    {},
}

// PresignUpload creates a presigned upload URL for an authenticated user.
func (s *Service) PresignUpload(ctx context.Context, userID uuid.UUID, req PresignRequest) (*PresignResponse, error) {
	contentType := strings.TrimSpace(req.ContentType)
	purpose := strings.TrimSpace(req.Purpose)
	if contentType == "" || purpose == "" {
		return nil, fmt.Errorf("%w: content_type and purpose are required", ErrValidation)
	}
	if _, ok := allowedPurposes[purpose]; !ok {
		return nil, fmt.Errorf("%w: unsupported purpose", ErrValidation)
	}
	if s.objects == nil {
		return nil, ErrUnavailable
	}

	fileID := uuid.New()
	objectKey := fmt.Sprintf("uploads/%s/%s/%s", userID.String(), purpose, fileID.String())
	at := s.now()

	uploadURL, err := s.objects.PresignPut(ctx, objectKey, contentType, presignTTL)
	if err != nil {
		return nil, fmt.Errorf("presign upload: %w", err)
	}

	if _, err := s.store.Create(ctx, &UploadedFile{
		ID:          fileID,
		UserID:      userID,
		ObjectKey:   objectKey,
		ContentType: contentType,
		Purpose:     purpose,
		CreatedAt:   at,
	}); err != nil {
		return nil, err
	}

	return &PresignResponse{
		FileID:    fileID,
		UploadURL: uploadURL,
		ObjectKey: objectKey,
		ExpiresIn: int(presignTTL.Seconds()),
	}, nil
}

// Delete removes an uploaded file owned by the authenticated user.
func (s *Service) Delete(ctx context.Context, userID, fileID uuid.UUID) error {
	return s.store.SoftDelete(ctx, userID, fileID, s.now())
}

// ResolveFileURL returns the public URL for an owned uploaded file.
func (s *Service) ResolveFileURL(ctx context.Context, userID, fileID uuid.UUID) (string, error) {
	file, err := s.store.GetByID(ctx, fileID)
	if err != nil {
		return "", err
	}
	if file.UserID != userID {
		return "", ErrForbidden
	}
	if s.objects == nil {
		return file.ObjectKey, nil
	}
	return s.objects.PublicURL(file.ObjectKey), nil
}
