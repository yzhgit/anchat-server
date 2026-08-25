package handler

import (
	"context"

	filev1 "flamingo/api/file/v1"

	empty "github.com/golang/protobuf/ptypes/empty"
)

// fileProxy implements FileServiceHTTPServer by forwarding to the gRPC client.
type fileProxy struct {
	client filev1.FileServiceClient
}

func (p *fileProxy) GenerateUploadToken(ctx context.Context, req *filev1.GenerateUploadTokenRequest) (*filev1.GenerateUploadTokenResponse, error) {
	return p.client.GenerateUploadToken(ctx, req)
}

func (p *fileProxy) CompleteUpload(ctx context.Context, req *filev1.CompleteUploadRequest) (*filev1.FileInfo, error) {
	return p.client.CompleteUpload(ctx, req)
}

func (p *fileProxy) GenerateDownloadURL(ctx context.Context, req *filev1.GenerateDownloadURLRequest) (*filev1.GenerateDownloadURLResponse, error) {
	return p.client.GenerateDownloadURL(ctx, req)
}

func (p *fileProxy) GetFileInfo(ctx context.Context, req *filev1.GetFileInfoRequest) (*filev1.FileInfo, error) {
	return p.client.GetFileInfo(ctx, req)
}

func (p *fileProxy) DeleteFile(ctx context.Context, req *filev1.DeleteFileRequest) (*empty.Empty, error) {
	return p.client.DeleteFile(ctx, req)
}

func (p *fileProxy) ListUserFiles(ctx context.Context, req *filev1.ListUserFilesRequest) (*filev1.ListUserFilesResponse, error) {
	return p.client.ListUserFiles(ctx, req)
}
