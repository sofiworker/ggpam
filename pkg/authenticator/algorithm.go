package authenticator

import (
	"time"

	"gpam/pkg/config"
	"gpam/pkg/otp"
)

type Algorithm interface {
	Verify(cfg *config.Config, secret []byte, code int, opts VerifyOptions, now time.Time) (Result, error)
}

type totpAlgorithm struct {
	owner *Authenticator
}

func (t *totpAlgorithm) Verify(cfg *config.Config, secret []byte, code int, opts VerifyOptions, now time.Time) (Result, error) {
	step := cfg.Step()
	window := cfg.Window()
	tm := now.Unix() / int64(step)
	targetSkew := cfg.Options.TimeSkew
	if cfg.Options.DisallowReuse {
		if err := cfg.ResetDisallowList(tm+int64(targetSkew), window); err != nil {
			return Result{}, err
		}
	}
	for offset := -(window - 1) / 2; offset <= window/2; offset++ {
		counter := tm + int64(targetSkew) + int64(offset)
		if counter < 0 {
			continue
		}
		if otp.Compute(secret, uint64(counter)) == code {
			if err := cfg.CheckReuse(counter); err != nil {
				return Result{}, err
			}
			cfg.RecordUsedTimestamp(counter)
			return Result{
				Type:      ResultTOTP,
				Timestamp: counter,
			}, nil
		}
	}
	if opts.DisableSkewAdjustment {
		return Result{}, ErrInvalidCode
	}
	if skew, found := t.detectSkew(secret, tm, code); found {
		if cfg.RecordSkewObservation(tm, skew) {
			return Result{
				Type:          ResultTOTP,
				Timestamp:     tm + int64(skew),
				ConfigChanged: true,
			}, nil
		}
	}
	return Result{}, ErrInvalidCode
}

func (t *totpAlgorithm) detectSkew(secret []byte, tm int64, code int) (int, bool) {
	if t.owner != nil {
		return t.owner.detectSkew(secret, tm, code)
	}
	return detectSkew(secret, tm, code)
}

type hotpAlgorithm struct{}

func (h *hotpAlgorithm) Verify(cfg *config.Config, secret []byte, code int, opts VerifyOptions, _ time.Time) (Result, error) {
	counter := cfg.Options.HOTPCounter
	window := cfg.Window()
	for i := 0; i < window; i++ {
		value := counter + int64(i)
		if value < 0 {
			continue
		}
		if otp.Compute(secret, uint64(value)) == code {
			cfg.Options.HOTPCounter = value + 1
			cfg.MarkDirty()
			return Result{
				Type:    ResultHOTP,
				Counter: value,
			}, nil
		}
	}
	if !opts.NoIncrementHOTP {
		cfg.Options.HOTPCounter = counter + 1
		cfg.MarkDirty()
	}
	return Result{}, ErrInvalidCode
}

func (a *Authenticator) detectSkew(secret []byte, tm int64, code int) (int, bool) {
	return detectSkew(secret, tm, code)
}

func detectSkew(secret []byte, tm int64, code int) (int, bool) {
	const maxIterations = 25 * 60
	for i := 1; i < maxIterations; i++ {
		if tm-int64(i) >= 0 {
			if otp.Compute(secret, uint64(tm-int64(i))) == code {
				return -i, true
			}
		}
		if otp.Compute(secret, uint64(tm+int64(i))) == code {
			return i, true
		}
	}
	return 0, false
}
