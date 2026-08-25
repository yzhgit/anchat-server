package server

import (
	"github.com/go-kratos/kratos/v2/transport/grpc"
)

// NewGRPCServer is a no-op provider for the realtime service.  The realtime
// service only exposes an HTTP/WebSocket surface; passing a nil gRPC server
// to newApp keeps the realtime service on the same uniform newApp signature
// as gRPC-backed services.
func NewGRPCServer() *grpc.Server {
	return nil
}
