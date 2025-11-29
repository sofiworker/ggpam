package cmd

import (
	"errors"
	"fmt"
	"os"
	"os/user"
	"strings"
	"time"

	qrcode "github.com/skip2/go-qrcode"
	"github.com/spf13/cobra"

	"gpam/pkg/config"
	"gpam/pkg/otp"
)

const (
	defaultScratchCodes = 5
	maxScratchCodes     = 10

	msgDisallowReuse = "Do you want to disallow multiple uses of the same authentication\ntoken? This restricts you to one login about every 30s, but it increases\nyour chances to notice or even prevent man-in-the-middle attacks"
	msgTotpWindow    = "By default, a new token is generated every 30 seconds by the mobile app.\nIn order to compensate for possible time-skew between the client and the server,\nwe allow an extra token before and after the current time. This allows for a\ntime skew of up to 30 seconds between authentication server and client. If you\nexperience problems with poor time synchronization, you can increase the window\nfrom its default size of 3 permitted codes (one previous code, the current code,\nthe next code) to 17 permitted codes (the 8 previous codes, the current code,\nand the 8 next codes). This will permit for a time skew of up to 4 minutes between\nclient and server.\nDo you want to do so?"
	msgHotpWindow    = "By default, three tokens are valid at any one time. This accounts for\ngenerated-but-not-used tokens and failed login attempts. In order to decrease the\nlikelihood of synchronization problems, this window can be increased from its\ndefault size of 3 to 17. Do you want to do so?"
	msgRateLimit     = "If the computer that you are logging into isn't hardened against brute-force\nlogin attempts, you can enable rate-limiting for the authentication module.\nBy default, this limits attackers to no more than 3 login attempts every 30s.\nDo you want to enable rate-limiting?"
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
	Short: "初始化 ~/.google_authenticator 配置",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runInit(cmd, initOpts)
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().StringVar(&initOpts.path, "path", defaultSecretPath(), "配置文件路径")
	initCmd.Flags().StringVarP(&initOpts.secretFile, "secret", "s", "", "指定密钥文件位置")
	initCmd.Flags().BoolVarP(&initOpts.force, "force", "f", false, "写文件前不再提示确认")
	initCmd.Flags().StringVar(&initOpts.mode, "mode", "", "认证模式: totp 或 hotp (默认交互选择)")
	initCmd.Flags().BoolVarP(&initOpts.timeBased, "time-based", "t", false, "设置为时间同步模式 (TOTP)")
	initCmd.Flags().BoolVarP(&initOpts.counterBased, "counter-based", "c", false, "设置为计数器模式 (HOTP)")
	initCmd.Flags().IntVarP(&initOpts.step, "step-size", "S", config.DefaultStepSize, "TOTP 步长（秒），范围 1..60")
	initCmd.Flags().IntVarP(&initOpts.windowSize, "window-size", "w", 0, "窗口大小 (1..21)，不指定则交互提示")
	initCmd.Flags().BoolVarP(&initOpts.minimalWindow, "minimal-window", "W", false, "使用最小窗口 (TOTP=3, HOTP=1)")
	initCmd.Flags().IntVarP(&initOpts.rateAttempts, "rate-limit", "r", 3, "每个窗口允许的登录次数 (1..10)")
	initCmd.Flags().DurationVarP(&initOpts.rateInterval, "rate-time", "R", 30*time.Second, "速率限制窗口长度 (15s..600s)")
	initCmd.Flags().BoolVarP(&initOpts.disableRate, "no-rate-limit", "u", false, "禁用速率限制")
	initCmd.Flags().IntVarP(&initOpts.scratch, "emergency-codes", "e", defaultScratchCodes, "应急码数量 (0..10)")
	initCmd.Flags().IntVar(&initOpts.scratch, "scratch-codes", defaultScratchCodes, "应急码数量 (兼容旧参数)")
	initCmd.Flags().BoolVarP(&initOpts.disallow, "disallow-reuse", "d", false, "禁止重复使用 TOTP")
	initCmd.Flags().BoolVarP(&initOpts.allowReuse, "allow-reuse", "D", false, "允许重复使用 TOTP")
	initCmd.Flags().StringVarP(&initOpts.label, "label", "l", defaultLabel(), "otpauth URL 的 label")
	initCmd.Flags().StringVarP(&initOpts.issuer, "issuer", "i", "", "otpauth URL 的 issuer")
	initCmd.Flags().BoolVarP(&initOpts.quiet, "quiet", "q", false, "静默模式，仅输出必要信息")
	initCmd.Flags().StringVarP(&initOpts.qrMode, "qr-mode", "Q", "ansi", "二维码输出模式: none/ansi/ansi-inverse/ansi-grey/utf8/utf8-inverse/utf8-grey")
	initCmd.Flags().BoolVar(&initOpts.qrInverse, "qr-inverse", false, "二维码反色显示 (兼容参数)")
	initCmd.Flags().BoolVar(&initOpts.qrUTF8, "qr-utf8", false, "二维码使用 UTF8 渲染 (兼容参数)")
	initCmd.Flags().BoolVar(&initOpts.confirm, "confirm", true, "生成后要求输入验证码确认")
	initCmd.Flags().BoolVarP(&initOpts.noConfirm, "no-confirm", "C", false, "不要求验证码确认 (适合非交互环境)")
}

func runInit(cmd *cobra.Command, opts initOptions) error {
	if opts.allowReuse && opts.disallow {
		return errors.New("allow-reuse 与 disallow-reuse 互斥")
	}
	if opts.counterBased && opts.timeBased {
		return errors.New("counter-based 与 time-based 互斥")
	}
	secretPath := opts.path
	if opts.secretFile != "" {
		secretPath = opts.secretFile
	}
	if secretPath == "" {
		secretPath = defaultSecretPath()
	}
	path, err := expandPath(secretPath)
	if err != nil {
		return err
	}
	if fileExists(path) && !opts.force {
		writeInfo("警告: %s 已存在，如果继续将覆盖旧配置。", path)
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
		return errors.New("step-size 需在 1..60 范围内")
	}
	if window < 1 || window > 21 {
		return errors.New("window-size 需在 1..21 范围内")
	}
	if opts.scratch < 0 || opts.scratch > maxScratchCodes {
		return fmt.Errorf("emergency-codes 需在 0..%d 范围内", maxScratchCodes)
	}

	secret, err := randomSecret(20)
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
		msg := fmt.Sprintf("Do you want me to update your \"%s\" file?", path)
		if !promptYesNo(msg) {
			writeInfo("已取消，不会更新 %s", path)
			return nil
		}
	}

	if err := ensureParent(path); err != nil {
		return err
	}
	if err := cfg.Save(path, 0o600); err != nil {
		return err
	}
	if !opts.quiet {
		writeInfo("已写入配置 %s", path)
	}
	return nil
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
		return promptYesNo("Do you want authentication tokens to be time-based"), nil
	default:
		return false, fmt.Errorf("未知 mode: %s", opts.mode)
	}
}

func determineReuse(opts initOptions, useTOTP bool) (bool, error) {
	if !useTOTP {
		if opts.disallow || opts.allowReuse {
			return false, errors.New("HOTP 模式下不支持 -d/-D")
		}
		return false, nil
	}
	if opts.disallow {
		return true, nil
	}
	if opts.allowReuse {
		return false, nil
	}
	return promptYesNo(msgDisallowReuse), nil
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
		if promptYesNo(msgTotpWindow) {
			return 17, nil
		}
		return config.DefaultWindow, nil
	}
	if promptYesNo(msgHotpWindow) {
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
		return nil, errors.New("必须同时设置 --rate-limit 与 --rate-time")
	}
	if opts.rateAttempts < 0 || opts.rateAttempts > 10 {
		return nil, errors.New("rate-limit 需在 1..10 范围内")
	}
	secs := int(opts.rateInterval / time.Second)
	if secs < 0 {
		return nil, errors.New("rate-time 必须为正数")
	}
	if attChanged && intChanged {
		if secs < 15 || secs > 600 {
			return nil, errors.New("rate-time 需在 15..600 秒")
		}
		if opts.rateAttempts < 1 {
			return nil, errors.New("rate-limit 需在 1..10")
		}
		return &config.RateLimit{Attempts: opts.rateAttempts, Interval: opts.rateInterval}, nil
	}
	if promptYesNo(msgRateLimit) {
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
	builder := newOtpauthBuilder(label, issuer, params, cfg.Mode())
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
	for {
		fmt.Print("Enter code from app (-1 to skip): ")
		line, err := stdinReader.ReadString('\n')
		if err != nil {
			return err
		}
		code := strings.TrimSpace(line)
		if code == "-1" {
			fmt.Println("Code confirmation skipped")
			return nil
		}
		if len(code) == 0 {
			continue
		}
		secret, err := cfg.SecretBytes()
		if err != nil {
			return err
		}
		tm := time.Now().Unix() / int64(cfg.Step())
		correct := otp.Compute(secret, uint64(tm))
		if code == fmt.Sprintf("%06d", correct) {
			fmt.Println("Code confirmed")
			return nil
		}
		fmt.Printf("Code incorrect (correct code %06d). Try again.\n", correct)
	}
}

func printSetupInfo(cfg *config.Config, url string, opts initOptions) {
	fmt.Println("将以下信息添加到身份验证器应用：")
	fmt.Printf("otpauth URL: %s\n", url)
	if opts.qrMode != "none" {
		renderQRCode(url, opts)
	}
	fmt.Printf("Your new secret key is: %s\n", cfg.Secret)
	if cfg.Mode() == config.ModeTOTP {
		fmt.Println("This secret is time-based and will generate a new code every 30 seconds.")
	} else {
		fmt.Println("This secret is counter-based. Each code can only be used once.")
	}
	fmt.Println("If you are using a mobile client, scan the QR code above or enter the secret manually.")
	if len(cfg.ScratchCodes) > 0 {
		fmt.Println("应急码：")
		for _, sc := range cfg.ScratchCodes {
			fmt.Printf("  %08d\n", sc)
		}
	}
}

func renderQRCode(data string, opts initOptions) {
	qr, err := qrcode.New(data, qrcode.Medium)
	if err != nil {
		fmt.Printf("生成二维码失败: %v\n", err)
		return
	}
	mode := strings.ToLower(opts.qrMode)
	inverse := opts.qrInverse || strings.Contains(mode, "inverse")
	useUTF8 := opts.qrUTF8 || strings.Contains(mode, "utf8")
	if useUTF8 {
		fmt.Println(qrcodeToUTF8(qr.Bitmap(), inverse))
		return
	}
	fmt.Println(qr.ToSmallString(inverse))
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
