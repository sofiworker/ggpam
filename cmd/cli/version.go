package main

import (
	"fmt"
	"ggpam/pkg/i18n"

	"github.com/spf13/cobra"

	"ggpam/pkg/version"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: i18n.Resolve(i18n.MsgShowVersionShort),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf(i18n.Resolve(i18n.MsgVersion)+": %s\n", version.Version)
		fmt.Printf(i18n.Resolve(i18n.MsgGitSha)+": %s\n", version.GitCommit)
		fmt.Printf(i18n.Resolve(i18n.MsgBuildTime)+": %s\n", version.BuildDate)
		fmt.Printf(i18n.Resolve(i18n.MsgGolangVersion)+": %s\n", version.GoVersion)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
