package platform

import (
	"fmt"
	"os"
	"runtime"
)

// SetPrivateKeyPerm 将私钥文件权限设为仅所有者可读写。
// Unix 使用 0600；Windows 不强制 Unix 语义（OpenSSH 客户端自行处理 ACL）。
func SetPrivateKeyPerm(path string) error {
	if runtime.GOOS == "windows" {
		return nil
	}
	if err := os.Chmod(path, 0o600); err != nil {
		return fmt.Errorf("platform: chmod private key %q: %w", path, err)
	}
	return nil
}
