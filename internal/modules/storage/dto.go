package storage

// FileReference is a logical pointer to a stored file.
//
// In this phase only the URL field is populated. Future implementations may
// add fields such as bucket, key, content type, or size without changing the
// public service surface.
type FileReference struct {
	URL string `json:"url"`
}
