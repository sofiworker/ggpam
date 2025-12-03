# ggpam

## 项目概述
- Go 版本的 Google Authenticator 实现，涵盖 CLI 与 PAM 模块，复刻原版速率限制、时间偏移自适应、应急码等行为。
- 代码模块化：配置解析（`pkg/config`）、验证器（`pkg/authenticator`/`pkg/otp`）、日志（`pkg/logging`）、PAM 参数解析与文件校验（`pkg/pam`）。
- 产物：静态 CLI 可执行文件 `ggpam`，以及可直接放入 PAM 的 `pam_ggpam.so`/`pam_ggpam.h`。
- 国际化：面向用户的提示走 i18n（支持中英文）；日志与错误消息保持英文便于程序化处理。

## 仓库结构
- `cmd/cli`：Cobra CLI，含 `init`/`verify`/`version`。
- `cmd/pam`：PAM 入口，使用 cgo 暴露 `pam_sm_authenticate`/`pam_sm_setcred`。
- `pkg/config`：解析/序列化 `~/.ggpam_authenticator`（兼容 `.google_authenticator`）格式。
- `pkg/authenticator`、`pkg/otp`：TOTP/HOTP 计算、应急码验证。
- `pkg/pam`：PAM 参数、密钥文件校验、持久化。
- `pkg/logging`：可配置文件+stderr 输出，支持环境变量。
- `scripts/`：依赖检查、构建、打包（deb/rpm）、Docker 内验证脚本。

## 快速开始
```bash
# 安装依赖（gcc/clang、pam 开发头、go1.18+）
./scripts/check_deps.sh

# 构建 CLI 与 PAM
make build            # 生成 bin/ggpam 和 bin/pam_ggpam.so/h

# 运行测试
GOCACHE=$(pwd)/.cache go test ./...   # 避免 /root/.cache 权限问题

# 查看版本信息
./bin/ggpam version
```

## CLI 使用
```bash
# 初始化配置（默认 ~/.ggpam_authenticator，支持交互确认）
./bin/ggpam init --mode totp --path ~/.ggpam_authenticator

# 验证一次性密码（可用 --code 或直接传参数）
./bin/ggpam verify --code 123456

# 静默验证（仅退出码反映结果）
./bin/ggpam verify --quiet --code 123456
```
常用参数：
- `--mode totp|hotp`、`--time-based/--counter-based`：选择模式。
- `--window-size`、`--step-size`：窗口与步长。
- `--rate-limit/--rate-time/--no-rate-limit`：速率限制。
- `--emergency-codes`：生成应急码数量。
- `--no-confirm`：跳过写文件确认（自动化场景）。
- `--qr-mode`/`--qr-inverse`/`--qr-utf8`：二维码输出样式。

## PAM 模块
1) 构建后将 `bin/pam_ggpam.so`/`bin/pam_ggpam.h` 安装到系统 PAM 目录（如 `/lib64/security`），并在目标服务的 pam.d 文件里添加：
   ```
   auth required pam_ggpam.so secret=/path/to/.ggpam_authenticator try_first_pass grace_period=30
   ```
2) 重要参数（见 `pkg/pam/params.go`）：
   - `secret=`：密钥文件模板，支持 `%u`/`%h`/`~`；默认 `~/.ggpam_authenticator`。
   - `try_first_pass`/`use_first_pass`/`forward_pass`：与现有密码交互的方式。
   - `prompt_template=`：自定义提示模板（可用 `{{.User}}`/`{{.Rhost}}` 等变量）。
   - `grace_period=`：宽限期（秒），允许同一主机在窗口内跳过验证。
   - `allowed_perm=`、`no_strict_owner`：文件权限与所有者校验。
   - `allow_readonly`：只读场景下忽略写入失败。
   - `debug`：输出调试日志。

## 日志与配置
- 环境变量：
  - `GGPAM_LOG_LEVEL`：`debug`/`info`/`warn`/`error`（默认 `info`）。
  - `GGPAM_LOG_FILE`：日志文件路径；未设置且 `DefaultHomeLogging=true` 时会写入 `$HOME/ggpam.log`，并同时输出到 stderr。
- PAM 调用会自动将日志写入 syslog，同步到 `pkg/logging` 输出。

## 构建与打包
- `make fmt` / `make test` / `make lint`：格式化、测试、vet。
- `make deb` / `make rpm`：调用 `scripts/build_deb.sh` / `scripts/build_rpm.sh` 生成包，产物位于 `dist/`。
- `./scripts/build.sh`：单次构建 CLI 与 PAM。
- `./scripts/verify_docker.sh`：在 Docker 中验证 deb/rpm 安装/卸载流程（需设置 `DEB_IMAGE`/`RPM_IMAGE`，可用 `SKIP_BUILD=true` 复用现有包）。
- 版本信息通过 `-ldflags` 注入（`Version`/`GitCommit`/`BuildDate`/`GoVersion`），`make build` 已内置。

## 开发指引
- Go 1.18+，遵循 idiomatic Go（tabs 缩进，错误上下文包装，避免 panic）。
- 涉及阻塞操作请将 `context` 作为首参；日志/错误保持英文，展示给用户的文本通过 i18n。
- 提交前运行 `gofmt`、`go test ./...`，如依赖变更请执行 `go mod tidy`。
