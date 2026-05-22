package cmd

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/fuckssh/fuckssh/internal/config"
	"github.com/fuckssh/fuckssh/internal/i18n"
	"github.com/fuckssh/fuckssh/internal/sshclient"
	"github.com/fuckssh/fuckssh/internal/wizard"
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

	if err := os.MkdirAll(dirOf(configPath), 0o700); err != nil {
		return &os.PathError{Op: "mkdir", Path: dirOf(configPath), Err: err}
	}

	result, err := runWizardFn(configPath)
	if err != nil {
		return err
	}

	if result.PasswordFlowComplete {
		printAddSuccess(stdout, stderr, configPath, result)
		return nil
	}

	configExisted := configFileExists(configPath)
	bakPath, err := config.Backup(configPath)
	if err != nil {
		return err
	}

	entry := config.HostEntry{
		Alias:        result.Alias,
		HostName:     result.HostName,
		User:         result.User,
		Port:         result.Port,
		IdentityFile: result.IdentityFile,
	}
	if err := config.AppendHost(configPath, entry); err != nil {
		_ = config.RollbackAfterAddFailure(configPath, bakPath, configExisted, true)
		return err
	}

	result.BackupPath = bakPath
	printAddSuccess(stdout, stderr, configPath, result)
	return nil
}

func printAddSuccess(stdout, stderr io.Writer, configPath string, result *wizard.WizardResult) {
	wizard.WriteAddSuccessSummary(stderr, result, configPath)
	fmt.Fprintf(stdout, "ssh %s\n", result.Alias)
}

func configFileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// requireSSH 检测系统 ssh；缺失时输出警告与安装指引并终止。
func requireSSH(stderr io.Writer) error {
	_, err := checkSSHFn()
	if err == nil {
		return nil
	}
	if errors.Is(err, sshclient.ErrSSHNotFound) {
		fmt.Fprintf(stderr, "%s\n%s\n", i18n.T(i18n.KeySSHMissingWarning), i18n.InstallOpenSSHGuide())
		return err
	}
	return err
}

func dirOf(path string) string {
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == os.PathSeparator || path[i] == '/' {
			return path[:i]
		}
	}
	return "."
}
