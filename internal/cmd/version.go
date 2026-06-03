package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// 以下变量由 GoReleaser -ldflags 注入；本地 go build 默认为 devel。
var (
	version = "devel"
	commit  = ""
	date    = ""
)

var versionCmd = &cobra.Command{
	Use:     "version",
	Aliases: []string{"v"},
	Short:   "Print version information",
	Long:    "Print the fuckssh version, commit hash, and build date.",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(formatVersion())
	},
}

func init() {
	rootCmd.Version = formatVersion()
	rootCmd.AddCommand(versionCmd)
}

// formatVersion 供 cobra --version 和 version 子命令使用。
func formatVersion() string {
	if commit == "" {
		return version
	}
	if date == "" {
		return fmt.Sprintf("%s (%s)", version, commit)
	}
	return fmt.Sprintf("%s (%s, %s)", version, commit, date)
}
