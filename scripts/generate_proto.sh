#!/usr/bin/env bash
# Generate Go code from protobuf definitions.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

cd "$PROJECT_ROOT"

# Ensure protoc-gen-go and protoc-gen-go-grpc are installed
if ! command -v protoc-gen-go &>/dev/null; then
    echo "Installing protoc-gen-go..."
    go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
fi

if ! command -v protoc-gen-go-grpc &>/dev/null; then
    echo "Installing protoc-gen-go-grpc..."
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
fi

export PATH="$PATH:$(go env GOPATH)/bin"

mkdir -p proto/aether/v1

protoc \
    --proto_path=proto \
    --go_out=proto --go_opt=paths=source_relative \
    --go-grpc_out=proto --go-grpc_opt=paths=source_relative \
    proto/aether/v1/tunnel.proto

echo "Proto generation complete."