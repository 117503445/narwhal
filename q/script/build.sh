#!/usr/bin/env bash

cd /workspace/q
CGO_ENABLED=0 go build -buildvcs=false -o /workspace/q/q .

cd /workspace/q/assets/fc-worker
CGO_ENABLED=0 go build -buildvcs=false -o fc-worker .