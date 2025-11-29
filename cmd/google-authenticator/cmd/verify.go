package cmd

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"gpam/pkg/authenticator"
	"gpam/pkg/config"
)

type verifyOptions struct {
	path        string
	code        string
	noSkew      bool
	noIncrement bool
	quiet       bool
}

var verifyOpts verifyOptions

var verifyCmd = &cobra.Command{
	Use:   "verify",
	Short: "验证一次性密码或应急码",
	RunE: func(cmd *cobra.Command, args []string) error {
		if verifyOpts.code == "" && len(args) > 0 {
			verifyOpts.code = args[0]
		}
		return runVerify(verifyOpts)
	},
}

func init() {
	rootCmd.AddCommand(verifyCmd)
	verifyCmd.Flags().StringVar(&verifyOpts.path, "path", defaultSecretPath(), "配置文件路径")
	verifyCmd.Flags().StringVar(&verifyOpts.code, "code", "", "待验证的 6 位验证码或 8 位应急码，也可作为参数提供")
	verifyCmd.Flags().BoolVar(&verifyOpts.noSkew, "no-skew-adjust", false, "禁用自动时间偏移探测")
	verifyCmd.Flags().BoolVar(&verifyOpts.noIncrement, "no-increment-hotp", false, "失败后不推进 HOTP 计数器")
	verifyCmd.Flags().BoolVar(&verifyOpts.quiet, "quiet", false, "不输出成功提示")
}

func runVerify(opts verifyOptions) error {
	if opts.code == "" {
		return errors.New("请通过 --code 或参数提供验证码")
	}
	path, err := expandPath(opts.path)
	if err != nil {
		return err
	}
	cfg, err := config.Load(path)
	if err != nil {
		return err
	}
	auth := &authenticator.Authenticator{
		Responder: verifyResponder{quiet: opts.quiet},
	}
	_, err = auth.VerifyCode(cfg, opts.code, authenticator.VerifyOptions{
		DisableSkewAdjustment: opts.noSkew,
		NoIncrementHOTP:       opts.noIncrement,
	})
	if err != nil {
		if errors.Is(err, config.ErrRateLimited) {
			return fmt.Errorf("当前登录次数过多，请稍后再试")
		}
		return err
	}
	if cfg.Dirty {
		if err := cfg.Save(path, 0o600); err != nil {
			return err
		}
	}
	return nil
}

type verifyResponder struct {
	quiet bool
}

func (v verifyResponder) OnSuccess(res authenticator.Result) {
	if v.quiet {
		return
	}
	switch res.Type {
	case authenticator.ResultScratch:
		writeInfo("已使用应急码，建议尽快补充新的应急码")
	case authenticator.ResultHOTP:
		writeInfo("HOTP 验证成功，当前计数器=%d", res.Counter)
	default:
		writeInfo("TOTP 验证成功")
	}
}

func (v verifyResponder) OnError(error) {}
