#!/usr/bin/env bash

set -e

docker compose up -d && docker compose exec -T q-dev /workspace/q/script/proto.sh # 生成 proto 文件