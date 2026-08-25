package model

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

// FileType represents file category.
type FileType int16

// FileStatus represents file lifecycle status.
type FileStatus int16

// UploadStatus represents chunked upload lifecycle status.
type UploadStatus int16

// File file metadata model
type File struct {
	ID            int64        `gorm:"column:id;primaryKey;autoIncrement"`
	FileID        string       `gorm:"column:file_id;not null;uniqueIndex"`
	UserID        string       `gorm:"column:user_id;not null"`
	FileName      string       `gorm:"column:file_name;not null"`
	FileType      FileType     `gorm:"column:file_type;type:smallint;not null"`
	FileSize      int64        `gorm:"column:file_size;not null"`
	MimeType      string       `gorm:"column:mime_type;not null"`
	StoragePath   string       `gorm:"column:storage_path;not null"`
	ThumbnailPath string       `gorm:"column:thumbnail_path"`
	BucketName    string       `gorm:"column:bucket_name;not null"`
	Status        FileStatus   `gorm:"column:status;type:smallint;default:1"`
	CreatedAt     time.Time    `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP"`
	ExpiresAt     *time.Time   `gorm:"column:expires_at"`
	Metadata      FileMetadata `gorm:"column:metadata;type:jsonb"`
}

// FileMetadata file extended metadata
type FileMetadata struct {
	Width    int    `json:"width,omitempty"`
	Height   int    `json:"height,omitempty"`
	Duration int    `json:"duration,omitempty"`
	Format   string `json:"format,omitempty"`
}

// Scan implements sql.Scanner interface
func (m *FileMetadata) Scan(value interface{}) error {
	if value == nil {
		*m = FileMetadata{}
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}

	return json.Unmarshal(bytes, m)
}

// Value implements driver.Valuer interface
func (m FileMetadata) Value() (driver.Value, error) {
	if m.Width == 0 && m.Height == 0 && m.Duration == 0 && m.Format == "" {
		return nil, nil
	}
	return json.Marshal(m)
}

// TableName returns table name
func (File) TableName() string {
	return "files"
}

// FileUpload file upload tracking model (for chunked upload)
type FileUpload struct {
	ID             int64        `gorm:"column:id;primaryKey;autoIncrement"`
	UploadID       string       `gorm:"column:upload_id;not null;uniqueIndex"`
	UserID         string       `gorm:"column:user_id;not null"`
	FileName       string       `gorm:"column:file_name;not null"`
	FileSize       int64        `gorm:"column:file_size;not null"`
	ChunkSize      int64        `gorm:"column:chunk_size;not null"`
	UploadedChunks ChunkInfo    `gorm:"column:uploaded_chunks;type:jsonb"`
	Status         UploadStatus `gorm:"column:status;type:smallint;default:1"`
	CreatedAt      time.Time    `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt      time.Time    `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP"`
	CompletedAt    *time.Time
	ExpiresAt      time.Time `gorm:"column:expires_at;not null"`
}

// ChunkInfo chunk information
type ChunkInfo struct {
	Chunks []int `json:"chunks"`
}

// Scan implements sql.Scanner interface
func (c *ChunkInfo) Scan(value interface{}) error {
	if value == nil {
		*c = ChunkInfo{}
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}

	return json.Unmarshal(bytes, c)
}

// Value implements driver.Valuer interface
func (c ChunkInfo) Value() (driver.Value, error) {
	return json.Marshal(c)
}

// TableName returns table name
func (FileUpload) TableName() string {
	return "file_uploads"
}

// file status constants
const (
	FileStatusDeleted    FileStatus = 0
	FileStatusActive     FileStatus = 1
	FileStatusProcessing FileStatus = 2
)

// file type constants
const (
	FileTypeUnspecified FileType = 0
	FileTypeImage       FileType = 1
	FileTypeVideo       FileType = 2
	FileTypeAudio       FileType = 3
	FileTypeFile        FileType = 4
	FileTypeLog         FileType = 5
)

// bucket name constants
const (
	BucketAvatar      = "avatar"
	BucketGroupAvatar = "group-avatar"
	BucketChatFile    = "chat-file"
)

// file size limit constants (bytes)
const (
	MaxImageSize = 10 * 1024 * 1024  // 10MB
	MaxVideoSize = 100 * 1024 * 1024 // 100MB
	MaxAudioSize = 20 * 1024 * 1024  // 20MB
	MaxFileSize  = 50 * 1024 * 1024  // 50MB
	MaxLogSize   = 50 * 1024 * 1024  // 50MB
)

// upload status constants
const (
	UploadStatusPending   UploadStatus = 1
	UploadStatusUploading UploadStatus = 2
	UploadStatusCompleted UploadStatus = 3
	UploadStatusFailed    UploadStatus = 4
)
