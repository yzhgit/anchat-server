package service

import (
	"context"
	"fmt"
	"path"
	"strings"
	"time"

	filev1 "flamingo/api/file/v1"

	"flamingo/app/file/internal/handler"
	"flamingo/app/file/internal/model"
	"flamingo/pkg/errors"
	"flamingo/pkg/oss"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"
)

// fileServiceImpl file service implementation
type fileServiceImpl struct {
	fileRepo    FileRepository
	minioClient *oss.MinioClient
	db          *gorm.DB
	log         *log.Helper
}

var _ handler.FileService = (*fileServiceImpl)(nil)

// NewFileService creates file service
func NewFileService(fileRepo FileRepository, minioClient *oss.MinioClient, db *gorm.DB, logger log.Logger) handler.FileService {
	return &fileServiceImpl{
		fileRepo:    fileRepo,
		minioClient: minioClient,
		db:          db,
		log:         log.NewHelper(logger),
	}
}

// GenerateUploadToken generates upload token
func (s *fileServiceImpl) GenerateUploadToken(ctx context.Context, userID string, req *filev1.GenerateUploadTokenRequest) (*filev1.GenerateUploadTokenResponse, error) {
	// validate file size
	reqFileType := model.FileType(req.FileType)
	if err := s.validateFileSize(reqFileType, req.FileSize); err != nil {
		return nil, err
	}

	// generate unique file_id
	fileID := fmt.Sprintf("file-%s", uuid.New().String())

	// determine bucket
	bucketName := s.getBucketName(reqFileType)

	// generate storage path: {bucket}/{user_id}/{date}/{uuid}.{ext}
	now := time.Now()
	dateStr := now.Format("2006-01-02")
	ext := s.getFileExtension(req.FileName)
	storagePath := fmt.Sprintf("%s/%s/%s.%s", userID, dateStr, uuid.New().String(), ext)

	// calculate expiration time
	var expiresAt *time.Time
	if req.ExpiresHours > 0 {
		expires := now.Add(time.Duration(req.ExpiresHours) * time.Hour)
		expiresAt = &expires
	}

	// generate presigned upload URL (1 hour validity)
	uploadURL, err := s.minioClient.PresignedPutObject(ctx, bucketName, storagePath, time.Hour)
	if err != nil {
		return nil, errors.NewBusiness(errors.CodeFileUploadFailed, "failed to generate upload URL")
	}

	// create file record (status is processing)
	file := &model.File{
		FileID:      fileID,
		UserID:      userID,
		FileName:    req.FileName,
		FileType:    reqFileType,
		FileSize:    req.FileSize,
		MimeType:    req.MimeType,
		StoragePath: storagePath,
		BucketName:  bucketName,
		Status:      model.FileStatusProcessing,
		CreatedAt:   now,
		ExpiresAt:   expiresAt,
	}

	if err := s.fileRepo.Create(ctx, file); err != nil {
		return nil, errors.NewBusiness(errors.CodeInternalError, "failed to create file record")
	}

	return &filev1.GenerateUploadTokenResponse{
		FileId:    fileID,
		UploadUrl: uploadURL.String(),
		ExpiresAt: now.Add(time.Hour).Unix(),
	}, nil
}

// CompleteUpload completes upload
func (s *fileServiceImpl) CompleteUpload(ctx context.Context, fileID, userID string) (*filev1.FileInfo, error) {
	// validate file record
	file, err := s.fileRepo.GetByFileIDAndUserID(ctx, fileID, userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewBusiness(errors.CodeFileNotFound, "file not found")
		}
		return nil, errors.NewBusiness(errors.CodeInternalError, "failed to get file")
	}

	// validate file status
	if file.Status != model.FileStatusProcessing {
		return nil, errors.NewBusiness(errors.CodeInvalidFileID, "file already completed or deleted")
	}

	// validate file exists in MinIO
	_, err = s.minioClient.StatObject(ctx, file.BucketName, file.StoragePath)
	if err != nil {
		return nil, errors.NewBusiness(errors.CodeFileUploadFailed, "file not found in storage")
	}

	// update status to active
	file.Status = model.FileStatusActive
	if err := s.fileRepo.Update(ctx, file); err != nil {
		return nil, errors.NewBusiness(errors.CodeInternalError, "failed to update file status")
	}

	return s.toFileInfoResponse(file), nil
}

// GenerateDownloadURL generates download URL
func (s *fileServiceImpl) GenerateDownloadURL(ctx context.Context, fileID, userID string, expiresMinutes *int32) (*filev1.GenerateDownloadURLResponse, error) {
	// validate permission
	file, err := s.fileRepo.GetByFileIDAndUserID(ctx, fileID, userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewBusiness(errors.CodeFileNotFound, "file not found")
		}
		return nil, errors.NewBusiness(errors.CodeInternalError, "failed to get file")
	}

	// validate file status
	if file.Status != model.FileStatusActive {
		return nil, errors.NewBusiness(errors.CodeFileNotFound, "file is not active")
	}

	// check if file is expired
	if file.ExpiresAt != nil && file.ExpiresAt.Before(time.Now()) {
		return nil, errors.NewBusiness(errors.CodeFileExpired, "file has expired")
	}

	// generate download URL
	expires := 60 * time.Minute
	if expiresMinutes != nil && *expiresMinutes > 0 {
		expires = time.Duration(*expiresMinutes) * time.Minute
	}

	downloadURL, err := s.minioClient.PresignedGetObject(ctx, file.BucketName, file.StoragePath, expires)
	if err != nil {
		return nil, errors.NewBusiness(errors.CodeInternalError, "failed to generate download URL")
	}

	resp := &filev1.GenerateDownloadURLResponse{
		DownloadUrl: downloadURL.String(),
		ExpiresAt:   time.Now().Add(expires).Unix(),
	}

	return resp, nil
}

// GetFileInfo gets file info
func (s *fileServiceImpl) GetFileInfo(ctx context.Context, fileID, userID string) (*filev1.FileInfo, error) {
	file, err := s.fileRepo.GetByFileIDAndUserID(ctx, fileID, userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewBusiness(errors.CodeFileNotFound, "file not found")
		}
		return nil, errors.NewBusiness(errors.CodeInternalError, "failed to get file")
	}

	return s.toFileInfoResponse(file), nil
}

// DeleteFile deletes file
func (s *fileServiceImpl) DeleteFile(ctx context.Context, fileID, userID string) error {
	l := s.log.WithContext(ctx)
	file, err := s.fileRepo.GetByFileIDAndUserID(ctx, fileID, userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.NewBusiness(errors.CodeFileNotFound, "file not found")
		}
		return errors.NewBusiness(errors.CodeInternalError, "failed to get file")
	}

	// Delete from MinIO first so that, if DB deletion fails, we keep
	// the record (and can retry) instead of leaving orphaned objects.
	if err := s.minioClient.RemoveObject(ctx, file.BucketName, file.StoragePath); err != nil {
		l.Warnw("msg", "failed to delete file from MinIO", "error", err, "fileID", fileID)
	}

	if file.ThumbnailPath != "" {
		_ = s.minioClient.RemoveObject(ctx, file.BucketName, file.ThumbnailPath)
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		fileRepo := s.fileRepo.WithTx(tx)
		if err := fileRepo.Delete(ctx, fileID); err != nil {
			return errors.NewBusiness(errors.CodeInternalError, "failed to delete file record")
		}
		return nil
	})
}

// ListUserFiles lists user files
func (s *fileServiceImpl) ListUserFiles(ctx context.Context, userID string, fileType *model.FileType, page, pageSize int) (*filev1.ListUserFilesResponse, error) {
	files, total, err := s.fileRepo.ListByUserID(ctx, userID, fileType, page, pageSize)
	if err != nil {
		return nil, errors.NewBusiness(errors.CodeInternalError, "failed to list files")
	}

	fileInfos := make([]*filev1.FileInfo, 0, len(files))
	for _, file := range files {
		fileInfos = append(fileInfos, s.toFileInfoResponse(file))
	}

	return &filev1.ListUserFilesResponse{
		Total: total,
		Files: fileInfos,
	}, nil
}

// BatchGetFileInfo batch gets file info
func (s *fileServiceImpl) BatchGetFileInfo(ctx context.Context, fileIDs []string, userID string) ([]*filev1.FileInfo, error) {
	if len(fileIDs) == 0 {
		return []*filev1.FileInfo{}, nil
	}

	files, err := s.fileRepo.BatchGetByFileIDs(ctx, fileIDs)
	if err != nil {
		return nil, errors.NewBusiness(errors.CodeInternalError, "failed to batch get files")
	}

	fileInfos := make([]*filev1.FileInfo, 0, len(files))
	for _, file := range files {
		if file.UserID == userID {
			fileInfos = append(fileInfos, s.toFileInfoResponse(file))
		}
	}

	return fileInfos, nil
}

// validateFileSize validates file size
func (s *fileServiceImpl) validateFileSize(fileType model.FileType, fileSize int64) error {
	var maxSize int64

	switch fileType {
	case model.FileTypeImage:
		maxSize = model.MaxImageSize
	case model.FileTypeVideo:
		maxSize = model.MaxVideoSize
	case model.FileTypeAudio:
		maxSize = model.MaxAudioSize
	case model.FileTypeFile:
		maxSize = model.MaxFileSize
	case model.FileTypeLog:
		maxSize = model.MaxLogSize
	default:
		return errors.NewBusiness(errors.CodeFileTypeNotAllowed, "invalid file type")
	}

	if fileSize > maxSize {
		return errors.NewBusiness(errors.CodeFileSizeExceeded, fmt.Sprintf("file size exceeds limit: %d bytes", maxSize))
	}

	return nil
}

func (s *fileServiceImpl) getBucketName(_ model.FileType) string {
	return model.BucketChatFile
}

func (s *fileServiceImpl) getFileExtension(fileName string) string {
	ext := path.Ext(fileName)
	if ext != "" {
		ext = strings.TrimPrefix(ext, ".")
	}
	if ext == "" {
		ext = "bin"
	}
	return ext
}

// toFileInfoResponse converts to DTO
func (s *fileServiceImpl) toFileInfoResponse(file *model.File) *filev1.FileInfo {
	resp := &filev1.FileInfo{
		FileId:        file.FileID,
		FileName:      file.FileName,
		FileType:      filev1.FileType(file.FileType),
		FileSize:      file.FileSize,
		MimeType:      file.MimeType,
		StoragePath:   file.StoragePath,
		ThumbnailPath: file.ThumbnailPath,
		BucketName:    file.BucketName,
		Status:        int32(file.Status),
		CreatedAt:     timestamppb.New(file.CreatedAt),
	}

	if file.ExpiresAt != nil {
		resp.ExpiresAt = timestamppb.New(*file.ExpiresAt)
	}

	return resp
}
