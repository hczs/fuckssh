package cmd

import (
	"errors"
	"fmt"
	"io"

	"github.com/fuckssh/fuckssh/internal/sshclient"
	"github.com/spf13/cobra"
)

// checkSSHFn 可在测试中注入，默认调用 sshclient.CheckSSH。
var checkSSHFn = sshclient.CheckSSH

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a VPS host via interactive wizard",
	Long:  "Run the interactive wizard to generate keys, update ssh config, and optionally deploy a public key.",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := warnIfSSHMissing(cmd.ErrOrStderr()); err != nil {
			return err
		}
		return fmt.Errorf("add: not implemented yet")
	},
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
