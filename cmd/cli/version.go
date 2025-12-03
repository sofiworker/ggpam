package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"ggpam/pkg/version"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "显示版本信息",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("版本: %s\n", version.Version)
		fmt.Printf("Git 提交: %s\n", version.GitCommit)
		fmt.Printf("构建时间: %s\n", version.BuildDate)
		fmt.Printf("Go 版本: %s\n", version.GoVersion)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
