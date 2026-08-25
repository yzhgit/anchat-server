package service

import (
	"context"

	"flamingo/app/file/internal/model"

	"gorm.io/gorm"
)

// FileRepository file repository interface
type FileRepository interface {
	// Create creates file record
	Create(ctx context.Context, file *model.File) error

	// GetByFileID gets file by file_id
	GetByFileID(ctx context.Context, fileID string) (*model.File, error)

	// GetByFileIDAndUserID gets file by file_id and user_id (permission validation)
	GetByFileIDAndUserID(ctx context.Context, fileID, userID string) (*model.File, error)

	// BatchGetByFileIDs batch get files
	BatchGetByFileIDs(ctx context.Context, fileIDs []string) ([]*model.File, error)

	// ListByUserID lists user files (supports pagination and type filter)
	ListByUserID(ctx context.Context, userID string, fileType *model.FileType, page, pageSize int) ([]*model.File, int64, error)

	// Update updates file info
	Update(ctx context.Context, file *model.File) error

	// UpdateStatus updates file status
	UpdateStatus(ctx context.Context, fileID string, status model.FileStatus) error

	// Delete soft deletes file
	Delete(ctx context.Context, fileID string) error

	// DeleteExpired cleans up expired files
	DeleteExpired(ctx context.Context) error

	// WithTx uses transaction
	WithTx(tx *gorm.DB) FileRepository
}
