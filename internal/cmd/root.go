package cmd

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/fuckssh/fuckssh/internal/i18n"
	"github.com/fuckssh/fuckssh/internal/platform"
	"github.com/fuckssh/fuckssh/internal/wizard"
	"github.com/spf13/cobra"
)

var helpLocalizedOnce sync.Once

// currentRunArgs 保存本轮 Execute 的参数（正式运行来自 os.Args，测试来自 ExecuteWithArgs）。
var currentRunArgs []string

var (
	rootCmd = &cobra.Command{
		Use:   "fuckssh",
		Short: "Manage ~/.ssh/config for VPS hosts",
		Long:  "fuckssh is a cross-platform CLI for SSH config, host listing, and search.",
	}
	// configFileFlag 允许用 --config 覆盖默认的 ~/.ssh/config 路径。
	configFileFlag string
)

// runArgsForHelp 返回用于判断是否输出 help 耗时的参数列表。
func runArgsForHelp() []string {
	if currentRunArgs != nil {
		return currentRunArgs
	}
	if len(os.Args) > 1 {
		return os.Args[1:]
	}
	return nil
}

// Execute runs the root command (args from os.Args).
func Execute() error {
	args := os.Args[1:]
	if len(args) == 0 {
		return executeWithArgs(nil)
	}
	return executeWithArgs(args)
}

// ExecuteWithArgs runs the root command with explicit args (for tests).
func ExecuteWithArgs(args []string) error {
	return executeWithArgs(args)
}

func executeWithArgs(args []string) error {
	currentRunArgs = args
	defer func() { currentRunArgs = nil }()

	resetHelpFlags(rootCmd)
	runArgs := args
	if runArgs == nil {
		runArgs = runArgsForHelp()
	}

	rootCmd.SetArgs(args)
	start := time.Now()
	err := rootCmd.Execute()
	if err != nil {
		printCommandError(rootCmd, args, err)
	}
	// 交互模式下 add 命令自行管理耗时输出，此处跳过
	if !helpInArgs(runArgs) && !skipElapsedOutput {
		printCmdElapsed(rootCmd.ErrOrStderr(), time.Since(start))
	}
	skipElapsedOutput = false
	resetHelpFlags(rootCmd)
	return err
}

// printCommandError 输出 RunE 失败信息；add 子命令取消时仅一行友好提示。
func printCommandError(root *cobra.Command, args []string, err error) {
	cmd, _, findErr := root.Find(args)
	if findErr != nil {
		cmd = root
	}
	w := cmd.ErrOrStderr()
	if wizard.IsCancelled(err) {
		_, _ = fmt.Fprintln(w, wizard.CancelMessage(err))
		return
	}
	if cmd.SilenceErrors || root.SilenceErrors {
		_, _ = fmt.Fprintln(w, cmd.ErrPrefix(), err.Error())
	}
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
	runArgs := runArgsForHelp()
	help := cmd.Flags().Changed("help") && helpInArgs(runArgs)
	if help {
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

// isReadonlyCmd 判断是否为只读/非交互子命令（跳过交互式语言选择以加快响应）。
// add 命令带 --host 时为非交互模式，同样跳过。
func isReadonlyCmd(cmd *cobra.Command) bool {
	switch cmd.Name() {
	case "list", "search", "version":
		return true
	case "add":
		return isNonInteractive()
	default:
		return false
	}
}

func applyLocalizedHelpOnce() {
	helpLocalizedOnce.Do(applyLocalizedHelp)
}
