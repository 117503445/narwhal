#!/usr/bin/env bash
cd /workspace/q
protoc --go_out=/workspace/q --go-grpc_out=/workspace/q --proto_path /workspace/types/proto /workspace/types/proto/narwhal.proto