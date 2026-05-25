package config

import (
	"fmt"
	"io"
	"os"
)

// RestoreFromBackup 用备份文件覆盖当前 config（用于 add 失败回滚）。
func RestoreFromBackup(backupPath, configPath string) error {
	if backupPath == "" {
		return fmt.Errorf("config: 无备份路径，无法恢复")
	}
	src, err := os.Open(backupPath)
	if err != nil {
		return fmt.Errorf("config: open backup %q: %w", backupPath, err)
	}
	defer func() { _ = src.Close() }()

	dst, err := os.OpenFile(configPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		return fmt.Errorf("config: open %q for restore: %w", configPath, err)
	}
	defer func() { _ = dst.Close() }()

	if _, err := io.Copy(dst, src); err != nil {
		return fmt.Errorf("config: restore %q: %w", configPath, err)
	}
	return nil
}

// RollbackAfterAddFailure 在 add 写入 config 失败后恢复。
//
//   - backupPath 非空：一律用备份文件覆盖当前 config（表单提交后已备份的场景）。
//   - backupPath 为空且本次前无 config、且已写入过 Host：删除本次新建的 config。
//   - 其余情况（如别名冲突、表单校验失败）：不改动 config。
func RollbackAfterAddFailure(configPath, backupPath string, configExistedBefore, configModified bool) error {
	if backupPath != "" {
		return RestoreFromBackup(backupPath, configPath)
	}
	if configModified && !configExistedBefore {
		err := os.Remove(configPath)
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	return nil
}

// RollbackConfig 兼容旧调用，等价于 configModified=true。
func RollbackConfig(configPath, backupPath string, configExistedBefore bool) error {
	return RollbackAfterAddFailure(configPath, backupPath, configExistedBefore, true)
}
