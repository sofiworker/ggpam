package pam

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type PassMode int

const (
	ModePrompt PassMode = iota
	ModeTryFirst
	ModeUseFirst
)

type Params struct {
	SecretSpec      string
	Prompt          string
	PromptOverride  bool
	PromptTemplate  string
	PassMode        PassMode
	ForwardPass     bool
	EchoCode        bool
	NullOK          bool
	Debug           bool
	NoSkewAdjust    bool
	NoIncrementHOTP bool
	AllowReadonly   bool
	NoStrictOwner   bool
	AllowedPerm     os.FileMode
	GracePeriod     time.Duration
	ForcedUser      string
}

func DefaultParams() Params {
	return Params{
		Prompt:      "Verification code: ",
		PassMode:    ModePrompt,
		AllowedPerm: 0o600,
	}
}

func ParseParams(args []string) (Params, error) {
	params := DefaultParams()
	for _, arg := range args {
		switch {
		case strings.HasPrefix(arg, "secret="):
			params.SecretSpec = strings.TrimPrefix(arg, "secret=")
		case strings.HasPrefix(arg, "authtok_prompt="):
			params.Prompt = strings.TrimPrefix(arg, "authtok_prompt=")
			params.PromptOverride = true
		case strings.HasPrefix(arg, "prompt_file=") || strings.HasPrefix(arg, "prompt_template="):
			parts := strings.SplitN(arg, "=", 2)
			if len(parts) != 2 || parts[1] == "" {
				return params, fmt.Errorf("prompt_file requires a path")
			}
			params.PromptTemplate = parts[1]
		case strings.HasPrefix(arg, "user="):
			params.ForcedUser = strings.TrimPrefix(arg, "user=")
		case strings.HasPrefix(arg, "allowed_perm="):
			value := strings.TrimPrefix(arg, "allowed_perm=")
			perm, err := strconv.ParseUint(value, 8, 32)
			if err != nil || perm == 0 {
				return params, fmt.Errorf("invalid allowed_perm %q", value)
			}
			params.AllowedPerm = os.FileMode(perm)
		case strings.HasPrefix(arg, "grace_period="):
			value := strings.TrimPrefix(arg, "grace_period=")
			secs, err := strconv.Atoi(value)
			if err != nil || secs < 0 {
				return params, fmt.Errorf("grace_period must be a non-negative integer seconds: %q", value)
			}
			params.GracePeriod = time.Duration(secs) * time.Second
		case arg == "try_first_pass":
			params.PassMode = ModeTryFirst
		case arg == "use_first_pass":
			params.PassMode = ModeUseFirst
		case arg == "forward_pass":
			params.ForwardPass = true
		case arg == "echo-verification-code" || arg == "echo_verification_code":
			params.EchoCode = true
		case arg == "nullok":
			params.NullOK = true
		case arg == "debug":
			params.Debug = true
		case arg == "noskewadj":
			params.NoSkewAdjust = true
		case arg == "no_increment_hotp":
			params.NoIncrementHOTP = true
		case arg == "no_strict_owner":
			params.NoStrictOwner = true
		case arg == "allow_readonly":
			params.AllowReadonly = true
		default:
			return params, fmt.Errorf("unknown parameter %q", arg)
		}
	}
	if params.ForwardPass && !params.PromptOverride {
		params.Prompt = "Password & verification code: "
	}
	return params, nil
}
