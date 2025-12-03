#!/usr/bin/env bash
set -euo pipefail

# 在多个 Docker 镜像中验证 deb/rpm 安装，并通过 sshd 的 PAM 链路自动化测试 OTP 成功/失败场景。
# 依赖：本机已安装 docker，且能拉取 Debian/Ubuntu/Rocky 基础镜像；容器内会安装 openssh-server、pamtester、oathtool。

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
DIST_DIR="${ROOT_DIR}/dist"
DEFAULT_DEB_IMAGES="debian:12 debian:13 ubuntu:22.04 ubuntu:24.04"
DEFAULT_RPM_IMAGES="rockylinux:8 rockylinux:9"
DEB_IMAGES_STR="${DEB_IMAGES:-$DEFAULT_DEB_IMAGES}"
RPM_IMAGES_STR="${RPM_IMAGES:-$DEFAULT_RPM_IMAGES}"
IFS=' ' read -r -a DEB_IMAGES <<< "$DEB_IMAGES_STR"
IFS=' ' read -r -a RPM_IMAGES <<< "$RPM_IMAGES_STR"
SKIP_BUILD="${SKIP_BUILD:-false}"
SSH_USER="${SSH_USER:-root}"
SSH_PASSWORD="${SSH_PASSWORD:-P@ssw0rd!}"
USE_CHINA_MIRROR="${USE_CHINA_MIRROR:-false}"

need_cmd() {
	if ! command -v "$1" >/dev/null 2>&1; then
		echo "缺少依赖: $1" >&2
		exit 1
	fi
}

need_cmd docker

find_latest_pkg() {
	local pattern="$1"
	ls -t "${DIST_DIR}"/${pattern} 2>/dev/null | head -1 || true
}

if [[ "$SKIP_BUILD" != "true" ]]; then
	"${ROOT_DIR}/scripts/build_deb.sh"
	"${ROOT_DIR}/scripts/build_rpm.sh"
fi

DEB_PKG="${DEB_PKG:-$(find_latest_pkg "ggpam_*_*.deb")}"
RPM_PKG="${RPM_PKG:-$(find_latest_pkg "ggpam-*.rpm")}"

if [[ -z "$DEB_PKG" || ! -f "$DEB_PKG" ]]; then
	echo "未找到 deb 包，请先生成 dist/ggpam_<version>_<arch>.deb" >&2
	exit 1
fi
if [[ -z "$RPM_PKG" || ! -f "$RPM_PKG" ]]; then
	echo "未找到 rpm 包，请先生成 dist/ggpam-<version>-<release>.<arch>.rpm" >&2
	exit 1
fi

container_script() {
	cat <<'EOF'
set -euo pipefail

install_deb() {
	setup_apt_mirror
	apt-get update >/dev/null
	DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends \
		openssh-server pamtester oathtool libpam0g libpam-modules ca-certificates >/dev/null
	dpkg -i "/pkgs/${PKG}"
}

install_rpm() {
	setup_dnf_mirror
	dnf -y install epel-release >/dev/null
	patch_epel_mirror
	dnf -y install openssh-server pamtester oathtool >/dev/null
	rpm -Uvh --replacepkgs "/pkgs/${PKG}" >/dev/null
}

setup_apt_mirror() {
	if [[ "${USE_CHINA_MIRROR}" != "true" ]]; then
		return
	fi
	if [[ ! -f /etc/os-release ]]; then
		return
	fi
	source /etc/os-release
	case "$ID" in
	debian)
		if [[ -f /etc/apt/sources.list ]]; then
			sed -i 's@deb.debian.org@mirrors.ustc.edu.cn@g' /etc/apt/sources.list || true
		fi
		if [[ -f /etc/apt/sources.list.d/debian.sources ]]; then
			sed -i 's@deb.debian.org@mirrors.ustc.edu.cn@g' /etc/apt/sources.list.d/debian.sources || true
		fi
		;;
	ubuntu)
		if [[ -f /etc/apt/sources.list ]]; then
			sed -i 's@//.*archive.ubuntu.com@//mirrors.ustc.edu.cn@g' /etc/apt/sources.list || true
		fi
		if [[ -f /etc/apt/sources.list.d/ubuntu.sources ]]; then
			sed -i 's@//.*archive.ubuntu.com@//mirrors.ustc.edu.cn@g' /etc/apt/sources.list.d/ubuntu.sources || true
		fi
		;;
	*)
		;;
	esac
}

setup_dnf_mirror() {
	if [[ "${USE_CHINA_MIRROR}" != "true" ]]; then
		return
	fi
	if [[ ! -f /etc/os-release ]]; then
		return
	fi
	source /etc/os-release
	case "$ID" in
	rocky)
		case "${VERSION_ID%%.*}" in
		8)
			sed -e 's|^mirrorlist=|#mirrorlist=|g' \
				-e 's|^#baseurl=http://dl.rockylinux.org/$contentdir|baseurl=https://mirrors.ustc.edu.cn/rocky|g' \
				-i.bak \
				/etc/yum.repos.d/Rocky-AppStream.repo \
				/etc/yum.repos.d/Rocky-BaseOS.repo \
				/etc/yum.repos.d/Rocky-Extras.repo \
				/etc/yum.repos.d/Rocky-PowerTools.repo || true
			;;
		9)
			sed -e 's|^mirrorlist=|#mirrorlist=|g' \
				-e 's|^#baseurl=http://dl.rockylinux.org/$contentdir|baseurl=https://mirrors.ustc.edu.cn/rocky|g' \
				-i.bak \
				/etc/yum.repos.d/rocky-extras.repo \
				/etc/yum.repos.d/rocky.repo || true
			;;
		esac
		;;
	fedora)
		sed -e 's|^metalink=|#metalink=|g' \
			-e 's|^#baseurl=http://download.example/pub/fedora/linux|baseurl=https://mirrors.ustc.edu.cn/fedora|g' \
			-i.bak \
			/etc/yum.repos.d/fedora.repo \
			/etc/yum.repos.d/fedora-updates.repo || true
		;;
	esac
}

patch_epel_mirror() {
	if [[ "${USE_CHINA_MIRROR}" != "true" ]]; then
		return
	fi
	if [[ -f /etc/yum.repos.d/epel.repo ]]; then
		sed -i.bak -E \
			-e 's|^metalink=|#metalink=|g' \
			-e 's|^#?baseurl=https?://[^/]+/pub/epel|baseurl=https://mirrors.ustc.edu.cn/epel|g' \
			/etc/yum.repos.d/epel.repo /etc/yum.repos.d/epel-testing.repo 2>/dev/null || true
	fi
}

ensure_pam_module() {
	if [[ -f /lib/security/pam_ggpam.so ]]; then
		if [[ -d /lib/x86_64-linux-gnu/security && ! -e /lib/x86_64-linux-gnu/security/pam_ggpam.so ]]; then
			ln -s /lib/security/pam_ggpam.so /lib/x86_64-linux-gnu/security/pam_ggpam.so 2>/dev/null || true
		fi
		if [[ -d /lib64/security && ! -e /lib64/security/pam_ggpam.so ]]; then
			ln -s /lib/security/pam_ggpam.so /lib64/security/pam_ggpam.so 2>/dev/null || true
		fi
	fi
	if [[ -f /lib/x86_64-linux-gnu/security/pam_ggpam.so && ! -e /lib/security/pam_ggpam.so ]]; then
		ln -s /lib/x86_64-linux-gnu/security/pam_ggpam.so /lib/security/pam_ggpam.so 2>/dev/null || true
	fi
	if [[ -f /lib64/security/pam_ggpam.so && ! -e /lib/security/pam_ggpam.so ]]; then
		ln -s /lib64/security/pam_ggpam.so /lib/security/pam_ggpam.so 2>/dev/null || true
	fi
	local module=""
	for path in /lib/security/pam_ggpam.so /lib64/security/pam_ggpam.so /lib/x86_64-linux-gnu/security/pam_ggpam.so; do
		if [[ -f "$path" ]]; then
			module="$path"
			break
		fi
	done
	if [[ -z "$module" ]]; then
		echo "未找到 pam_ggpam.so" >&2
		exit 1
	fi
	PAM_GGPAM_PATH="$module"
}

prepare_auth_env() {
	echo "${SSH_USER}:${SSH_PASSWORD}" | chpasswd
	ggpam init --force --no-confirm --quiet --no-rate-limit --time-based \
		--disallow-reuse --window-size 3 --label docker --issuer docker \
		--path "/root/.ggpam_authenticator"

	local module="${PAM_GGPAM_PATH:-pam_ggpam.so}"
	sed -i "1i auth required ${module} secret=/root/.ggpam_authenticator debug" /etc/pam.d/sshd
	if [[ -f /etc/ssh/sshd_config ]]; then
		if grep -q '^UsePAM' /etc/ssh/sshd_config; then
			sed -i 's/^UsePAM.*/UsePAM yes/' /etc/ssh/sshd_config
		else
			echo "UsePAM yes" >> /etc/ssh/sshd_config
		fi
	fi
}

run_pam_tests() {
	local secret otp wrong_code
	secret="$(head -n1 /root/.ggpam_authenticator | tr -d ' ')"
	otp="$(oathtool --totp -b "$secret")"
	wrong_code="$(printf "%06d" $(((10#$otp + 1) % 1000000)))"

	if ! printf "%s\n%s\n" "$otp" "$SSH_PASSWORD" | pamtester sshd "$SSH_USER" authenticate >/tmp/pamtester.ok 2>&1; then
		echo "PAM OTP 验证失败" >&2
		cat /tmp/pamtester.ok >&2
		exit 1
	fi

	if printf "%s\n%s\n" "$wrong_code" "$SSH_PASSWORD" | pamtester sshd "$SSH_USER" authenticate >/tmp/pamtester.bad 2>&1; then
		echo "错误验证码未触发失败" >&2
		cat /tmp/pamtester.bad >&2
		exit 1
	fi
}

if [[ "$PKG_TYPE" == "deb" ]]; then
	install_deb
else
	install_rpm
fi

ggpam version >/dev/null
ensure_pam_module
prepare_auth_env
run_pam_tests

echo "镜像 ${PKG_TYPE}-${IMAGE} 验证通过"
EOF
}

run_container() {
	local image="$1"
	local pkg="$2"
	local pkg_type="$3"

	echo "==> 验证 ${pkg_type} 镜像: ${image}"
	docker run --rm \
		-e PKG="$(basename "$pkg")" \
		-e PKG_TYPE="$pkg_type" \
		-e IMAGE="$image" \
		-e SSH_USER="$SSH_USER" \
		-e SSH_PASSWORD="$SSH_PASSWORD" \
		-e USE_CHINA_MIRROR="$USE_CHINA_MIRROR" \
		-v "${DIST_DIR}:/pkgs:ro" \
		"$image" bash -euo pipefail -c "$(container_script)"
}

for image in "${DEB_IMAGES[@]}"; do
	run_container "$image" "$DEB_PKG" "deb"
done

for image in "${RPM_IMAGES[@]}"; do
	run_container "$image" "$RPM_PKG" "rpm"
done

echo "所有 Docker 包与 PAM/sshd 认证验证完成。"
