#!/usr/bin/env bash
set -euo pipefail

usage() {
	cat <<'USAGE'
用法: manage_deb.sh <install|upgrade|reinstall|remove|purge> <package.deb>
环境变量:
  PACKAGE_NAME  目标包名（默认 gpam）
示例:
  sudo ./packaging/manage_deb.sh install dist/gpam_0.1.0_amd64.deb
  sudo ./packaging/manage_deb.sh reinstall dist/gpam_0.1.0_amd64.deb
  sudo ./packaging/manage_deb.sh remove
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
install | upgrade | reinstall)
	if [[ -z "$PKG_PATH" || ! -f "$PKG_PATH" ]]; then
		echo "找不到包文件: $PKG_PATH" >&2
		exit 1
	fi
	require_root
	if [[ "$ACTION" == "reinstall" ]]; then
		# dpkg 会在版本相同场景重新解包，--force-confnew 确保覆盖配置
		dpkg -i --force-confnew "$PKG_PATH"
	else
		dpkg -i "$PKG_PATH"
	fi
	;;
remove)
	require_root
	dpkg -r "$PKG_NAME"
	;;
purge)
	require_root
	dpkg -P "$PKG_NAME"
	;;
*)
	usage
	exit 1
	;;
esac
