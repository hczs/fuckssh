package cmd

import (
	"fmt"
	"io"
	"time"

	"github.com/fuckssh/fuckssh/internal/config"
	"github.com/fuckssh/fuckssh/internal/i18n"
	"github.com/spf13/cobra"
)

// --- 耗时输出 ---

func printCmdElapsed(w io.Writer, elapsed time.Duration) {
	_, _ = fmt.Fprint(w, i18n.T(i18n.KeyCmdElapsedMs, elapsed.Milliseconds()))
}

func helpInArgs(args []string) bool {
	for _, a := range args {
		switch a {
		case "-h", "--help", "-help":
			return true
		}
	}
	return false
}

// resetHelpFlags 清除命令树上 help 的 Changed 状态，避免测试间残留。
func resetHelpFlags(cmd *cobra.Command) {
	if cmd == nil {
		return
	}
	if f := cmd.Flags().Lookup("help"); f != nil {
		f.Changed = false
		_ = cmd.Flags().Set("help", "false")
	}
	for _, sub := range cmd.Commands() {
		resetHelpFlags(sub)
	}
}

// --- 本地化帮助文案 ---

// applyLocalizedHelp 在语言确定后刷新 Cobra 帮助文案。
func applyLocalizedHelp() {
	rootCmd.Short = i18n.T(i18n.KeyRootShort)
	rootCmd.Long = i18n.T(i18n.KeyRootLong)
	addCmd.Short = i18n.T(i18n.KeyAddShort)
	addCmd.Long = i18n.T(i18n.KeyAddLong)
	listCmd.Short = i18n.T(i18n.KeyListShort)
	listCmd.Long = i18n.T(i18n.KeyListLong)
	searchCmd.Short = i18n.T(i18n.KeySearchShort)
	searchCmd.Long = i18n.T(i18n.KeySearchLong)
	if f := searchCmd.Flags().Lookup("user"); f != nil {
		f.Usage = i18n.T(i18n.KeySearchFlagUser)
	}
	if f := searchCmd.Flags().Lookup("host"); f != nil {
		f.Usage = i18n.T(i18n.KeySearchFlagHost)
	}
	if f := searchCmd.Flags().Lookup("port"); f != nil {
		f.Usage = i18n.T(i18n.KeySearchFlagPort)
	}
	_ = rootCmd.PersistentFlags().Lookup("config")
	if f := rootCmd.PersistentFlags().Lookup("config"); f != nil {
		f.Usage = i18n.T(i18n.KeyConfigFlag)
	}
}

// --- Include 指令警告 ---

// maybeWarnInclude 在 list/search 前提示 config 含 Include 时 MVP 不会展开。
func maybeWarnInclude(stderr io.Writer, configPath string) {
	has, err := config.HasIncludeDirective(configPath)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "warning: %v\n", err)
		return
	}
	if !has {
		return
	}
	_, _ = fmt.Fprintf(stderr, "%s\n", i18n.T(i18n.KeyConfigIncludeSkipped))
}
