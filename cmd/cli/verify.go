package main

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"ggpam/pkg/authenticator"
	"ggpam/pkg/config"
	"ggpam/pkg/i18n"
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
	Short: i18n.Resolve(i18n.MsgCmdVerifyShort),
	RunE: func(cmd *cobra.Command, args []string) error {
		if verifyOpts.code == "" && len(args) > 0 {
			verifyOpts.code = args[0]
		}
		return runVerify(verifyOpts)
	},
}

func init() {
	rootCmd.AddCommand(verifyCmd)
	verifyCmd.Flags().StringVar(&verifyOpts.path, "path", defaultSecretPath(), i18n.Resolve(i18n.MsgCliFlagPath))
	verifyCmd.Flags().StringVar(&verifyOpts.code, "code", "", i18n.Resolve(i18n.MsgCliFlagVerifyCode))
	verifyCmd.Flags().BoolVar(&verifyOpts.noSkew, "no-skew-adjust", false, i18n.Resolve(i18n.MsgCliFlagNoSkew))
	verifyCmd.Flags().BoolVar(&verifyOpts.noIncrement, "no-increment-hotp", false, i18n.Resolve(i18n.MsgCliFlagNoIncrementHOTP))
	verifyCmd.Flags().BoolVar(&verifyOpts.quiet, "quiet", false, i18n.Resolve(i18n.MsgCliFlagVerifyQuiet))
}

func runVerify(opts verifyOptions) error {
	if opts.code == "" {
		return errors.New(i18n.Resolve(i18n.MsgCliVerifyNeedCode))
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
			return fmt.Errorf("%s", i18n.Resolve(i18n.MsgCliVerifyRateLimited))
		}
		return err
	}
	if cfg.Dirty {
		if err := cfg.Save(path, DefaultSecretFilePerm); err != nil {
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
		writeInfo("%s", i18n.Resolve(i18n.MsgCliVerifyScratchUsed))
	case authenticator.ResultHOTP:
		writeInfo(i18n.Resolve(i18n.MsgCliVerifyHOTPSuccess), res.Counter)
	default:
		writeInfo("%s", i18n.Resolve(i18n.MsgCliVerifyTOTPSuccess))
	}
}

func (v verifyResponder) OnError(error) {}
