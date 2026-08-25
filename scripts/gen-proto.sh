#!/bin/bash

# gen-proto.sh
# Scan each subdirectory under api/ and generate Go + gRPC pb files

set -e

PROTO_ROOT="api"


# --- Check required tools ---
check_cmd() {
    if ! command -v "$1" &> /dev/null; then
        echo "Error: $1 not found."
        case "$1" in
            protoc)
                echo "Please install protoc: https://github.com/protocolbuffers/protobuf/releases"
                ;;
            protoc-gen-go)
                echo "Run: go install google.golang.org/protobuf/cmd/protoc-gen-go@latest"
                ;;
            protoc-gen-go-grpc)
                echo "Run: go install google.golang.org/protobuf/cmd/protoc-gen-go-grpc@latest"
                ;;
            protoc-gen-go-http)
                echo "Run: go install github.com/go-kratos/kratos/cmd/protoc-gen-go-http/v2@latest"
                ;;
        esac
        exit 1
    fi
}

check_cmd protoc
check_cmd protoc-gen-go
check_cmd protoc-gen-go-grpc
check_cmd protoc-gen-go-http

# Collect all .proto files (relative paths, for protoc arguments)
proto_files=$(find $PROTO_ROOT -name "*.proto" -printf "%P\n")

# Invoke protoc once, automatically handling dependencies
protoc \
    --proto_path=$PROTO_ROOT \
    --proto_path=./third_party \
    --go_out=$PROTO_ROOT \
    --go_opt=paths=source_relative \
    --go-grpc_out=$PROTO_ROOT \
    --go-grpc_opt=paths=source_relative \
    --go-http_out=$PROTO_ROOT \
    --go-http_opt=paths=source_relative \
    --experimental_allow_proto3_optional \
    $proto_files

echo "✓ Proto generation completed!"
