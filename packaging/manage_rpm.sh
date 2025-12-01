#!/usr/bin/env bash
set -euo pipefail

usage() {
	cat <<'USAGE'
用法: manage_rpm.sh <install|upgrade|reinstall|remove> <package.rpm>
环境变量:
  PACKAGE_NAME  目标包名（默认 gpam）
示例:
  sudo ./packaging/manage_rpm.sh install dist/gpam-0.1.0-1.x86_64.rpm
  sudo ./packaging/manage_rpm.sh reinstall dist/gpam-0.1.0-1.x86_64.rpm
  sudo ./packaging/manage_rpm.sh remove
USAGE
}

require_root() {
	if [[ "${EUID:-$(id -u)}" -ne 0 ]]; then
		echo "需要 root 权限" >&2
		exit 1
	}
}

ACTION="${1:-}"
PKG_PATH="${2:-}"
PKG_NAME="${PACKAGE_NAME:-gpam}"

if [[ -z "$ACTION" ]]; then
	usage
	exit 1
fi

case "$ACTION" in
install)
	if [[ -z "$PKG_PATH" || ! -f "$PKG_PATH" ]]; then
		echo "找不到包文件: $PKG_PATH" >&2
		exit 1
	fi
	require_root
	rpm -ivh --nosignature "$PKG_PATH"
	;;
upgrade)
	if [[ -z "$PKG_PATH" || ! -f "$PKG_PATH" ]]; then
		echo "找不到包文件: $PKG_PATH" >&2
		exit 1
	fi
	require_root
	rpm -Uvh --nosignature "$PKG_PATH"
	;;
reinstall)
	if [[ -z "$PKG_PATH" || ! -f "$PKG_PATH" ]]; then
		echo "找不到包文件: $PKG_PATH" >&2
		exit 1
	fi
	require_root
	# 允许相同版本重新安装
	rpm -Uvh --nosignature --replacepkgs "$PKG_PATH"
	;;
remove)
	require_root
	rpm -evh "$PKG_NAME"
	;;
*)
	usage
	exit 1
	;;
esac
