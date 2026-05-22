package platform

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// SSHDir 返回本机 OpenSSH 配置目录（通常为 ~/.ssh）。
func SSHDir() (string, error) {
	home, err := userHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".ssh"), nil
}

// DefaultConfigPath 返回默认的 ssh config 文件路径。
func DefaultConfigPath() (string, error) {
	dir, err := SSHDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config"), nil
}

// ExpandPath 将路径中的 ~ 展开为用户主目录。
func ExpandPath(path string) (string, error) {
	if path == "" {
		return "", nil
	}
	if path[0] != '~' {
		return path, nil
	}

	home, err := userHomeDir()
	if err != nil {
		return "", err
	}

	if path == "~" {
		return home, nil
	}
	if strings.HasPrefix(path, "~/") || strings.HasPrefix(path, `~\`) {
		return filepath.Join(home, path[2:]), nil
	}

	return "", fmt.Errorf("platform: unsupported tilde path %q", path)
}

// userHomeDir 按操作系统读取用户主目录环境变量。
func userHomeDir() (string, error) {
	var key string
	if runtime.GOOS == "windows" {
		key = "USERPROFILE"
	} else {
		key = "HOME"
	}

	home := os.Getenv(key)
	if home == "" {
		return "", fmt.Errorf("platform: %s is not set", key)
	}
	return home, nil
}
