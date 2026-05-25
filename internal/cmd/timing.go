package cmd

import (
	"fmt"
	"io"
	"time"

	"github.com/fuckssh/fuckssh/internal/i18n"
	"github.com/spf13/cobra"
)

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
