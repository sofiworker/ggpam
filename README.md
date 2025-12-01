## 项目简介

`gpam` 以 Go 语言重写 `google-authenticator-libpam`，涵盖 CLI、核心 OTP 逻辑与 PAM 模块：

- `pkg/config`：解析 `~/.google_authenticator` 文件，复刻速率限制、时间漂移自适应、应急码等功能；
- `pkg/authenticator` 与 `pkg/otp`：提供 TOTP/HOTP 计算、应急码验证，供 CLI 与 PAM 复用；
- `cmd/google-authenticator`：基于 Cobra 的命令行，包含 `init`、`verify`、`version` 等子命令，支持通过 `-ldflags` 注入版本、Git、构建时间、Go 版本信息；
- `cmd/pam_google_authenticator`：cgo 实现的 PAM 模块，提供 `pam_sm_authenticate`/`pam_sm_setcred` 并复用 Go 逻辑；
- 构建脚本、依赖检查、Deb/RPM 打包集中于 `scripts/` 与 `packaging/`。

## 开发与构建

```bash
# 依赖检查
./scripts/check_deps.sh

# 代码格式化 & 测试
make fmt
make test

# 构建 CLI 与 PAM 模块（默认注入 VERSION/GIT/DATE/GO 版本信息）
make build
# 或执行
./scripts/build.sh
```

构建完成后：

- `bin/google-authenticator`：CLI，可用 `google-authenticator version` 查看注入信息；
- `bin/pam_google_authenticator.so/.h`：PAM 模块共享对象与头文件（`go build -buildmode=c-shared` 生成）。

## 使用示例

```bash
# 初始化配置
google-authenticator init --mode totp --path ~/.google_authenticator
# 验证一次性密码（flag 或参数均可）
google-authenticator verify --code 123456
# 打印版本信息
google-authenticator version
```

## 打包

```bash
# 生成 deb/rpm（默认版本 0.1.0，可通过 env 覆盖 VERSION/RELEASE）
make package
# 或单独执行
make deb
make rpm
```

产物位于 `dist/`，包含 CLI、PAM `.so` 与头文件。示例：

```bash
VERSION=0.2.0 ./packaging/build_deb.sh
RELEASE=2 ./packaging/build_rpm.sh
```

### 安装/升级/卸载辅助脚本

- `./packaging/manage_deb.sh <install|upgrade|reinstall|remove|purge> dist/gpam_<ver>_<arch>.deb`
- `./packaging/manage_rpm.sh <install|upgrade|reinstall|remove> dist/gpam-<ver>-<release>.<arch>.rpm`

安装/升级场景用 `install`/`upgrade`，重复安装同版本用 `reinstall`，卸载使用 `remove`（deb 还可 `purge`）。

### Docker 内验证包

`DEB_IMAGE=debian:12 RPM_IMAGE=rockylinux:9 ./packaging/verify_docker.sh`

脚本会（默认）先构建 deb/rpm，再在对应容器里执行安装→重复安装→卸载流程，校验 CLI 与 PAM 模块文件的存在与清理；可通过 `SKIP_BUILD=true` 复用已有包。
