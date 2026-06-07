CREATE TABLE uploaded_files (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    object_key TEXT NOT NULL,
    content_type VARCHAR(100) NOT NULL,
    size_bytes BIGINT,
    purpose VARCHAR(50) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_uploaded_files_user_id ON uploaded_files (user_id);
CREATE INDEX idx_uploaded_files_purpose ON uploaded_files (purpose);
CREATE UNIQUE INDEX idx_uploaded_files_object_key ON uploaded_files (object_key);
