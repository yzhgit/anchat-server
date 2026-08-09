package grpc

import (
	"context"

	commonpb "github.com/anychat/server/api/proto/common"
	versionpb "github.com/anychat/server/api/proto/version"
	"github.com/anychat/server/internal/version/service"
	pkgerrors "github.com/anychat/server/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type VersionServer struct {
	versionpb.UnimplementedVersionServiceServer
	svc service.VersionService
}

func NewVersionServer(serice service.VersionService) *VersionServer {
	return &VersionServer{
		svc: serice,
	}
}

func (s *VersionServer) CheckVersion(ctx context.Context, req *versionpb.CheckVersionRequest) (*versionpb.CheckVersionResponse, error) {
	resp, err := s.svc.CheckVersion(ctx, req)
	if err != nil {
		return nil, convertError(err)
	}
	return resp, nil
}

func (s *VersionServer) GetLatestVersion(ctx context.Context, req *versionpb.GetLatestVersionRequest) (*versionpb.GetLatestVersionResponse, error) {
	resp, err := s.svc.GetLatestVersion(ctx, req)
	if err != nil {
		return nil, convertError(err)
	}
	return resp, nil
}

func (s *VersionServer) ListVersions(ctx context.Context, req *versionpb.ListVersionsRequest) (*versionpb.ListVersionsResponse, error) {
	resp, err := s.svc.ListVersions(ctx, req)
	if err != nil {
		return nil, convertError(err)
	}
	return resp, nil
}

func (s *VersionServer) CreateVersion(ctx context.Context, req *versionpb.CreateVersionRequest) (*versionpb.CreateVersionResponse, error) {
	resp, err := s.svc.CreateVersion(ctx, req)
	if err != nil {
		return nil, convertError(err)
	}
	return resp, nil
}

func (s *VersionServer) GetVersion(ctx context.Context, req *versionpb.GetVersionRequest) (*versionpb.GetVersionResponse, error) {
	resp, err := s.svc.GetVersion(ctx, req)
	if err != nil {
		return nil, convertError(err)
	}
	return resp, nil
}

func (s *VersionServer) DeleteVersion(ctx context.Context, req *versionpb.DeleteVersionRequest) (*commonpb.Empty, error) {
	err := s.svc.DeleteVersion(ctx, req)
	if err != nil {
		return nil, convertError(err)
	}
	return &commonpb.Empty{}, nil
}

func (s *VersionServer) ReportVersion(ctx context.Context, req *versionpb.ReportVersionRequest) (*commonpb.Empty, error) {
	err := s.svc.ReportVersion(ctx, req)
	if err != nil {
		return nil, convertError(err)
	}
	return &commonpb.Empty{}, nil
}

func convertError(err error) error {
	if err == nil {
		return nil
	}

	if bizErr, ok := err.(*pkgerrors.Business); ok {
		switch bizErr.Code {
		case pkgerrors.CodeNotFound:
			return status.Error(codes.NotFound, bizErr.Message)
		case pkgerrors.CodeParamError:
			return status.Error(codes.InvalidArgument, bizErr.Message)
		case pkgerrors.CodeUnauthorized:
			return status.Error(codes.Unauthenticated, bizErr.Message)
		case pkgerrors.CodeForbidden:
			return status.Error(codes.PermissionDenied, bizErr.Message)
		default:
			return status.Error(codes.Internal, bizErr.Message)
		}
	}

	return status.Error(codes.Internal, err.Error())
}
