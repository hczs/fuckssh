package cmd

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/fuckssh/fuckssh/internal/config"
	"github.com/fuckssh/fuckssh/internal/sshclient"
	"github.com/fuckssh/fuckssh/internal/wizard"
	"github.com/spf13/cobra"
)

// checkSSHFn 可在测试中注入，默认调用 sshclient.CheckSSH。
var checkSSHFn = sshclient.CheckSSH

// runWizardFn 可在测试中注入，默认调用交互式向导。
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
	if err := warnIfSSHMissing(stderr); err != nil {
		return err
	}

	result, err := runWizardFn()
	if err != nil {
		return err
	}

	configPath, err := ConfigFilePath()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dirOf(configPath), 0o700); err != nil {
		return &os.PathError{Op: "mkdir", Path: dirOf(configPath), Err: err}
	}

	bakPath, err := config.Backup(configPath)
	if err != nil {
		return err
	}
	if bakPath != "" {
		fmt.Fprintf(stderr, "已备份 config 至 %s\n", bakPath)
	}

	entry := config.HostEntry{
		Alias:        result.Alias,
		HostName:     result.HostName,
		User:         result.User,
		Port:         result.Port,
		IdentityFile: result.IdentityFile,
	}
	if err := config.AppendHost(configPath, entry); err != nil {
		return err
	}

	fmt.Fprintf(stdout, "配置已写入 %s\n", configPath)
	fmt.Fprintf(stdout, "现在可以执行: ssh %s\n", result.Alias)
	return nil
}

// warnIfSSHMissing 检测系统 ssh；缺失时打印指引但不阻止后续流程。
func warnIfSSHMissing(stderr io.Writer) error {
	_, err := checkSSHFn()
	if err == nil {
		return nil
	}
	if errors.Is(err, sshclient.ErrSSHNotFound) {
		fmt.Fprintf(stderr, "警告: 未在 PATH 中找到 ssh 客户端\n%v\n", err)
		return nil
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
