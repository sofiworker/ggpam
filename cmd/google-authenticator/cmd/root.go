package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"gpam/pkg/version"
)

var rootCmd = &cobra.Command{
	Use:   "google-authenticator",
	Short: "Google Authenticator (Go 版本)",
	Long:  "Google Authenticator CLI，提供配置初始化、验证码验证等功能。",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println(cmd.UsageString())
		return nil
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "执行失败: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.Version = version.Version
	rootCmd.SetVersionTemplate("google-authenticator {{.Version}}\n")
	rootCmd.PersistentFlags().BoolP("help", "h", false, "显示帮助")
	rootCmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		fmt.Println(buildUsage())
	})
}

func buildUsage() string {
	return fmt.Sprintf(`google-authenticator %s
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
  -e, --emergency-codes=N           Number of emergency codes to generate
`, version.Version)
}
