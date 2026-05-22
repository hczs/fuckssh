package config

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

// Backup 在 config 同目录创建带时间戳的备份副本。
//
// 命名：config.fuckssh.bak.<timestamp>（与架构 §5.4 一致）。
func Backup(path string) (bakPath string, err error) {
	src, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			// 首次写入 config 时原文件可能不存在，备份空文件无意义，跳过即可。
			return "", nil
		}
		return "", err
	}
	defer src.Close()

	dir := filepath.Dir(path)
	base := filepath.Base(path)
	ts := time.Now().UTC().Format("20060102T150405Z")
	bakPath = filepath.Join(dir, base+".fuckssh.bak."+ts)

	dst, err := os.OpenFile(bakPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return "", fmt.Errorf("config: create backup %q: %w", bakPath, err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return "", fmt.Errorf("config: write backup: %w", err)
	}
	return bakPath, nil
}
