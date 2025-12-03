package main

import (
	"bufio"
	"errors"
	"fmt"
	"ggpam/pkg/util"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"time"

	"github.com/skip2/go-qrcode"
	"github.com/spf13/cobra"

	"ggpam/pkg/config"
	"ggpam/pkg/i18n"
	"ggpam/pkg/otp"
)

const (
	defaultScratchCodes   = 5
	maxScratchCodes       = 10
	DefaultSecretFilename = ".ggpam_authenticator"
	DefaultSecretFilePerm = 0o600
	DefaultSecretDirPerm  = 0o700
	EnvSecretPath         = "GPAM_SECRET_PATH"
)

type initOptions struct {
	path          string
	secretFile    string
	force         bool
	mode          string
	timeBased     bool
	counterBased  bool
	step          int
	windowSize    int
	minimalWindow bool
	rateAttempts  int
	rateInterval  time.Duration
	disableRate   bool
	scratch       int
	disallow      bool
	allowReuse    bool
	label         string
	issuer        string
	quiet         bool
	qrMode        string
	qrInverse     bool
	qrUTF8        bool
	confirm       bool
	noConfirm     bool
}

var initOpts = initOptions{
	step:         config.DefaultStepSize,
	windowSize:   0,
	rateAttempts: 3,
	rateInterval: 30 * time.Second,
	scratch:      defaultScratchCodes,
	confirm:      true,
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: i18n.Resolve(i18n.MsgCmdInitShort),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runInit(cmd, initOpts)
	},
}

func init() {
	rootCmd.AddCommand(initCmd)

	initCmd.Flags().StringVar(&initOpts.path, "path", defaultSecretPath(), i18n.Resolve(i18n.MsgCliFlagPath))
	initCmd.Flags().StringVarP(&initOpts.secretFile, "secret", "s", "", i18n.Resolve(i18n.MsgCliFlagSecret))
	initCmd.Flags().BoolVarP(&initOpts.force, "force", "f", false, i18n.Resolve(i18n.MsgCliFlagForce))
	initCmd.Flags().StringVar(&initOpts.mode, "mode", "", i18n.Resolve(i18n.MsgCliFlagMode))
	initCmd.Flags().BoolVarP(&initOpts.timeBased, "time-based", "t", false, i18n.Resolve(i18n.MsgCliFlagTimeBased))
	initCmd.Flags().BoolVarP(&initOpts.counterBased, "counter-based", "c", false, i18n.Resolve(i18n.MsgCliFlagCounterBased))
	initCmd.Flags().IntVarP(&initOpts.step, "step-size", "S", config.DefaultStepSize, i18n.Resolve(i18n.MsgCliFlagStepSize))
	initCmd.Flags().IntVarP(&initOpts.windowSize, "window-size", "w", 0, i18n.Resolve(i18n.MsgCliFlagWindowSize))
	initCmd.Flags().BoolVarP(&initOpts.minimalWindow, "minimal-window", "W", false, i18n.Resolve(i18n.MsgCliFlagMinimalWindow))
	initCmd.Flags().IntVarP(&initOpts.rateAttempts, "rate-limit", "r", 3, i18n.Resolve(i18n.MsgCliFlagRateLimit))
	initCmd.Flags().DurationVarP(&initOpts.rateInterval, "rate-time", "R", 30*time.Second, i18n.Resolve(i18n.MsgCliFlagRateTime))
	initCmd.Flags().BoolVarP(&initOpts.disableRate, "no-rate-limit", "u", false, i18n.Resolve(i18n.MsgCliFlagDisableRate))
	initCmd.Flags().IntVarP(&initOpts.scratch, "emergency-codes", "e", defaultScratchCodes, i18n.Resolve(i18n.MsgCliFlagEmergencyCodes))
	initCmd.Flags().IntVar(&initOpts.scratch, "scratch-codes", defaultScratchCodes, i18n.Resolve(i18n.MsgCliFlagScratchCodes))
	initCmd.Flags().BoolVarP(&initOpts.disallow, "disallow-reuse", "d", false, i18n.Resolve(i18n.MsgCliFlagDisallowReuse))
	initCmd.Flags().BoolVarP(&initOpts.allowReuse, "allow-reuse", "D", false, i18n.Resolve(i18n.MsgCliFlagAllowReuse))
	initCmd.Flags().StringVarP(&initOpts.label, "label", "l", defaultLabel(), i18n.Resolve(i18n.MsgCliFlagLabel))
	initCmd.Flags().StringVarP(&initOpts.issuer, "issuer", "i", "", i18n.Resolve(i18n.MsgCliFlagIssuer))
	initCmd.Flags().BoolVarP(&initOpts.quiet, "quiet", "q", false, i18n.Resolve(i18n.MsgCliFlagQuiet))
	initCmd.Flags().StringVarP(&initOpts.qrMode, "qr-mode", "Q", "ansi", i18n.Resolve(i18n.MsgCliFlagQRMode))
	initCmd.Flags().BoolVar(&initOpts.qrInverse, "qr-inverse", false, i18n.Resolve(i18n.MsgCliFlagQRInverse))
	initCmd.Flags().BoolVar(&initOpts.qrUTF8, "qr-utf8", false, i18n.Resolve(i18n.MsgCliFlagQRUTF8))
	initCmd.Flags().BoolVar(&initOpts.confirm, "confirm", true, i18n.Resolve(i18n.MsgCliFlagConfirm))
	initCmd.Flags().BoolVarP(&initOpts.noConfirm, "no-confirm", "C", false, i18n.Resolve(i18n.MsgCliFlagNoConfirm))
}

func runInit(cmd *cobra.Command, opts initOptions) error {
	if opts.allowReuse && opts.disallow {
		return errors.New(msg(i18n.MsgCliAllowDisallowConflict))
	}
	if opts.counterBased && opts.timeBased {
		return errors.New(msg(i18n.MsgCliCounterTimeConflict))
	}
	secretPath := opts.path
	if opts.secretFile != "" {
		secretPath = opts.secretFile
	}
	if secretPath == "" {
		secretPath = defaultSecretPath()
	}
	path, err := util.ExpandPath(secretPath)
	if err != nil {
		return err
	}
	if util.FileExists(path) && !opts.force {
		fmt.Printf(msg(i18n.MsgCliFileExistsWarn, path))
	}

	useTOTP, err := determineMode(opts)
	if err != nil {
		return err
	}
	reqDisallow, err := determineReuse(opts, useTOTP)
	if err != nil {
		return err
	}
	window, err := determineWindow(opts, useTOTP)
	if err != nil {
		return err
	}
	rateLimit, err := determineRateLimit(cmd, opts)
	if err != nil {
		return err
	}
	if opts.step < 1 || opts.step > 60 {
		return errors.New(msg(i18n.MsgCliStepRange))
	}
	if window < 1 || window > 21 {
		return errors.New(msg(i18n.MsgCliWindowRange))
	}
	if opts.scratch < 0 || opts.scratch > maxScratchCodes {
		return fmt.Errorf("%s", msg(i18n.MsgCliScratchRange, maxScratchCodes))
	}

	secret, err := util.RandomSecret(20)
	if err != nil {
		return err
	}
	scratchCodes, err := otp.GenerateScratchCodesDefault(opts.scratch)
	if err != nil {
		return err
	}

	cfg := &config.Config{
		Secret:       secret,
		ScratchCodes: scratchCodes,
		Options: config.Options{
			StepSize:   opts.step,
			WindowSize: window,
			Additional: map[string]string{},
		},
	}
	if useTOTP {
		cfg.Options.TOTPAuth = true
	} else {
		cfg.Options.HOTPConfigured = true
		cfg.Options.HOTPCounter = 1
	}
	cfg.Options.DisallowReuse = reqDisallow
	if !reqDisallow {
		cfg.Options.DisallowedTimestamps = nil
	}
	cfg.Options.RateLimit = rateLimit

	if !opts.quiet {
		url := buildOtpauthURL(cfg, opts)
		printSetupInfo(cfg, url, opts)
		if !opts.noConfirm && opts.confirm && cfg.Mode() == config.ModeTOTP {
			if err := confirmCode(cfg); err != nil {
				return err
			}
		}
	}

	if !opts.force {
		prompt := msg(i18n.MsgCliUpdateFilePrompt, path)
		if !util.PromptYesNo(prompt) {
			fmt.Printf(msg(i18n.MsgCliConfigCancelled, path))
			return nil
		}
	}

	if err := util.MkDirWithPerm(path, DefaultSecretDirPerm); err != nil {
		return err
	}
	if err := cfg.Save(path, DefaultSecretFilePerm); err != nil {
		return err
	}
	if !opts.quiet {
		fmt.Println(i18n.Msgf(i18n.MsgCliConfigWritten, path))
	}
	return nil
}

func defaultSecretPath() string {
	if p := os.Getenv(EnvSecretPath); strings.TrimSpace(p) != "" {
		return p
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join("~", DefaultSecretFilename)
	}
	return filepath.Join(home, DefaultSecretFilename)
}

func determineMode(opts initOptions) (bool, error) {
	mode := strings.ToLower(opts.mode)
	if opts.timeBased {
		mode = "totp"
	}
	if opts.counterBased {
		mode = "hotp"
	}
	switch mode {
	case "totp", "time", "time-based":
		return true, nil
	case "hotp", "counter", "counter-based":
		return false, nil
	case "":
		return util.PromptYesNo(msg(i18n.MsgCliPromptTimeBased)), nil
	default:
		return false, fmt.Errorf("%s", msg(i18n.MsgCliUnknownMode, opts.mode))
	}
}

func determineReuse(opts initOptions, useTOTP bool) (bool, error) {
	if !useTOTP {
		if opts.disallow || opts.allowReuse {
			return false, errors.New(msg(i18n.MsgCliHotpNoReuse))
		}
		return false, nil
	}
	if opts.disallow {
		return true, nil
	}
	if opts.allowReuse {
		return false, nil
	}
	return util.PromptYesNo(msg(i18n.MsgCliDisallowReusePrompt)), nil
}

func determineWindow(opts initOptions, useTOTP bool) (int, error) {
	if opts.minimalWindow {
		if useTOTP {
			return max(3, opts.windowSize), nil
		}
		return 1, nil
	}
	if opts.windowSize > 0 {
		return opts.windowSize, nil
	}
	if useTOTP {
		if util.PromptYesNo(msg(i18n.MsgCliTotpWindowPrompt)) {
			return 17, nil
		}
		return config.DefaultWindow, nil
	}
	if util.PromptYesNo(msg(i18n.MsgCliHotpWindowPrompt)) {
		return 17, nil
	}
	return config.DefaultWindow, nil
}

func determineRateLimit(cmd *cobra.Command, opts initOptions) (*config.RateLimit, error) {
	attChanged := cmd.Flags().Changed("rate-limit")
	intChanged := cmd.Flags().Changed("rate-time")
	if opts.disableRate {
		return nil, nil
	}
	if attChanged != intChanged {
		return nil, errors.New(msg(i18n.MsgCliRateArgsMismatch))
	}
	if opts.rateAttempts < 0 || opts.rateAttempts > 10 {
		return nil, errors.New(msg(i18n.MsgCliRateLimitRange))
	}
	secs := int(opts.rateInterval / time.Second)
	if secs < 0 {
		return nil, errors.New(msg(i18n.MsgCliRateTimePositive))
	}
	if attChanged && intChanged {
		if secs < 15 || secs > 600 {
			return nil, errors.New(msg(i18n.MsgCliRateTimeRange))
		}
		if opts.rateAttempts < 1 {
			return nil, errors.New(msg(i18n.MsgCliRateLimitRange))
		}
		return &config.RateLimit{Attempts: opts.rateAttempts, Interval: opts.rateInterval}, nil
	}
	if util.PromptYesNo(msg(i18n.MsgCliRateLimitPrompt)) {
		return &config.RateLimit{Attempts: 3, Interval: 30 * time.Second}, nil
	}
	return nil, nil
}

func buildOtpauthURL(cfg *config.Config, opts initOptions) string {
	label := opts.label
	issuer := opts.issuer
	if issuer == "" {
		issuer = label
	}
	params := map[string]string{
		"secret":    cfg.Secret,
		"issuer":    issuer,
		"digits":    "6",
		"algorithm": "SHA1",
	}
	switch cfg.Mode() {
	case config.ModeHOTP:
		params["counter"] = fmt.Sprintf("%d", cfg.Options.HOTPCounter)
	default:
		params["period"] = fmt.Sprintf("%d", cfg.Step())
	}
	builder := otp.NewOTPAuthBuilder(label, issuer, params, cfg.Mode())
	return builder.String()
}

func defaultLabel() string {
	current, err := user.Current()
	name := "user"
	if err == nil && current.Username != "" {
		name = current.Username
	}
	host, err := os.Hostname()
	if err != nil || host == "" {
		host = "unix"
	}
	return fmt.Sprintf("%s@%s", name, host)
}

func confirmCode(cfg *config.Config) error {
	var stdinReader = bufio.NewReader(os.Stdin)
	for {
		fmt.Print(msg(i18n.MsgCliEnterCode))
		line, err := stdinReader.ReadString('\n')
		if err != nil {
			return err
		}
		code := strings.TrimSpace(line)
		if code == "-1" {
			fmt.Println(msg(i18n.MsgCliCodeSkipped))
			return nil
		}
		if len(code) == 0 {
			continue
		}
		ok, hint, err := validateTOTPInput(cfg, code)
		if err != nil {
			return err
		}
		if ok {
			fmt.Println(msg(i18n.MsgCliCodeConfirmed))
			return nil
		}
		fmt.Println(msg(i18n.MsgCliCodeIncorrect, hint))
	}
}

func validateTOTPInput(cfg *config.Config, code string) (bool, string, error) {
	if len(code) != 6 || strings.IndexFunc(code, func(r rune) bool { return r < '0' || r > '9' }) >= 0 {
		return false, "", fmt.Errorf("%s", msg(i18n.MsgCliCodeInvalidDigits))
	}
	secret, err := cfg.SecretBytes()
	if err != nil {
		return false, "", err
	}
	step := cfg.Step()
	if step < 1 {
		step = config.DefaultStepSize
	}
	window := cfg.Window()
	if window < 1 {
		window = 1
	}
	now := time.Now()
	counter := now.Unix() / int64(step)
	skew := int64(cfg.Options.TimeSkew)
	baseCounter := counter + skew
	if baseCounter < 0 {
		baseCounter = 0
	}
	bestCode := fmt.Sprintf("%06d", otp.Compute(secret, uint64(baseCounter)))
	for offset := -(window - 1) / 2; offset <= window/2; offset++ {
		value := baseCounter + int64(offset)
		if value < 0 {
			continue
		}
		if code == fmt.Sprintf("%06d", otp.Compute(secret, uint64(value))) {
			return true, bestCode, nil
		}
	}
	return false, bestCode, nil
}

func printSetupInfo(cfg *config.Config, url string, opts initOptions) {
	fmt.Println(msg(i18n.MsgCliSetupAddInfo))
	fmt.Printf(msg(i18n.MsgCliSetupURL)+"\n", url)
	if opts.qrMode != "none" {
		renderQRCode(url, opts)
	}
	fmt.Printf(msg(i18n.MsgCliSetupSecret)+"\n", cfg.Secret)
	if cfg.Mode() == config.ModeTOTP {
		fmt.Println(msg(i18n.MsgCliSetupTimeBased))
	} else {
		fmt.Println(msg(i18n.MsgCliSetupCounterBased))
	}
	fmt.Println(msg(i18n.MsgCliSetupManual))
	if len(cfg.ScratchCodes) > 0 {
		fmt.Println(msg(i18n.MsgCliScratchListHeader))
		for _, sc := range cfg.ScratchCodes {
			fmt.Printf("  %08d\n", sc)
		}
	}
}

func renderQRCode(data string, opts initOptions) {
	qr, err := qrcode.New(data, qrcode.Medium)
	if err != nil {
		fmt.Printf(msg(i18n.MsgCliQRFail)+"\n", err)
		return
	}
	mode := strings.ToLower(opts.qrMode)
	inverse := opts.qrInverse || strings.Contains(mode, "inverse")
	useUTF8 := opts.qrUTF8 || strings.Contains(mode, "utf8")
	if useUTF8 {
		fmt.Println(util.QRCodeToUTF8(qr.Bitmap(), inverse))
		return
	}
	fmt.Println(qr.ToSmallString(inverse))
}

func msg(key string, args ...any) string {
	return i18n.Msgf(key, args...)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
