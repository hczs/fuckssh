package keys

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/fuckssh/fuckssh/internal/platform"
)

// CopyKeyPair 将已有私钥（及可选 .pub）复制到 destDir/baseName。
// 目标已存在时返回 ErrKeyFileExists；不修改源文件。
func CopyKeyPair(srcPrivPath, destDir, baseName string) (destPrivPath string, err error) {
	destPrivPath = filepath.Join(destDir, baseName)
	destPubPath := destPrivPath + ".pub"

	if err := refuseIfExists(destPrivPath, destPubPath); err != nil {
		return "", err
	}

	if err := os.MkdirAll(destDir, 0o700); err != nil {
		return "", fmt.Errorf("keys: mkdir %q: %w", destDir, err)
	}

	if err := copyFile(srcPrivPath, destPrivPath, 0o600); err != nil {
		return "", err
	}
	if err := platform.SetPrivateKeyPerm(destPrivPath); err != nil {
		_ = os.Remove(destPrivPath)
		return "", err
	}

	srcPub := srcPrivPath + ".pub"
	if _, err := os.Stat(srcPub); err == nil {
		if err := copyFile(srcPub, destPubPath, 0o644); err != nil {
			_ = os.Remove(destPrivPath)
			return "", err
		}
	} else if !os.IsNotExist(err) {
		_ = os.Remove(destPrivPath)
		return "", fmt.Errorf("keys: stat %q: %w", srcPub, err)
	}

	return destPrivPath, nil
}

func copyFile(src, dst string, mode os.FileMode) error {
	in, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("keys: open %q: %w", src, err)
	}
	defer in.Close()

	out, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_EXCL, mode)
	if err != nil {
		return fmt.Errorf("keys: create %q: %w", dst, err)
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		_ = os.Remove(dst)
		return fmt.Errorf("keys: copy to %q: %w", dst, err)
	}
	return nil
}
