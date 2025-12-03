package i18n

import (
	"fmt"
	"os"
	"strings"
)

const (
	// PAM 相关
	MsgInvalidArgs                = "invalidArgs"
	MsgUserLookupFailed           = "userLookupFailed"
	MsgFallbackUser               = "fallbackUser"
	MsgDropPrivilegesFailed       = "dropPrivilegesFailed"
	MsgResolveSecretFailed        = "resolveSecretFailed"
	MsgUserNoSecretNullOK         = "userNoSecretNullOK"
	MsgReadConfigFailed           = "readConfigFailed"
	MsgPromptTemplateFailed       = "promptTemplateFailed"
	MsgGraceSkip                  = "graceSkip"
	MsgUserAuthFailed             = "userAuthFailed"
	MsgAuthFailedGeneric          = "authFailedGeneric"
	MsgInternalError              = "internalError"
	MsgUpdateAuthtokFailed        = "updateAuthtokFailed"
	MsgUserAuthSuccess            = "userAuthSuccess"
	MsgEmptyUsername              = "emptyUsername"
	MsgSerializeConfigFailed      = "serializeConfigFailed"
	MsgSecretChangedDuringProcess = "secretChangedDuringProcess"
	MsgSecretChangedRetry         = "secretChangedRetry"
	MsgReadonlyWriteIgnored       = "readonlyWriteIgnored"
	MsgWriteConfigFailed          = "writeConfigFailed"
	MsgUpdateConfigFailed         = "updateConfigFailed"
	MsgPromptTooLarge             = "promptTooLarge"
	MsgDummyPassword              = "dummyPassword"

	// CLI 相关
	MsgCliDisallowReusePrompt   = "cliDisallowReusePrompt"
	MsgCliTotpWindowPrompt      = "cliTotpWindowPrompt"
	MsgCliHotpWindowPrompt      = "cliHotpWindowPrompt"
	MsgCliRateLimitPrompt       = "cliRateLimitPrompt"
	MsgCliPromptTimeBased       = "cliPromptTimeBased"
	MsgCliAllowDisallowConflict = "cliAllowDisallowConflict"
	MsgCliCounterTimeConflict   = "cliCounterTimeConflict"
	MsgCliFileExistsWarn        = "cliFileExistsWarn"
	MsgCliStepRange             = "cliStepRange"
	MsgCliWindowRange           = "cliWindowRange"
	MsgCliScratchRange          = "cliScratchRange"
	MsgCliRateArgsMismatch      = "cliRateArgsMismatch"
	MsgCliRateLimitRange        = "cliRateLimitRange"
	MsgCliRateTimePositive      = "cliRateTimePositive"
	MsgCliRateTimeRange         = "cliRateTimeRange"
	MsgCliConfigCancelled       = "cliConfigCancelled"
	MsgCliConfigWritten         = "cliConfigWritten"
	MsgCliEnterCode             = "cliEnterCode"
	MsgCliCodeSkipped           = "cliCodeSkipped"
	MsgCliCodeInvalidDigits     = "cliCodeInvalidDigits"
	MsgCliCodeConfirmed         = "cliCodeConfirmed"
	MsgCliCodeIncorrect         = "cliCodeIncorrect"
	MsgCliUpdateFilePrompt      = "cliUpdateFilePrompt"
	MsgCliUnknownMode           = "cliUnknownMode"
	MsgCliHotpNoReuse           = "cliHotpNoReuse"
	MsgCliQRFail                = "cliQRFail"
	MsgCliSetupAddInfo          = "cliSetupAddInfo"
	MsgCliSetupURL              = "cliSetupURL"
	MsgCliSetupSecret           = "cliSetupSecret"
	MsgCliSetupTimeBased        = "cliSetupTimeBased"
	MsgCliSetupCounterBased     = "cliSetupCounterBased"
	MsgCliSetupManual           = "cliSetupManual"
	MsgCliScratchListHeader     = "cliScratchListHeader"
	MsgCliUsage                 = "cliUsage"
	MsgCliShort                 = "cliShort"
	MsgCliLong                  = "cliLong"
	MsgCmdInitShort             = "cmdInitShort"
	MsgCmdVerifyShort           = "cmdVerifyShort"
	MsgCliFlagHelp              = "cliFlagHelp"
	MsgCliFlagPath              = "cliFlagPath"
	MsgCliFlagSecret            = "cliFlagSecret"
	MsgCliFlagForce             = "cliFlagForce"
	MsgCliFlagMode              = "cliFlagMode"
	MsgCliFlagTimeBased         = "cliFlagTimeBased"
	MsgCliFlagCounterBased      = "cliFlagCounterBased"
	MsgCliFlagStepSize          = "cliFlagStepSize"
	MsgCliFlagWindowSize        = "cliFlagWindowSize"
	MsgCliFlagMinimalWindow     = "cliFlagMinimalWindow"
	MsgCliFlagRateLimit         = "cliFlagRateLimit"
	MsgCliFlagRateTime          = "cliFlagRateTime"
	MsgCliFlagDisableRate       = "cliFlagDisableRate"
	MsgCliFlagEmergencyCodes    = "cliFlagEmergencyCodes"
	MsgCliFlagScratchCodes      = "cliFlagScratchCodes"
	MsgCliFlagDisallowReuse     = "cliFlagDisallowReuse"
	MsgCliFlagAllowReuse        = "cliFlagAllowReuse"
	MsgCliFlagLabel             = "cliFlagLabel"
	MsgCliFlagIssuer            = "cliFlagIssuer"
	MsgCliFlagQuiet             = "cliFlagQuiet"
	MsgCliFlagQRMode            = "cliFlagQRMode"
	MsgCliFlagQRInverse         = "cliFlagQRInverse"
	MsgCliFlagQRUTF8            = "cliFlagQRUTF8"
	MsgCliFlagConfirm           = "cliFlagConfirm"
	MsgCliFlagNoConfirm         = "cliFlagNoConfirm"
	MsgCliFlagVerifyCode        = "cliFlagVerifyCode"
	MsgCliFlagNoSkew            = "cliFlagNoSkew"
	MsgCliFlagNoIncrementHOTP   = "cliFlagNoIncrementHOTP"
	MsgCliFlagVerifyQuiet       = "cliFlagVerifyQuiet"
	MsgCliVerifyNeedCode        = "cliVerifyNeedCode"
	MsgCliVerifyRateLimited     = "cliVerifyRateLimited"
	MsgCliVerifyScratchUsed     = "cliVerifyScratchUsed"
	MsgCliVerifyHOTPSuccess     = "cliVerifyHOTPSuccess"
	MsgCliVerifyTOTPSuccess     = "cliVerifyTOTPSuccess"
)

var translations = map[string]map[string]string{
	// PAM
	MsgInvalidArgs: {
		"en": "Invalid parameters: %v",
		"zh": "参数错误: %v",
	},
	MsgUserLookupFailed: {
		"en": "Failed to locate user %s: %v",
		"zh": "无法定位用户 %s: %v",
	},
	MsgFallbackUser: {
		"en": "Fallback to user %s for file access",
		"zh": "降级使用用户 %s 访问文件",
	},
	MsgDropPrivilegesFailed: {
		"en": "Failed to drop privileges to user %s: %v",
		"zh": "无法降级到用户 %s: %v",
	},
	MsgResolveSecretFailed: {
		"en": "Failed to resolve secret: %v",
		"zh": "无法解析密钥: %v",
	},
	MsgUserNoSecretNullOK: {
		"en": "User %s has no secret configured; nullok honored",
		"zh": "用户 %s 未配置密钥，nullok 生效",
	},
	MsgReadConfigFailed: {
		"en": "Failed to read %s: %v",
		"zh": "读取 %s 失败: %v",
	},
	MsgPromptTemplateFailed: {
		"en": "Failed to load prompt template: %v",
		"zh": "加载 prompt 模板失败: %v",
	},
	MsgGraceSkip: {
		"en": "Host %s is within grace period, skip verification",
		"zh": "主机 %s 在宽限期内，跳过验证码",
	},
	MsgUserAuthFailed: {
		"en": "User %s failed verification: %v",
		"zh": "用户 %s 验证失败: %v",
	},
	MsgAuthFailedGeneric: {
		"en": "Verification failed: %v",
		"zh": "验证失败: %v",
	},
	MsgInternalError: {
		"en": "Internal error",
		"zh": "内部错误",
	},
	MsgUpdateAuthtokFailed: {
		"en": "Failed to update PAM_AUTHTOK",
		"zh": "无法更新 PAM_AUTHTOK",
	},
	MsgUserAuthSuccess: {
		"en": "User %s authenticated (%s)",
		"zh": "用户 %s 验证成功 (%s)",
	},
	MsgEmptyUsername: {
		"en": "username is empty",
		"zh": "用户名为空",
	},
	MsgSerializeConfigFailed: {
		"en": "Failed to serialize config: %v",
		"zh": "序列化配置失败: %v",
	},
	MsgSecretChangedDuringProcess: {
		"en": "Secret file changed during processing, please retry",
		"zh": "密钥文件在处理期间发生变化，请重试",
	},
	MsgSecretChangedRetry: {
		"en": "Secret file changed, please retry",
		"zh": "密钥文件发生变化，请重试",
	},
	MsgReadonlyWriteIgnored: {
		"en": "Readonly mode; ignoring write failure: %v",
		"zh": "只读模式，忽略写入失败: %v",
	},
	MsgWriteConfigFailed: {
		"en": "Failed to write %s: %v",
		"zh": "写入 %s 失败: %v",
	},
	MsgUpdateConfigFailed: {
		"en": "Failed to update Google Authenticator config",
		"zh": "更新 Google Authenticator 配置失败",
	},
	MsgPromptTooLarge: {
		"en": "Prompt template exceeds %d bytes",
		"zh": "prompt 模板超过 %d 字节",
	},
	MsgDummyPassword: {
		"en": "Dummy password supplied by PAM. Did OpenSSH 'PermitRootLogin <anything but yes>' or some other config block this login?",
		"zh": "PAM 收到哑密码。请检查 OpenSSH 的 PermitRootLogin 或其他配置是否阻止登录。",
	},

	// CLI
	MsgCliDisallowReusePrompt: {
		"en": "Do you want to disallow multiple uses of the same authentication token? This restricts you to one login about every 30s, but it increases your chances to notice or even prevent man-in-the-middle attacks",
		"zh": "是否禁止同一 TOTP 重复使用？这会将登录限制为约每 30 秒一次，有助于发现并阻止中间人攻击。",
	},
	MsgCliTotpWindowPrompt: {
		"en": "By default, a new token is generated every 30 seconds by the mobile app. To compensate time skew, an extra token before and after the current time is allowed (window=3). Increase to 17 to allow ~4 minutes skew. Do you want to do so?",
		"zh": "默认每 30 秒生成一次新验证码，窗口大小为 3（当前码及前后各 1 个）。可将窗口增大到 17 以容忍约 4 分钟时间偏移，是否增大？",
	},
	MsgCliHotpWindowPrompt: {
		"en": "By default, three tokens are valid at any one time. Increase window to 17 to reduce sync issues. Do you want to do so?",
		"zh": "默认同时允许 3 个 HOTP 码。可将窗口增大到 17 以减少同步问题，是否增大？",
	},
	MsgCliRateLimitPrompt: {
		"en": "Enable rate limiting? Defaults to 3 attempts every 30s if enabled.",
		"zh": "是否启用速率限制？启用后默认每 30 秒最多 3 次尝试。",
	},
	MsgCliPromptTimeBased: {
		"en": "Do you want authentication tokens to be time-based",
		"zh": "是否使用基于时间的验证码（TOTP）",
	},
	MsgCliAllowDisallowConflict: {
		"en": "allow-reuse and disallow-reuse are mutually exclusive",
		"zh": "allow-reuse 与 disallow-reuse 互斥",
	},
	MsgCliCounterTimeConflict: {
		"en": "counter-based and time-based are mutually exclusive",
		"zh": "counter-based 与 time-based 互斥",
	},
	MsgCliFileExistsWarn: {
		"en": "Warning: %s exists and will be overwritten.",
		"zh": "警告: %s 已存在，将被覆盖。",
	},
	MsgCliStepRange: {
		"en": "step-size must be in 1..60",
		"zh": "step-size 需在 1..60 范围内",
	},
	MsgCliWindowRange: {
		"en": "window-size must be in 1..21",
		"zh": "window-size 需在 1..21 范围内",
	},
	MsgCliScratchRange: {
		"en": "emergency-codes must be in 0..%d",
		"zh": "emergency-codes 需在 0..%d 范围内",
	},
	MsgCliRateArgsMismatch: {
		"en": "Both --rate-limit and --rate-time must be set together",
		"zh": "--rate-limit 与 --rate-time 必须同时设置",
	},
	MsgCliRateLimitRange: {
		"en": "rate-limit must be in 1..10",
		"zh": "rate-limit 需在 1..10 范围内",
	},
	MsgCliRateTimePositive: {
		"en": "rate-time must be positive",
		"zh": "rate-time 必须为正数",
	},
	MsgCliRateTimeRange: {
		"en": "rate-time must be in 15..600 seconds",
		"zh": "rate-time 需在 15..600 秒",
	},
	MsgCliConfigCancelled: {
		"en": "Cancelled; not updating %s",
		"zh": "已取消，不会更新 %s",
	},
	MsgCliConfigWritten: {
		"en": "Config written to %s",
		"zh": "已写入配置 %s",
	},
	MsgCliUpdateFilePrompt: {
		"en": "Do you want me to update your \"%s\" file?",
		"zh": "是否更新 \"%s\" 文件？",
	},
	MsgCliUnknownMode: {
		"en": "Unknown mode: %s",
		"zh": "未知模式: %s",
	},
	MsgCliHotpNoReuse: {
		"en": "-d/-D is not supported in HOTP mode",
		"zh": "HOTP 模式下不支持 -d/-D",
	},
	MsgCliEnterCode: {
		"en": "Enter code from app (-1 to skip): ",
		"zh": "请输入手机验证码（-1 跳过）：",
	},
	MsgCliCodeSkipped: {
		"en": "Code confirmation skipped",
		"zh": "已跳过验证码确认",
	},
	MsgCliCodeInvalidDigits: {
		"en": "code must be 6 digits",
		"zh": "验证码必须为 6 位数字",
	},
	MsgCliCodeConfirmed: {
		"en": "Code confirmed",
		"zh": "验证码确认成功",
	},
	MsgCliCodeIncorrect: {
		"en": "Code incorrect (example valid code %s). Try again.",
		"zh": "验证码不正确（示例正确值 %s），请重试。",
	},
	MsgCliSetupAddInfo: {
		"en": "Add the following information to your authenticator app:",
		"zh": "请在身份验证器应用中添加以下信息：",
	},
	MsgCliSetupURL: {
		"en": "otpauth URL: %s",
		"zh": "otpauth URL: %s",
	},
	MsgCliSetupSecret: {
		"en": "Your new secret key is: %s",
		"zh": "新的密钥为: %s",
	},
	MsgCliSetupTimeBased: {
		"en": "This secret is time-based and will generate a new code every 30 seconds.",
		"zh": "该密钥为时间同步模式，每 30 秒生成一个新验证码。",
	},
	MsgCliSetupCounterBased: {
		"en": "This secret is counter-based. Each code can only be used once.",
		"zh": "该密钥为计数器模式，每个验证码只能使用一次。",
	},
	MsgCliSetupManual: {
		"en": "If you are using a mobile client, scan the QR code above or enter the secret manually.",
		"zh": "如使用手机客户端，请扫描上方二维码或手动输入密钥。",
	},
	MsgCliScratchListHeader: {
		"en": "Emergency codes:",
		"zh": "应急码：",
	},
	MsgCliQRFail: {
		"en": "Failed to generate QR code: %v",
		"zh": "生成二维码失败: %v",
	},
	MsgCmdInitShort: {
		"en": "Initialize ~/.google_authenticator config",
		"zh": "初始化 ~/.google_authenticator 配置",
	},
	MsgCmdVerifyShort: {
		"en": "Verify one-time password or emergency code",
		"zh": "验证一次性密码或应急码",
	},
	MsgCliFlagHelp: {
		"en": "Show help",
		"zh": "显示帮助",
	},
	MsgCliFlagPath: {
		"en": "Config file path",
		"zh": "配置文件路径",
	},
	MsgCliFlagSecret: {
		"en": "Specify secret file path",
		"zh": "指定密钥文件路径",
	},
	MsgCliFlagForce: {
		"en": "Do not prompt before writing file",
		"zh": "写文件前不再提示确认",
	},
	MsgCliFlagMode: {
		"en": "Auth mode: totp or hotp (default interactive)",
		"zh": "认证模式: totp 或 hotp（默认交互选择）",
	},
	MsgCliFlagTimeBased: {
		"en": "Use time-based (TOTP) mode",
		"zh": "设置为时间同步模式 (TOTP)",
	},
	MsgCliFlagCounterBased: {
		"en": "Use counter-based (HOTP) mode",
		"zh": "设置为计数器模式 (HOTP)",
	},
	MsgCliFlagStepSize: {
		"en": "TOTP step size (seconds), range 1..60",
		"zh": "TOTP 步长（秒），范围 1..60",
	},
	MsgCliFlagWindowSize: {
		"en": "Window size (1..21); prompt if unset",
		"zh": "窗口大小 (1..21)，不指定则交互提示",
	},
	MsgCliFlagMinimalWindow: {
		"en": "Use minimal window (TOTP=3, HOTP=1)",
		"zh": "使用最小窗口 (TOTP=3, HOTP=1)",
	},
	MsgCliFlagRateLimit: {
		"en": "Allowed attempts per window (1..10)",
		"zh": "每个窗口允许的登录次数 (1..10)",
	},
	MsgCliFlagRateTime: {
		"en": "Rate limit window length (15s..600s)",
		"zh": "速率限制窗口长度 (15s..600s)",
	},
	MsgCliFlagDisableRate: {
		"en": "Disable rate limiting",
		"zh": "禁用速率限制",
	},
	MsgCliFlagEmergencyCodes: {
		"en": "Number of emergency codes (0..10)",
		"zh": "应急码数量 (0..10)",
	},
	MsgCliFlagScratchCodes: {
		"en": "Number of emergency codes (compat flag)",
		"zh": "应急码数量 (兼容旧参数)",
	},
	MsgCliFlagDisallowReuse: {
		"en": "Disallow reuse of TOTP",
		"zh": "禁止重复使用 TOTP",
	},
	MsgCliFlagAllowReuse: {
		"en": "Allow reuse of TOTP",
		"zh": "允许重复使用 TOTP",
	},
	MsgCliFlagLabel: {
		"en": "Label for otpauth URL",
		"zh": "otpauth URL 的 label",
	},
	MsgCliFlagIssuer: {
		"en": "Issuer for otpauth URL",
		"zh": "otpauth URL 的 issuer",
	},
	MsgCliFlagQuiet: {
		"en": "Quiet mode, only essential output",
		"zh": "静默模式，仅输出必要信息",
	},
	MsgCliFlagQRMode: {
		"en": "QR output mode: none/ansi/ansi-inverse/ansi-grey/utf8/utf8-inverse/utf8-grey",
		"zh": "二维码输出模式: none/ansi/ansi-inverse/ansi-grey/utf8/utf8-inverse/utf8-grey",
	},
	MsgCliFlagQRInverse: {
		"en": "Invert QR colors (compat flag)",
		"zh": "二维码反色显示 (兼容参数)",
	},
	MsgCliFlagQRUTF8: {
		"en": "Render QR using UTF8 (compat flag)",
		"zh": "二维码使用 UTF8 渲染 (兼容参数)",
	},
	MsgCliFlagConfirm: {
		"en": "Require code confirmation after generation",
		"zh": "生成后要求输入验证码确认",
	},
	MsgCliFlagNoConfirm: {
		"en": "Skip code confirmation (non-interactive)",
		"zh": "不要求验证码确认 (适合非交互环境)",
	},
	MsgCliFlagVerifyCode: {
		"en": "6-digit OTP or 8-digit scratch code; can be provided as arg",
		"zh": "待验证的 6 位验证码或 8 位应急码，也可作为参数提供",
	},
	MsgCliFlagNoSkew: {
		"en": "Disable automatic time-skew detection",
		"zh": "禁用自动时间偏移探测",
	},
	MsgCliFlagNoIncrementHOTP: {
		"en": "Do not advance HOTP counter on failure",
		"zh": "失败后不推进 HOTP 计数器",
	},
	MsgCliFlagVerifyQuiet: {
		"en": "Suppress success output",
		"zh": "不输出成功提示",
	},
	MsgCliVerifyNeedCode: {
		"en": "Please provide code via --code or argument",
		"zh": "请通过 --code 或参数提供验证码",
	},
	MsgCliVerifyRateLimited: {
		"en": "Too many login attempts; please retry later",
		"zh": "当前登录次数过多，请稍后再试",
	},
	MsgCliVerifyScratchUsed: {
		"en": "Scratch code used; please refill emergency codes soon",
		"zh": "已使用应急码，建议尽快补充新的应急码",
	},
	MsgCliVerifyHOTPSuccess: {
		"en": "HOTP verification success, counter=%d",
		"zh": "HOTP 验证成功，当前计数器=%d",
	},
	MsgCliVerifyTOTPSuccess: {
		"en": "TOTP verification success",
		"zh": "TOTP 验证成功",
	},
	MsgCliUsage: {
		"en": `google-authenticator %s
Usage:
  google-authenticator [options]

Options:
  -h, --help                        Print this message
      --version                     Print version
  -c, --counter-based               Set up counter-based (HOTP) verification
  -C, --no-confirm                  Don't confirm code. For non-interactive setups
  -t, --time-based                  Set up time-based (TOTP) verification
  -d, --disallow-reuse              Disallow reuse of previously used TOTP tokens
  -D, --allow-reuse                 Allow reuse of previously used TOTP tokens
  -f, --force                       Write file without first confirming with user
  -l, --label=<label>               Override the default label in "otpauth://" URL
  -i, --issuer=<issuer>             Override the default issuer in "otpauth://" URL
  -q, --quiet                       Quiet mode
  -Q, --qr-mode=MODE                QRCode output mode
  -r, --rate-limit=N                Limit logins to N per every M seconds
  -R, --rate-time=M                 Limit logins to N per every M seconds
  -u, --no-rate-limit               Disable rate-limiting
  -s, --secret=<file>               Specify a non-standard file location
  -S, --step-size=S                 Set interval between token refreshes
  -w, --window-size=W               Set window of concurrently valid codes
  -W, --minimal-window              Disable window of concurrently valid codes
  -e, --emergency-codes=N           Number of emergency codes to generate`,
		"zh": `google-authenticator %s
用法:
  google-authenticator [选项]

选项:
  -h, --help                        显示帮助
      --version                     显示版本
  -c, --counter-based               设置为计数器模式 (HOTP)
  -C, --no-confirm                  不要求验证码确认，适合非交互环境
  -t, --time-based                  设置为时间同步模式 (TOTP)
  -d, --disallow-reuse              禁止重复使用已用过的 TOTP
  -D, --allow-reuse                 允许重复使用 TOTP
  -f, --force                       写文件前不再提示确认
  -l, --label=<label>               覆盖 otpauth:// URL 的 label
  -i, --issuer=<issuer>             覆盖 otpauth:// URL 的 issuer
  -q, --quiet                       静默模式
  -Q, --qr-mode=MODE                二维码输出模式
  -r, --rate-limit=N                每 M 秒最多 N 次登录
  -R, --rate-time=M                 配合 --rate-limit 设置时间窗口（秒）
  -u, --no-rate-limit               禁用速率限制
  -s, --secret=<file>               自定义密钥文件路径
  -S, --step-size=S                 设置 TOTP 刷新间隔（秒）
  -w, --window-size=W               设置允许同时有效的验证码数量
  -W, --minimal-window              使用最小窗口
  -e, --emergency-codes=N           生成的应急码数量`,
	},
	MsgCliShort: {
		"en": "Google Authenticator (Go version)",
		"zh": "Google Authenticator (Go 版本)",
	},
	MsgCliLong: {
		"en": "Google Authenticator CLI provides initialization and verification utilities.",
		"zh": "Google Authenticator CLI，提供配置初始化与验证码验证功能。",
	},
}

// Msgf returns the formatted translation.
func Msgf(key string, args ...interface{}) string {
	format := Resolve(key)
	if len(args) == 0 {
		return format
	}
	return fmt.Sprintf(format, args...)
}

// Resolve returns the translation for key, falling back to en or key itself.
func Resolve(key string) string {
	lang := DetectLang()
	if text := translations[key][lang]; text != "" {
		return text
	}
	if text := translations[key]["en"]; text != "" {
		return text
	}
	return key
}

// DetectLang reads locale env vars and normalizes to "en" / "zh".
func DetectLang() string {
	for _, env := range []string{"LC_ALL", "LC_MESSAGES", "LANG"} {
		if v := os.Getenv(env); v != "" {
			return normalizeLocale(v)
		}
	}
	return "en"
}

func normalizeLocale(locale string) string {
	lower := strings.ToLower(locale)
	if idx := strings.IndexAny(lower, "._@"); idx >= 0 {
		lower = lower[:idx]
	}
	switch {
	case strings.HasPrefix(lower, "zh"):
		return "zh"
	default:
		return "en"
	}
}
