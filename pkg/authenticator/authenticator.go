package authenticator

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"gpam/pkg/config"
)

type ResultType string

const (
	ResultScratch ResultType = "scratch"
	ResultTOTP    ResultType = "totp"
	ResultHOTP    ResultType = "hotp"
)

var (
	ErrInvalidCode = errors.New("验证码不匹配")
	ErrNoSecret    = errors.New("配置缺少共享密钥")
	ErrModeUnknown = errors.New("配置未启用 HOTP/TOTP")
)

type VerifyOptions struct {
	DisableSkewAdjustment bool
	NoIncrementHOTP       bool
}

type Result struct {
	Type          ResultType
	Counter       int64
	Timestamp     int64
	ConfigChanged bool
}

type ResponseHandler interface {
	OnSuccess(Result)
	OnError(error)
}

type nopResponder struct{}

func (nopResponder) OnSuccess(Result) {}
func (nopResponder) OnError(error)    {}

type Authenticator struct {
	Now        func() time.Time
	Responder  ResponseHandler
	algorithms map[config.Mode]Algorithm
}

func (a *Authenticator) now() time.Time {
	if a != nil && a.Now != nil {
		return a.Now()
	}
	return time.Now()
}

func (a *Authenticator) VerifyCode(cfg *config.Config, raw string, opts VerifyOptions) (Result, error) {
	responder := a.responder()
	if cfg == nil {
		err := errors.New("配置为空")
		responder.OnError(err)
		return Result{}, err
	}
	if strings.TrimSpace(cfg.Secret) == "" {
		responder.OnError(ErrNoSecret)
		return Result{}, ErrNoSecret
	}
	now := a.now()
	if err := cfg.EnforceRateLimit(now); err != nil {
		responder.OnError(err)
		return Result{}, err
	}
	dirtyBefore := cfg.Dirty
	token := strings.TrimSpace(raw)
	if token == "" {
		responder.OnError(ErrInvalidCode)
		return Result{}, ErrInvalidCode
	}
	if len(token) != 6 && len(token) != 8 {
		err := fmt.Errorf("验证码长度必须为 6 或 8 位: %s", token)
		responder.OnError(err)
		return Result{}, err
	}
	if strings.IndexFunc(token, func(r rune) bool { return r < '0' || r > '9' }) >= 0 {
		responder.OnError(ErrInvalidCode)
		return Result{}, ErrInvalidCode
	}
	value, _ := strconv.Atoi(token)
	if len(token) == 8 {
		if cfg.UseScratchCode(value) {
			res := Result{
				Type:          ResultScratch,
				ConfigChanged: cfg.Dirty != dirtyBefore,
			}
			responder.OnSuccess(res)
			return res, nil
		}
		responder.OnError(ErrInvalidCode)
		return Result{}, ErrInvalidCode
	}
	secret, err := cfg.SecretBytes()
	if err != nil {
		responder.OnError(err)
		return Result{}, err
	}
	algo := a.getAlgorithms()[cfg.Mode()]
	if algo == nil {
		responder.OnError(ErrModeUnknown)
		return Result{}, ErrModeUnknown
	}
	res, err := algo.Verify(cfg, secret, value, opts, now)
	if err != nil {
		responder.OnError(err)
		return Result{}, err
	}
	res.ConfigChanged = cfg.Dirty != dirtyBefore || res.ConfigChanged
	responder.OnSuccess(res)
	return res, nil
}

func (a *Authenticator) getAlgorithms() map[config.Mode]Algorithm {
	if a != nil && a.algorithms != nil {
		return a.algorithms
	}
	algoMap := map[config.Mode]Algorithm{
		config.ModeTOTP: &totpAlgorithm{owner: a},
		config.ModeHOTP: &hotpAlgorithm{},
	}
	if a != nil {
		a.algorithms = algoMap
	}
	return algoMap
}

func (a *Authenticator) responder() ResponseHandler {
	if a != nil && a.Responder != nil {
		return a.Responder
	}
	return nopResponder{}
}
