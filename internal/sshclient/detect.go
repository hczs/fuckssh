package sshclient

import (
	"errors"
	"fmt"
	"os/exec"

	"github.com/fuckssh/fuckssh/internal/platform"
)

// ErrSSHNotFound 表示 PATH 中找不到 ssh 可执行文件。
var ErrSSHNotFound = errors.New("ssh: not found in PATH")

// lookPath 可在测试中替换，默认使用 exec.LookPath。
var lookPath = exec.LookPath

// CheckSSH 在 PATH 中查找 ssh，成功时返回可执行文件路径。
// 未找到时返回 ErrSSHNotFound，并在错误信息中附带分平台安装指引。
func CheckSSH() (string, error) {
	path, err := lookPath("ssh")
	if err != nil {
		return "", fmt.Errorf("%w\n%s", ErrSSHNotFound, platform.InstallOpenSSHGuide())
	}
	return path, nil
}
