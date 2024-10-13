#!/usr/bin/env bash

cd /workspace/q
CGO_ENABLED=0 go build -o /workspace/q/q .