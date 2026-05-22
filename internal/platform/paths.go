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

// BackupDir 返回 config 备份目录（~/.ssh/backup）。
func BackupDir() (string, error) {
	dir, err := SSHDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "backup"), nil
}

// KeysDir 返回本工具管理的私钥目录（~/.ssh/keys）。
func KeysDir() (string, error) {
	dir, err := SSHDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "keys"), nil
}

// IdentityFileRef 返回适合写入 ssh config 的 IdentityFile 值。
//
// 当 absKeyPath 位于 ~/.ssh 下时，使用相对 HOME 的路径（如 .ssh/keys/id_ed25519_fuckssh_my），
// 便于整份 .ssh 目录迁移；否则回退为绝对路径。
func IdentityFileRef(absKeyPath string) (string, error) {
	if absKeyPath == "" {
		return "", fmt.Errorf("platform: empty key path")
	}
	absKeyPath = filepath.Clean(absKeyPath)

	sshDir, err := SSHDir()
	if err != nil {
		return "", err
	}
	sshDir = filepath.Clean(sshDir)

	rel, err := filepath.Rel(sshDir, absKeyPath)
	if err != nil || strings.HasPrefix(rel, "..") {
		return absKeyPath, nil
	}
	return filepath.ToSlash(filepath.Join(".ssh", rel)), nil
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
