-- File metadata table
CREATE TABLE IF NOT EXISTS files (
    id BIGSERIAL PRIMARY KEY,
    file_id VARCHAR(64) NOT NULL UNIQUE,
    user_id VARCHAR(36) NOT NULL,
    file_name VARCHAR(255) NOT NULL,
    file_type SMALLINT NOT NULL,  -- 1-image/2-video/3-audio/4-file/5-log
    file_size BIGINT NOT NULL,
    mime_type VARCHAR(100) NOT NULL,
    storage_path VARCHAR(500) NOT NULL,  -- MinIO object key
    thumbnail_path VARCHAR(500),
    bucket_name VARCHAR(50) NOT NULL,
    status SMALLINT DEFAULT 1,  -- 0-deleted, 1-active, 2-processing
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP,
    metadata JSONB
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_files_file_id ON files(file_id) WHERE status = 1;
CREATE INDEX IF NOT EXISTS idx_files_user_id ON files(user_id) WHERE status = 1;
CREATE INDEX IF NOT EXISTS idx_files_created_at ON files(created_at);
CREATE INDEX IF NOT EXISTS idx_files_expires_at ON files(expires_at) WHERE expires_at IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_files_bucket ON files(bucket_name);
CREATE INDEX IF NOT EXISTS idx_files_status ON files(status);

-- File upload tracking table (for chunked uploads, Phase 2)
CREATE TABLE IF NOT EXISTS file_uploads (
    id BIGSERIAL PRIMARY KEY,
    upload_id VARCHAR(64) NOT NULL UNIQUE,
    user_id VARCHAR(36) NOT NULL,
    file_name VARCHAR(255) NOT NULL,
    file_size BIGINT NOT NULL,
    chunk_size BIGINT NOT NULL,
    uploaded_chunks JSONB,
    status SMALLINT DEFAULT 1,  -- 1-pending/2-uploading/3-completed/4-failed
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP,
    expires_at TIMESTAMP NOT NULL
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_file_uploads_upload_id ON file_uploads(upload_id);
CREATE INDEX IF NOT EXISTS idx_file_uploads_user_id ON file_uploads(user_id);
CREATE INDEX IF NOT EXISTS idx_file_uploads_status ON file_uploads(status);
CREATE INDEX IF NOT EXISTS idx_file_uploads_expires_at ON file_uploads(expires_at);

-- Add comments
COMMENT ON TABLE files IS 'File metadata table';
COMMENT ON COLUMN files.file_id IS 'Unique file identifier';
COMMENT ON COLUMN files.user_id IS 'Uploader user ID';
COMMENT ON COLUMN files.file_type IS 'File type: 1-image/2-video/3-audio/4-file/5-log';
COMMENT ON COLUMN files.status IS 'Status: 0-deleted, 1-active, 2-processing';
COMMENT ON COLUMN files.storage_path IS 'MinIO storage path';
COMMENT ON COLUMN files.thumbnail_path IS 'Thumbnail path';
COMMENT ON COLUMN files.metadata IS 'Extended metadata (width, height, duration, etc.)';

COMMENT ON TABLE file_uploads IS 'File chunked upload tracking table';
COMMENT ON COLUMN file_uploads.upload_id IS 'Upload session ID';
COMMENT ON COLUMN file_uploads.uploaded_chunks IS 'Uploaded chunk information';
COMMENT ON COLUMN file_uploads.status IS 'Upload status: 1-pending/2-uploading/3-completed/4-failed';
