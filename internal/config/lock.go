package config

import (
	"fmt"

	"github.com/gofrs/flock"
)

func lockFilePath(configPath string) string {
	return configPath + ".fuckssh.lock"
}

// withConfigLock 对同一 config 路径串行化 Backup / AppendHost 等写操作，避免并行 add 损坏文件。
func withConfigLock(configPath string, fn func() error) error {
	lock := flock.New(lockFilePath(configPath))
	if err := lock.Lock(); err != nil {
		return fmt.Errorf("config: acquire lock on %q: %w", configPath, err)
	}
	defer func() { _ = lock.Unlock() }()
	return fn()
}
