package handler

import (
	"context"

	filev1 "flamingo/api/file/v1"

	"flamingo/app/file/internal/model"
	"flamingo/pkg/errors"
	"flamingo/pkg/md"

	"github.com/go-kratos/kratos/v2/log"
	empty "github.com/golang/protobuf/ptypes/empty"
)

// FileService file service interface
type FileService interface {
	// GenerateUploadToken generates upload token
	GenerateUploadToken(ctx context.Context, userID string, req *filev1.GenerateUploadTokenRequest) (*filev1.GenerateUploadTokenResponse, error)
	// CompleteUpload completes upload
	CompleteUpload(ctx context.Context, fileID, userID string) (*filev1.FileInfo, error)
	// GenerateDownloadURL generates download URL
	GenerateDownloadURL(ctx context.Context, fileID, userID string, expiresMinutes *int32) (*filev1.GenerateDownloadURLResponse, error)
	// GetFileInfo gets file info
	GetFileInfo(ctx context.Context, fileID, userID string) (*filev1.FileInfo, error)
	// DeleteFile deletes file
	DeleteFile(ctx context.Context, fileID, userID string) error
	// ListUserFiles lists user files
	ListUserFiles(ctx context.Context, userID string, fileType *model.FileType, page, pageSize int) (*filev1.ListUserFilesResponse, error)
	// BatchGetFileInfo batch gets file info
	BatchGetFileInfo(ctx context.Context, fileIDs []string, userID string) ([]*filev1.FileInfo, error)
}

type FileHandler struct {
	filev1.UnimplementedFileServiceServer
	svc FileService
	log *log.Helper
}

// NewFileServer creates gRPC handler
func NewFileHandler(svc FileService, logger log.Logger) *FileHandler {
	return &FileHandler{
		svc: svc,
		log: log.NewHelper(logger),
	}
}

// GenerateUploadToken generates upload token
func (s *FileHandler) GenerateUploadToken(ctx context.Context, req *filev1.GenerateUploadTokenRequest) (*filev1.GenerateUploadTokenResponse, error) {
	userID := md.MustGetUserID(ctx)
	resp, err := s.svc.GenerateUploadToken(ctx, userID, req)
	if err != nil {
		return nil, errors.ConvertError(ctx, err)
	}

	return resp, nil
}

// CompleteUpload completes upload
func (s *FileHandler) CompleteUpload(ctx context.Context, req *filev1.CompleteUploadRequest) (*filev1.FileInfo, error) {
	userID := md.MustGetUserID(ctx)
	resp, err := s.svc.CompleteUpload(ctx, req.FileId, userID)
	if err != nil {
		return nil, errors.ConvertError(ctx, err)
	}

	return resp, nil
}

// GenerateDownloadURL generates download URL
func (s *FileHandler) GenerateDownloadURL(ctx context.Context, req *filev1.GenerateDownloadURLRequest) (*filev1.GenerateDownloadURLResponse, error) {
	userID := md.MustGetUserID(ctx)
	var expiresMinutes *int32
	if req.ExpiresMinutes > 0 {
		expiresMinutes = &req.ExpiresMinutes
	}
	resp, err := s.svc.GenerateDownloadURL(ctx, req.FileId, userID, expiresMinutes)
	if err != nil {
		return nil, errors.ConvertError(ctx, err)
	}

	return resp, nil
}

// GetFileInfo gets file info
func (s *FileHandler) GetFileInfo(ctx context.Context, req *filev1.GetFileInfoRequest) (*filev1.FileInfo, error) {
	userID := md.MustGetUserID(ctx)
	resp, err := s.svc.GetFileInfo(ctx, req.FileId, userID)
	if err != nil {
		return nil, errors.ConvertError(ctx, err)
	}

	return resp, nil
}

// DeleteFile deletes file
func (s *FileHandler) DeleteFile(ctx context.Context, req *filev1.DeleteFileRequest) (*empty.Empty, error) {
	userID := md.MustGetUserID(ctx)
	err := s.svc.DeleteFile(ctx, req.FileId, userID)
	if err != nil {
		return nil, errors.ConvertError(ctx, err)
	}

	return &empty.Empty{}, nil
}

// ListUserFiles lists user files
func (s *FileHandler) ListUserFiles(ctx context.Context, req *filev1.ListUserFilesRequest) (*filev1.ListUserFilesResponse, error) {
	userID := md.MustGetUserID(ctx)
	ft := model.FileType(req.FileType)
	fileType := &ft
	resp, err := s.svc.ListUserFiles(ctx, userID, fileType, int(req.Page), int(req.PageSize))
	if err != nil {
		return nil, errors.ConvertError(ctx, err)
	}

	return resp, nil
}

// BatchGetFileInfo batch gets file info
func (s *FileHandler) BatchGetFileInfo(ctx context.Context, req *filev1.BatchGetFileInfoRequest) (*filev1.BatchGetFileInfoResponse, error) {
	files, err := s.svc.BatchGetFileInfo(ctx, req.FileIds, req.UserId)
	if err != nil {
		return nil, errors.ConvertError(ctx, err)
	}

	result := make(map[string]*filev1.FileInfo, len(files))
	for _, f := range files {
		result[f.FileId] = f
	}
	return &filev1.BatchGetFileInfoResponse{
		Files: result,
	}, nil
}
