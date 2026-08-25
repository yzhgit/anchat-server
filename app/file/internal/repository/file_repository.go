package repository

import (
	"context"
	"time"

	"flamingo/app/file/internal/model"
	"flamingo/app/file/internal/service"

	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"
)

// fileRepositoryImpl file repository implementation
type fileRepositoryImpl struct {
	db  *gorm.DB
	log *log.Helper
}

// NewFileRepository creates file repository
func NewFileRepository(db *gorm.DB, logger log.Logger) service.FileRepository {
	return &fileRepositoryImpl{
		db:  db,
		log: log.NewHelper(logger),
	}
}

// Create creates file record
func (r *fileRepositoryImpl) Create(ctx context.Context, file *model.File) error {
	return r.db.WithContext(ctx).Create(file).Error
}

// GetByFileID gets file by file_id
func (r *fileRepositoryImpl) GetByFileID(ctx context.Context, fileID string) (*model.File, error) {
	var file model.File
	err := r.db.WithContext(ctx).
		Where("file_id = ? AND status != ?", fileID, model.FileStatusDeleted).
		First(&file).Error
	if err != nil {
		return nil, err
	}
	return &file, nil
}

// GetByFileIDAndUserID gets file by file_id and user_id (permission validation)
func (r *fileRepositoryImpl) GetByFileIDAndUserID(ctx context.Context, fileID, userID string) (*model.File, error) {
	var file model.File
	err := r.db.WithContext(ctx).
		Where("file_id = ? AND user_id = ? AND status != ?", fileID, userID, model.FileStatusDeleted).
		First(&file).Error
	if err != nil {
		return nil, err
	}
	return &file, nil
}

// BatchGetByFileIDs batch get files
func (r *fileRepositoryImpl) BatchGetByFileIDs(ctx context.Context, fileIDs []string) ([]*model.File, error) {
	var files []*model.File
	err := r.db.WithContext(ctx).
		Where("file_id IN ? AND status != ?", fileIDs, model.FileStatusDeleted).
		Find(&files).Error
	if err != nil {
		return nil, err
	}
	return files, nil
}

// ListByUserID lists user files (supports pagination and type filter)
func (r *fileRepositoryImpl) ListByUserID(ctx context.Context, userID string, fileType *model.FileType, page, pageSize int) ([]*model.File, int64, error) {
	var files []*model.File
	var total int64

	query := r.db.WithContext(ctx).
		Where("user_id = ? AND status = ?", userID, model.FileStatusActive)

	// filter by file type
	if fileType != nil && *fileType != model.FileTypeUnspecified {
		query = query.Where("file_type = ?", *fileType)
	}

	// count total
	if err := query.Model(&model.File{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// paginated query
	offset := (page - 1) * pageSize
	err := query.Order("created_at DESC").
		Limit(pageSize).
		Offset(offset).
		Find(&files).Error

	if err != nil {
		return nil, 0, err
	}

	return files, total, nil
}

// Update updates file info
func (r *fileRepositoryImpl) Update(ctx context.Context, file *model.File) error {
	return r.db.WithContext(ctx).Save(file).Error
}

// UpdateStatus updates file status
func (r *fileRepositoryImpl) UpdateStatus(ctx context.Context, fileID string, status model.FileStatus) error {
	return r.db.WithContext(ctx).
		Model(&model.File{}).
		Where("file_id = ?", fileID).
		Update("status", status).Error
}

// Delete soft deletes file
func (r *fileRepositoryImpl) Delete(ctx context.Context, fileID string) error {
	return r.db.WithContext(ctx).
		Model(&model.File{}).
		Where("file_id = ?", fileID).
		Update("status", model.FileStatusDeleted).Error
}

// DeleteExpired cleans up expired files
func (r *fileRepositoryImpl) DeleteExpired(ctx context.Context) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&model.File{}).
		Where("expires_at IS NOT NULL AND expires_at < ? AND status != ?", now, model.FileStatusDeleted).
		Update("status", model.FileStatusDeleted).Error
}

// WithTx uses transaction
func (r *fileRepositoryImpl) WithTx(tx *gorm.DB) service.FileRepository {
	return &fileRepositoryImpl{db: tx}
}
