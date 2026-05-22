package sshclient

import (
	"errors"
	"os/exec"
)

// ErrSSHNotFound 表示 PATH 中找不到 ssh 可执行文件。
var ErrSSHNotFound = errors.New("ssh: not found in PATH")

// lookPath 可在测试中替换，默认使用 exec.LookPath。
var lookPath = exec.LookPath

// CheckSSH 在 PATH 中查找 ssh，成功时返回可执行文件路径。
func CheckSSH() (string, error) {
	path, err := lookPath("ssh")
	if err != nil {
		return "", ErrSSHNotFound
	}
	return path, nil
}
