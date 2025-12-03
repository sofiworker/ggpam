#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

#"${ROOT_DIR}/scripts/check_deps.sh"

echo "正在执行 make build ..."
cd "$ROOT_DIR"
make build
echo "构建完成: $ROOT_DIR/bin"
