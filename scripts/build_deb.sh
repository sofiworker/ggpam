#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
DIST_DIR="${ROOT_DIR}/dist"
VERSION="${VERSION:-0.1.0}"
ARCH="${ARCH:-$(dpkg --print-architecture 2>/dev/null || uname -m)}"
STAGE="${DIST_DIR}/deb/ggpam_${VERSION}_${ARCH}"

"${ROOT_DIR}/scripts/build.sh"

CLI_BIN="${ROOT_DIR}/bin/ggpam"
PAM_SO="${ROOT_DIR}/bin/pam_ggpam.so"
PAM_HEADER="${ROOT_DIR}/bin/pam_ggpam.h"

for file in "$CLI_BIN" "$PAM_SO" "$PAM_HEADER"; do
	if [[ ! -e "$file" ]]; then
		echo "缺少构建产物: $file" >&2
		exit 1
	fi
done

rm -rf "$STAGE"
mkdir -p "$STAGE/DEBIAN"
mkdir -p "$STAGE/usr/bin"
mkdir -p "$STAGE/lib/security"
mkdir -p "$STAGE/usr/include/ggpam"

cat >"$STAGE/DEBIAN/control" <<EOF_CTRL
Package: ggpam
Version: ${VERSION}
Section: utils
Priority: optional
Architecture: ${ARCH}
Maintainer: sofiworker <sofiworker@outlook.com>
Description: Google Authenticator PAM 模块的 Go 实现
EOF_CTRL

cat >"$STAGE/DEBIAN/postinst" <<'EOF_POSTINST'
#!/bin/sh
set -e
if command -v ldconfig >/dev/null 2>&1; then
	ldconfig
fi
exit 0
EOF_POSTINST

cat >"$STAGE/DEBIAN/postrm" <<'EOF_POSTRM'
#!/bin/sh
set -e
case "$1" in
remove|purge)
	if command -v ldconfig >/dev/null 2>&1; then
		ldconfig
	fi
	;;
esac
exit 0
EOF_POSTRM

chmod 0755 "$STAGE/DEBIAN/postinst" "$STAGE/DEBIAN/postrm"

install -m 0755 "$CLI_BIN" "$STAGE/usr/bin/ggpam"
install -m 0644 "$PAM_SO" "$STAGE/lib/security/pam_ggpam.so"
install -m 0644 "$PAM_HEADER" "$STAGE/usr/include/gpam/pam_ggpam.h"

mkdir -p "$DIST_DIR"
PACKAGE="${DIST_DIR}/gpam_${VERSION}_${ARCH}.deb"
dpkg-deb --build "$STAGE" "$PACKAGE"
echo "DEB 已生成: $PACKAGE"
