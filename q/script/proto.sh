#!/usr/bin/env bash
cd /workspace/q

echo "Generating proto files by types/proto/narwhal.proto"
protoc --go_out=/workspace/q --go-grpc_out=/workspace/q --proto_path /workspace/types/proto /workspace/types/proto/narwhal.proto

echo "Generating proto files by q/common/worker.proto"
protoc --go_out=/workspace/q --twirp_out=/workspace/q  common/worker.proto