#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
DIST_DIR="${ROOT_DIR}/dist"
BUILD_ROOT="${DIST_DIR}/rpm"
VERSION="${VERSION:-0.1.0}"
RELEASE="${RELEASE:-1}"
ARCH="${ARCH:-$(uname -m)}"

"${ROOT_DIR}/scripts/build.sh"

CLI_BIN="${ROOT_DIR}/bin/google-authenticator"
PAM_SO="${ROOT_DIR}/bin/pam_google_authenticator.so"
PAM_HEADER="${ROOT_DIR}/bin/pam_google_authenticator.h"

for file in "$CLI_BIN" "$PAM_SO" "$PAM_HEADER"; do
	if [[ ! -e "$file" ]]; then
		echo "缺少构建产物: $file" >&2
		exit 1
	fi
done

RPMROOT="${BUILD_ROOT}/rpmbuild"
for dir in BUILD RPMS SOURCES SPECS SRPMS; do
	mkdir -p "${RPMROOT}/${dir}"
done

SRC_DIR="${BUILD_ROOT}/src/gpam-${VERSION}"
rm -rf "$SRC_DIR"
mkdir -p "$SRC_DIR"
cp "$CLI_BIN" "$SRC_DIR/google-authenticator"
cp "$PAM_SO" "$SRC_DIR/pam_google_authenticator.so"
cp "$PAM_HEADER" "$SRC_DIR/pam_google_authenticator.h"
tar -C "$(dirname "$SRC_DIR")" -czf "${RPMROOT}/SOURCES/gpam-${VERSION}.tar.gz" "gpam-${VERSION}"

SPEC_FILE="${RPMROOT}/SPECS/gpam.spec"
cat >"$SPEC_FILE" <<EOF_SPEC
Name:           gpam
Version:        ${VERSION}
Release:        ${RELEASE}%{?dist}
Summary:        Google Authenticator PAM module (Go rewrite)

License:        Apache-2.0
URL:            https://github.com/example/gpam
Source0:        gpam-${VERSION}.tar.gz
Requires(post): /sbin/ldconfig
Requires(postun): /sbin/ldconfig

%description
Go 语言重写的 google-authenticator-libpam，提供 CLI 与 PAM 模块。

%prep
%setup -q

%build
# 预编译二进制由外部脚本提供

%install
install -D -m 0755 google-authenticator %{buildroot}/usr/bin/google-authenticator
install -D -m 0644 pam_google_authenticator.so %{buildroot}/lib/security/pam_google_authenticator.so
install -D -m 0644 pam_google_authenticator.h %{buildroot}/usr/include/gpam/pam_google_authenticator.h

%post
/sbin/ldconfig

%postun
/sbin/ldconfig

%files
/usr/bin/google-authenticator
/lib/security/pam_google_authenticator.so
/usr/include/gpam/pam_google_authenticator.h

%changelog
* $(date +"%a %b %d %Y") gpam <devnull@example.com> - ${VERSION}-${RELEASE}
- 初始构建
EOF_SPEC

rpmbuild --define "_topdir ${RPMROOT}" -bb "$SPEC_FILE"

mkdir -p "$DIST_DIR"
cp "${RPMROOT}/RPMS/${ARCH}/"*.rpm "$DIST_DIR"/
echo "RPM 已生成: $DIST_DIR"
