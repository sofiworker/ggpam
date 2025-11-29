#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

need_cmd() {
	if ! command -v "$1" >/dev/null 2>&1; then
		echo "缺少依赖: $1" >&2
		exit 1
	fi
}

check_go_version() {
	local min="1.21"
	local ver
	ver="$(go env GOVERSION 2>/dev/null | sed 's/^go//')"
	if [[ -z "$ver" ]]; then
		echo "无法检测 Go 版本" >&2
		exit 1
	fi
	if [[ "$(printf '%s\n%s\n' "$min" "$ver" | sort -V | tail -n1)" != "$ver" ]]; then
		echo "Go 版本过低 (当前: $ver, 需要 >= $min)" >&2
		exit 1
	fi
}

need_cmd go
need_cmd gcc
need_cmd pkg-config
need_cmd tar
need_cmd dpkg-deb
need_cmd rpmbuild

check_go_version

echo "依赖检查通过: $(go env GOROOT)"
echo "Go 版本: $(go env GOVERSION)"
echo "工作目录: $ROOT_DIR"
