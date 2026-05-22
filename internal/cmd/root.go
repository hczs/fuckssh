package cmd

import (
	"github.com/fuckssh/fuckssh/internal/platform"
	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:   "fuckssh",
		Short: "Manage ~/.ssh/config for VPS hosts",
		Long:  "fuckssh is a cross-platform CLI for SSH config, host listing, and search.",
	}
	// configFileFlag 允许用 --config 覆盖默认的 ~/.ssh/config 路径。
	configFileFlag string
)

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}

// ConfigFilePath 返回当前应读取的 ssh config 路径（优先 --config）。
func ConfigFilePath() (string, error) {
	if configFileFlag != "" {
		return platform.ExpandPath(configFileFlag)
	}
	return platform.DefaultConfigPath()
}

func init() {
	rootCmd.PersistentFlags().StringVar(&configFileFlag, "config", "", "path to ssh config file (default: ~/.ssh/config)")

	rootCmd.AddCommand(addCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(searchCmd)
}
