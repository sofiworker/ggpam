package main

import (
	"fmt"
	"os"
	"strings"
)

const (
	msgInvalidArgs                = "invalidArgs"
	msgUserLookupFailed           = "userLookupFailed"
	msgFallbackUser               = "fallbackUser"
	msgDropPrivilegesFailed       = "dropPrivilegesFailed"
	msgResolveSecretFailed        = "resolveSecretFailed"
	msgUserNoSecretNullOK         = "userNoSecretNullOK"
	msgReadConfigFailed           = "readConfigFailed"
	msgPromptTemplateFailed       = "promptTemplateFailed"
	msgGraceSkip                  = "graceSkip"
	msgUserAuthFailed             = "userAuthFailed"
	msgAuthFailedGeneric          = "authFailedGeneric"
	msgInternalError              = "internalError"
	msgUpdateAuthtokFailed        = "updateAuthtokFailed"
	msgUserAuthSuccess            = "userAuthSuccess"
	msgEmptyUsername              = "emptyUsername"
	msgSerializeConfigFailed      = "serializeConfigFailed"
	msgSecretChangedDuringProcess = "secretChangedDuringProcess"
	msgSecretChangedRetry         = "secretChangedRetry"
	msgReadonlyWriteIgnored       = "readonlyWriteIgnored"
	msgWriteConfigFailed          = "writeConfigFailed"
	msgUpdateConfigFailed         = "updateConfigFailed"
	msgPromptTooLarge             = "promptTooLarge"
	msgDummyPassword              = "dummyPassword"
)

var translations = map[string]map[string]string{
	msgInvalidArgs: {
		"en": "Invalid parameters: %v",
		"zh": "参数错误: %v",
	},
	msgUserLookupFailed: {
		"en": "Failed to locate user %s: %v",
		"zh": "无法定位用户 %s: %v",
	},
	msgFallbackUser: {
		"en": "Fallback to user %s for file access",
		"zh": "fallback to user %s for file access",
	},
	msgDropPrivilegesFailed: {
		"en": "Failed to drop privileges to user %s: %v",
		"zh": "无法降级到用户 %s: %v",
	},
	msgResolveSecretFailed: {
		"en": "Failed to resolve secret: %v",
		"zh": "无法解析 secret: %v",
	},
	msgUserNoSecretNullOK: {
		"en": "User %s has no secret configured; nullok honored",
		"zh": "用户 %s 未配置密钥，nullok 生效",
	},
	msgReadConfigFailed: {
		"en": "Failed to read %s: %v",
		"zh": "读取 %s 失败: %v",
	},
	msgPromptTemplateFailed: {
		"en": "Failed to load prompt template: %v",
		"zh": "加载 prompt 模板失败: %v",
	},
	msgGraceSkip: {
		"en": "Host %s is within grace period, skip verification",
		"zh": "主机 %s 在宽限期内，跳过验证码",
	},
	msgUserAuthFailed: {
		"en": "User %s failed verification: %v",
		"zh": "用户 %s 验证失败: %v",
	},
	msgAuthFailedGeneric: {
		"en": "Verification failed: %v",
		"zh": "验证失败: %v",
	},
	msgInternalError: {
		"en": "Internal error",
		"zh": "内部错误",
	},
	msgUpdateAuthtokFailed: {
		"en": "Failed to update PAM_AUTHTOK",
		"zh": "无法更新 PAM_AUTHTOK",
	},
	msgUserAuthSuccess: {
		"en": "User %s authenticated (%s)",
		"zh": "用户 %s 验证成功 (%s)",
	},
	msgEmptyUsername: {
		"en": "username is empty",
		"zh": "用户名为空",
	},
	msgSerializeConfigFailed: {
		"en": "Failed to serialize config: %v",
		"zh": "序列化配置失败: %v",
	},
	msgSecretChangedDuringProcess: {
		"en": "Secret file changed during processing, please retry",
		"zh": "密钥文件在处理期间发生变化，请重试",
	},
	msgSecretChangedRetry: {
		"en": "Secret file changed, please retry",
		"zh": "密钥文件发生变化，请重试",
	},
	msgReadonlyWriteIgnored: {
		"en": "Readonly mode; ignoring write failure: %v",
		"zh": "只读模式，忽略写入失败: %v",
	},
	msgWriteConfigFailed: {
		"en": "Failed to write %s: %v",
		"zh": "写入 %s 失败: %v",
	},
	msgUpdateConfigFailed: {
		"en": "Failed to update Google Authenticator config",
		"zh": "更新 Google Authenticator 配置失败",
	},
	msgPromptTooLarge: {
		"en": "Prompt template exceeds %d bytes",
		"zh": "prompt 模板超过 %d 字节",
	},
	msgDummyPassword: {
		"en": "Dummy password supplied by PAM. Did OpenSSH 'PermitRootLogin <anything but yes>' or some other config block this login?",
		"zh": "Dummy password supplied by PAM. Did OpenSSH 'PermitRootLogin <anything but yes>' or some other config block this login?",
	},
}

func msgf(key string, args ...interface{}) string {
	format := resolveTranslation(key)
	if len(args) == 0 {
		return format
	}
	return fmt.Sprintf(format, args...)
}

func resolveTranslation(key string) string {
	lang := detectLang()
	if text := translations[key][lang]; text != "" {
		return text
	}
	if text := translations[key]["en"]; text != "" {
		return text
	}
	return key
}

func detectLang() string {
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
