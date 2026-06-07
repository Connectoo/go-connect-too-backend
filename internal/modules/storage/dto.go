package storage

import "github.com/google/uuid"

// PresignRequest requests a presigned upload URL.
type PresignRequest struct {
	ContentType string `json:"content_type"`
	Purpose     string `json:"purpose"`
}

// PresignResponse returns upload instructions.
type PresignResponse struct {
	FileID    uuid.UUID `json:"file_id"`
	UploadURL string    `json:"upload_url"`
	ObjectKey string    `json:"object_key"`
	ExpiresIn int       `json:"expires_in_seconds"`
}

// FileResponse is a stored file metadata payload.
type FileResponse struct {
	ID          uuid.UUID `json:"id"`
	ObjectKey   string    `json:"object_key"`
	ContentType string    `json:"content_type"`
	Purpose     string    `json:"purpose"`
	CreatedAt   string    `json:"created_at"`
}
