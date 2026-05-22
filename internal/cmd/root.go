package cmd

import (
	"sync"

	"github.com/fuckssh/fuckssh/internal/i18n"
	"github.com/fuckssh/fuckssh/internal/platform"
	"github.com/spf13/cobra"
)

var helpLocalizedOnce sync.Once

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
	rootCmd.PersistentPreRunE = rootPersistentPreRun
	rootCmd.PersistentFlags().StringVar(&configFileFlag, "config", "", "path to ssh config file (default: ~/.ssh/config)")

	rootCmd.AddCommand(addCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(searchCmd)
}

func rootPersistentPreRun(cmd *cobra.Command, args []string) error {
	if cmd.Flags().Changed("help") {
		_, _ = i18n.Load()
		applyLocalizedHelp()
		return nil
	}
	var err error
	if isReadonlyCmd(cmd) {
		err = i18n.EnsureLoaded()
	} else {
		err = i18n.EnsureInteractive(cmd.ErrOrStderr())
	}
	if err != nil {
		return err
	}
	applyLocalizedHelpOnce()
	return nil
}

// isReadonlyCmd 判断是否为只读子命令（跳过交互式语言选择以加快响应）。
func isReadonlyCmd(cmd *cobra.Command) bool {
	switch cmd.Name() {
	case "list", "search":
		return true
	default:
		return false
	}
}

func applyLocalizedHelpOnce() {
	helpLocalizedOnce.Do(applyLocalizedHelp)
}
