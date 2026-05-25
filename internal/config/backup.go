package config

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/fuckssh/fuckssh/internal/platform"
)

// Backup 在 ~/.ssh/backup/ 创建带时间戳的 config 备份副本，并修剪为最近 20 份。
//
// 命名：config.fuckssh.bak.<timestamp>
func Backup(path string) (bakPath string, err error) {
	err = withConfigLock(path, func() error {
		var unlockErr error
		bakPath, unlockErr = backupUnlocked(path)
		return unlockErr
	})
	return bakPath, err
}

func backupUnlocked(path string) (bakPath string, err error) {
	src, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			// 首次写入 config 时原文件可能不存在，备份空文件无意义，跳过即可。
			return "", nil
		}
		return "", err
	}
	defer func() { _ = src.Close() }()

	backupDir, err := platform.BackupDir()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(backupDir, 0o700); err != nil {
		return "", fmt.Errorf("config: mkdir backup dir %q: %w", backupDir, err)
	}

	base := filepath.Base(path)
	ts := time.Now().UTC().Format("20060102T150405Z")
	bakPath = filepath.Join(backupDir, base+".fuckssh.bak."+ts)

	dst, err := os.OpenFile(bakPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return "", fmt.Errorf("config: create backup %q: %w", bakPath, err)
	}
	defer func() { _ = dst.Close() }()

	if _, err := io.Copy(dst, src); err != nil {
		return "", fmt.Errorf("config: write backup: %w", err)
	}
	if err := PruneBackups(backupDir, defaultBackupKeep); err != nil {
		return "", err
	}
	return bakPath, nil
}
