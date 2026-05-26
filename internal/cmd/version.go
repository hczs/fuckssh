package cmd

import "fmt"

// 以下变量由 GoReleaser -ldflags 注入；本地 go build 默认为 devel。
var (
	version = "devel"
	commit  = ""
	date    = ""
)

func init() {
	rootCmd.Version = formatVersion()
}

// formatVersion 供 cobra --version 使用。
func formatVersion() string {
	if commit == "" {
		return version
	}
	if date == "" {
		return fmt.Sprintf("%s (%s)", version, commit)
	}
	return fmt.Sprintf("%s (%s, %s)", version, commit, date)
}
