package grpc

import (
	"context"
	"encoding/json"

	filepb "github.com/anychat/server/api/proto/file"
	"github.com/anychat/server/internal/file/dto"
	"github.com/anychat/server/internal/file/model"
	"github.com/anychat/server/internal/file/service"
	"github.com/anychat/server/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// FileServer gRPC server
type FileServer struct {
	filepb.UnimplementedFileServiceServer
	svc service.FileService
}

// NewFileServer creates gRPC server
func NewFileServer(service service.FileService) *FileServer {
	return &FileServer{
		svc: service,
	}
}

// GenerateUploadToken generates upload token
func (s *FileServer) GenerateUploadToken(ctx context.Context, req *filepb.GenerateUploadTokenRequest) (*filepb.GenerateUploadTokenResponse, error) {
	dtoReq := &dto.GenerateUploadTokenRequest{
		FileName:     req.FileName,
		FileSize:     req.FileSize,
		MimeType:     req.MimeType,
		FileType:     int32(req.FileType),
		ExpiresHours: req.ExpiresHours,
	}

	resp, err := s.svc.GenerateUploadToken(ctx, req.UserId, dtoReq)
	if err != nil {
		return nil, convertError(err)
	}

	return &filepb.GenerateUploadTokenResponse{
		FileId:    resp.FileID,
		UploadUrl: resp.UploadURL,
		ExpiresIn: resp.ExpiresIn,
	}, nil
}

// CompleteUpload completes upload
func (s *FileServer) CompleteUpload(ctx context.Context, req *filepb.CompleteUploadRequest) (*filepb.FileInfo, error) {
	resp, err := s.svc.CompleteUpload(ctx, req.FileId, req.UserId)
	if err != nil {
		return nil, convertError(err)
	}

	return toProtoFileInfo(resp), nil
}

// GenerateDownloadURL generates download URL
func (s *FileServer) GenerateDownloadURL(ctx context.Context, req *filepb.GenerateDownloadURLRequest) (*filepb.GenerateDownloadURLResponse, error) {
	resp, err := s.svc.GenerateDownloadURL(ctx, req.FileId, req.UserId, req.ExpiresMinutes)
	if err != nil {
		return nil, convertError(err)
	}

	pbResp := &filepb.GenerateDownloadURLResponse{
		DownloadUrl: resp.DownloadURL,
		ExpiresIn:   resp.ExpiresIn,
	}

	if resp.ThumbnailURL != "" {
		pbResp.ThumbnailUrl = &resp.ThumbnailURL
	}

	return pbResp, nil
}

// GetFileInfo gets file info
func (s *FileServer) GetFileInfo(ctx context.Context, req *filepb.GetFileInfoRequest) (*filepb.FileInfo, error) {
	resp, err := s.svc.GetFileInfo(ctx, req.FileId, req.UserId)
	if err != nil {
		return nil, convertError(err)
	}

	return toProtoFileInfo(resp), nil
}

// DeleteFile deletes file
func (s *FileServer) DeleteFile(ctx context.Context, req *filepb.DeleteFileRequest) (*filepb.DeleteFileResponse, error) {
	err := s.svc.DeleteFile(ctx, req.FileId, req.UserId)
	if err != nil {
		return nil, convertError(err)
	}

	return &filepb.DeleteFileResponse{
		Success: true,
	}, nil
}

// ListUserFiles lists user files
func (s *FileServer) ListUserFiles(ctx context.Context, req *filepb.ListUserFilesRequest) (*filepb.ListUserFilesResponse, error) {
	var fileType *model.FileType
	if req.FileType != nil {
		v := model.FileType(*req.FileType)
		fileType = &v
	}
	resp, err := s.svc.ListUserFiles(ctx, req.UserId, fileType, int(req.Page), int(req.PageSize))
	if err != nil {
		return nil, convertError(err)
	}

	files := make([]*filepb.FileInfo, 0, len(resp.Files))
	for _, file := range resp.Files {
		files = append(files, toProtoFileInfo(file))
	}

	return &filepb.ListUserFilesResponse{
		Files:    files,
		Total:    resp.Total,
		Page:     int32(resp.Page),
		PageSize: int32(resp.PageSize),
	}, nil
}

// BatchGetFileInfo batch gets file info
func (s *FileServer) BatchGetFileInfo(ctx context.Context, req *filepb.BatchGetFileInfoRequest) (*filepb.BatchGetFileInfoResponse, error) {
	resp, err := s.svc.BatchGetFileInfo(ctx, req.FileIds, req.UserId)
	if err != nil {
		return nil, convertError(err)
	}

	files := make([]*filepb.FileInfo, 0, len(resp))
	for _, file := range resp {
		files = append(files, toProtoFileInfo(file))
	}

	return &filepb.BatchGetFileInfoResponse{
		Files: files,
	}, nil
}

// toProtoFileInfo converts to proto FileInfo
func toProtoFileInfo(file *dto.FileInfoResponse) *filepb.FileInfo {
	pbFile := &filepb.FileInfo{
		FileId:        file.FileID,
		UserId:        file.UserID,
		FileName:      file.FileName,
		FileType:      filepb.FileType(file.FileType),
		FileSize:      file.FileSize,
		MimeType:      file.MimeType,
		StoragePath:   file.StoragePath,
		ThumbnailPath: file.ThumbnailPath,
		BucketName:    file.BucketName,
		Status:        filepb.FileStatus(file.Status),
		CreatedAt:     file.CreatedAt,
	}

	if file.ExpiresAt != nil {
		pbFile.ExpiresAt = file.ExpiresAt
	}

	if file.Metadata != nil {
		metadataJSON, _ := json.Marshal(file.Metadata)
		metadata := string(metadataJSON)
		pbFile.Metadata = &metadata
	}

	if file.DownloadURL != "" {
		pbFile.DownloadUrl = &file.DownloadURL
	}

	if file.ThumbnailURL != "" {
		pbFile.ThumbnailUrl = &file.ThumbnailURL
	}

	return pbFile
}

// convertError converts error
func convertError(err error) error {
	if bizErr, ok := err.(*errors.Business); ok {
		switch bizErr.Code {
		case errors.CodeFileNotFound:
			return status.Error(codes.NotFound, bizErr.Message)
		case errors.CodeFileAccessDenied:
			return status.Error(codes.PermissionDenied, bizErr.Message)
		case errors.CodeFileSizeExceeded, errors.CodeFileTypeNotAllowed, errors.CodeInvalidFileID:
			return status.Error(codes.InvalidArgument, bizErr.Message)
		case errors.CodeFileUploadFailed:
			return status.Error(codes.Internal, bizErr.Message)
		case errors.CodeFileExpired:
			return status.Error(codes.FailedPrecondition, bizErr.Message)
		default:
			return status.Error(codes.Internal, bizErr.Message)
		}
	}
	return status.Error(codes.Internal, err.Error())
}
