package cmd

import (
	"errors"
	"os"

	"github.com/fuckssh/fuckssh/internal/config"
	"github.com/fuckssh/fuckssh/internal/sshclient"
	"github.com/fuckssh/fuckssh/internal/wizard"
)

// ExitCode 将错误映射为进程退出码（对齐架构 §4.4）。
func ExitCode(err error) int {
	if err == nil {
		return 0
	}
	if errors.Is(err, sshclient.ErrSSHNotFound) {
		return 5
	}
	if errors.Is(err, sshclient.ErrDeployFailed) {
		return 4
	}
	if errors.Is(err, wizard.ErrInvalidInput) || errors.Is(err, config.ErrHostExists) {
		return 1
	}
	var pe *config.ParseError
	if errors.As(err, &pe) {
		return 2
	}
	var pathErr *os.PathError
	if errors.As(err, &pathErr) {
		return 3
	}
	if errors.Is(err, os.ErrNotExist) || errors.Is(err, os.ErrPermission) {
		return 3
	}
	return 1
}
