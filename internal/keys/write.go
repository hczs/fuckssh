package keys

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/fuckssh/fuckssh/internal/platform"
)

// ErrKeyFileExists 表示目标密钥路径已存在，拒绝覆盖。
var ErrKeyFileExists = errors.New("keys: key file already exists")

// WriteKeyPair 将密钥对写入 dir，文件名为 baseName（私钥）与 baseName.pub（公钥）。
// 任一目标已存在时返回 ErrKeyFileExists，不静默覆盖。
func WriteKeyPair(dir, baseName string, kp KeyPair) error {
	privPath := filepath.Join(dir, baseName)
	pubPath := privPath + ".pub"

	if err := refuseIfExists(privPath, pubPath); err != nil {
		return err
	}

	// 私钥先写临时权限，落盘后再收紧为 0600（Unix）。
	if err := os.WriteFile(privPath, kp.PrivatePEM, 0o600); err != nil {
		return fmt.Errorf("keys: write private key: %w", err)
	}
	if err := platform.SetPrivateKeyPerm(privPath); err != nil {
		return err
	}

	if err := os.WriteFile(pubPath, []byte(kp.PublicLine), 0o644); err != nil {
		_ = os.Remove(privPath)
		return fmt.Errorf("keys: write public key: %w", err)
	}

	return nil
}

func refuseIfExists(paths ...string) error {
	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			return fmt.Errorf("%w: %s", ErrKeyFileExists, p)
		} else if !os.IsNotExist(err) {
			return fmt.Errorf("keys: stat %q: %w", p, err)
		}
	}
	return nil
}
