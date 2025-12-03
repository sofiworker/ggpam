#!/usr/bin/env bash
set -euo pipefail

# 在 Docker 中验证 deb/rpm 的安装、重复安装与卸载流程。
# 依赖：docker 客户端、可访问 Debian/Rocky 基础镜像。

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
DIST_DIR="${ROOT_DIR}/dist"
DEB_IMAGE="${DEB_IMAGE:-debian:12}"
RPM_IMAGE="${RPM_IMAGE:-rockylinux:9}"
SKIP_BUILD="${SKIP_BUILD:-false}"

if [[ "$SKIP_BUILD" != "true" ]]; then
	"${ROOT_DIR}/scripts/build_deb.sh"
	"${ROOT_DIR}/scripts/build_rpm.sh"
fi

DEB_PKG="${DEB_PKG:-$(ls -t "${DIST_DIR}"/gpam_*_*.deb 2>/dev/null | head -1 || true)}"
RPM_PKG="${RPM_PKG:-$(ls -t "${DIST_DIR}"/gpam-*.rpm 2>/dev/null | head -1 || true)}"

if [[ -z "$DEB_PKG" || ! -f "$DEB_PKG" ]]; then
	echo "未找到 deb 包，请先生成 dist/gpam_<version>_<arch>.deb" >&2
	exit 1
fi
if [[ -z "$RPM_PKG" || ! -f "$RPM_PKG" ]]; then
	echo "未找到 rpm 包，请先生成 dist/gpam-<version>-<release>.<arch>.rpm" >&2
	exit 1
fi

echo "使用 debian 镜像验证: $DEB_IMAGE"
docker run --rm -v "${DIST_DIR}:/pkgs:ro" "$DEB_IMAGE" bash -euo pipefail -c "
	apt-get update >/dev/null
	apt-get install -y --no-install-recommends libpam0g libpam-modules >/dev/null
	dpkg -i \"/pkgs/$(basename "$DEB_PKG")\"
	google-authenticator version >/dev/null
	test -f /lib/security/pam_google_authenticator.so
	# 重复安装（模拟 upgrade/reinstall 同版本）
	dpkg -i \"/pkgs/$(basename "$DEB_PKG")\"
	dpkg -r gpam
	if [[ -e /lib/security/pam_google_authenticator.so ]]; then
		echo '卸载后残留 pam 模块文件' >&2
		exit 1
	fi
	echo 'deb 包安装/重复安装/卸载验证完成'
"

echo "使用 rpm 镜像验证: $RPM_IMAGE"
docker run --rm -v "${DIST_DIR}:/pkgs:ro" "$RPM_IMAGE" bash -euo pipefail -c "
	dnf -y install /pkgs/$(basename "$RPM_PKG") >/dev/null
	test -f /lib/security/pam_google_authenticator.so
	/usr/bin/google-authenticator version >/dev/null
	# 重新安装（--replacepkgs 行为由 rpm -Uvh 决定）
	rpm -Uvh --replacepkgs /pkgs/$(basename "$RPM_PKG") >/dev/null
	rpm -e gpam
	if [[ -e /lib/security/pam_google_authenticator.so ]]; then
		echo '卸载后残留 pam 模块文件' >&2
		exit 1
	fi
	echo 'rpm 包安装/重复安装/卸载验证完成'
"

echo "Docker 包验证完成。"
