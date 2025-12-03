package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"ggpam/pkg/i18n"
	"ggpam/pkg/version"
)

var rootCmd = &cobra.Command{
	Use: "ggpam",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println(cmd.UsageString())
		return nil
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "exec cmd failed: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.Short = i18n.Resolve(i18n.MsgCliShort)
	rootCmd.Long = i18n.Resolve(i18n.MsgCliLong)
	rootCmd.Version = version.Version
	rootCmd.SetVersionTemplate("ggpam {{.Version}}\n")
	rootCmd.PersistentFlags().BoolP("help", "h", false, i18n.Resolve(i18n.MsgCliFlagHelp))
	rootCmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		fmt.Println(buildUsage())
	})
}

func buildUsage() string {
	return fmt.Sprintf(i18n.Resolve(i18n.MsgCliUsage), version.Version)
}
