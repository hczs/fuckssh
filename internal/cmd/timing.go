package cmd

import (
	"fmt"
	"io"
	"os"
	"reflect"
	"time"

	"github.com/fuckssh/fuckssh/internal/i18n"
	"github.com/spf13/cobra"
)

// skipCmdElapsed 为 true 时（例如 --help）不在 stderr 底部打印命令耗时。
var skipCmdElapsed bool

func printCmdElapsed(w io.Writer, elapsed time.Duration) {
	_, _ = fmt.Fprint(w, i18n.T(i18n.KeyCmdElapsedMs, elapsed.Milliseconds()))
}

// commandArgs 读取本次将要执行的参数（测试里来自 SetArgs，正式运行时来自 os.Args）。
func commandArgs(root *cobra.Command) []string {
	if root == nil {
		return nil
	}
	v := reflect.ValueOf(root).Elem().FieldByName("args")
	if v.IsValid() && v.Kind() == reflect.Slice && v.Len() > 0 {
		n := v.Len()
		out := make([]string, n)
		for i := 0; i < n; i++ {
			out[i] = v.Index(i).String()
		}
		return out
	}
	if len(os.Args) > 1 {
		return os.Args[1:]
	}
	return nil
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

// clearCommandArgs 清空 Cobra 记录的 args，避免上一轮 SetArgs/--help 影响下一轮测试或命令。
func clearCommandArgs(cmd *cobra.Command) {
	if cmd == nil {
		return
	}
	v := reflect.ValueOf(cmd).Elem().FieldByName("args")
	if v.IsValid() && v.CanSet() {
		v.Set(reflect.Zero(v.Type()))
	}
	for _, sub := range cmd.Commands() {
		clearCommandArgs(sub)
	}
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

