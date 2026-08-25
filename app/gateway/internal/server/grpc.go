package server

import (
	"github.com/go-kratos/kratos/v2/transport/grpc"
)

// NewGRPCServer is a no-op provider for gateway.  The gateway only exposes an
// HTTP surface; passing a nil gRPC server to newApp keeps the gateway on the
// same uniform newApp signature as gRPC-backed services.
func NewGRPCServer() *grpc.Server {
	return nil
}
