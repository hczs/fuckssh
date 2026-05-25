package cmd

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/fuckssh/fuckssh/internal/i18n"
	"github.com/fuckssh/fuckssh/internal/sshclient"
	"github.com/fuckssh/fuckssh/internal/wizard"
	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"
)

// checkSSHFn 可在测试中注入，默认调用 sshclient.CheckSSH。
var checkSSHFn = sshclient.CheckSSH

// runWizardFn 可在测试中注入，默认调用交互式向导（传入 config 路径）。
var runWizardFn = wizard.Run

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a VPS host via interactive wizard",
	Long:  "Run the interactive wizard to generate keys, update ssh config, and optionally deploy a public key.",
	// 向导内 Ctrl+C 由 executeWithArgs 输出单行取消提示，不附带 usage。
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runAdd(cmd.OutOrStdout(), cmd.ErrOrStderr())
	},
}

func runAdd(stdout, stderr io.Writer) error {
	if err := requireSSH(stderr); err != nil {
		return err
	}

	configPath, err := ConfigFilePath()
	if err != nil {
		return err
	}

	dir := filepath.Dir(configPath)
	if dir == "" || dir == "." {
		dir = "."
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return &os.PathError{Op: "mkdir", Path: dir, Err: err}
	}

	result, err := runWizardFn(configPath)
	if err != nil {
		return err
	}

	if result.PasswordFlowComplete {
		printAddSuccess(stdout, stderr, configPath, result)
		return nil
	}

	if err := wizard.RunKeyFlow(configPath, result); err != nil {
		return err
	}

	printAddSuccess(stdout, stderr, configPath, result)
	return nil
}

func printAddSuccess(stdout, stderr io.Writer, configPath string, result *wizard.WizardResult) {
	wizard.WriteAddSuccessSummary(stderr, result, configPath)
	_, _ = fmt.Fprintf(stdout, "ssh %s\n", result.Alias)
	if result.PasswordFlowComplete && isTerminalWriter(stdout) {
		_, _ = fmt.Fprintf(stderr, "%s\n", i18n.T(i18n.KeySummaryReadyHint))
	}
}

func isTerminalWriter(w io.Writer) bool {
	f, ok := w.(*os.File)
	if !ok {
		return false
	}
	return isatty.IsTerminal(f.Fd())
}

// requireSSH 检测系统 ssh；缺失时输出警告与安装指引并终止。
func requireSSH(stderr io.Writer) error {
	_, err := checkSSHFn()
	if err == nil {
		return nil
	}
	if errors.Is(err, sshclient.ErrSSHNotFound) {
		_, _ = fmt.Fprintf(stderr, "%s\n%s\n", i18n.T(i18n.KeySSHMissingWarning), i18n.InstallOpenSSHGuide())
		return err
	}
	return err
}
