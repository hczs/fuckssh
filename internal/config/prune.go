package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const defaultBackupKeep = 20

// PruneBackups 在 backupDir 中仅保留最近 keep 个 fuckssh config 备份（按修改时间）。
func PruneBackups(backupDir string, keep int) error {
	if keep <= 0 {
		return nil
	}

	entries, err := os.ReadDir(backupDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("config: read backup dir %q: %w", backupDir, err)
	}

	var files []os.FileInfo
	for _, e := range entries {
		if e.IsDir() || !strings.Contains(e.Name(), ".fuckssh.bak.") {
			continue
		}
		info, err := e.Info()
		if err != nil {
			return fmt.Errorf("config: stat backup %q: %w", e.Name(), err)
		}
		files = append(files, info)
	}

	if len(files) <= keep {
		return nil
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].ModTime().After(files[j].ModTime())
	})

	for _, info := range files[keep:] {
		p := filepath.Join(backupDir, info.Name())
		if err := os.Remove(p); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("config: remove old backup %q: %w", p, err)
		}
	}
	return nil
}
