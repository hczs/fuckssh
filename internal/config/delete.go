package config

import (
	"fmt"
	"os"
	"strings"
)

// DeleteHost 从 ssh config 中删除指定别名的 Host 块。
// 执行前自动备份，使用文件锁保证并发安全。
func DeleteHost(path, alias string) error {
	return withConfigLock(path, func() error {
		return deleteHostUnlocked(path, alias)
	})
}

func deleteHostUnlocked(path, alias string) error {
	alias = strings.TrimSpace(alias)
	if alias == "" {
		return fmt.Errorf("config: host alias must not be empty")
	}

	// 解析 config，找到目标条目。
	entries, err := ParseFile(path)
	if err != nil {
		return err
	}

	var target *HostEntry
	for i := range entries {
		for _, a := range entries[i].Aliases {
			if strings.EqualFold(a, alias) {
				target = &entries[i]
				break
			}
		}
		if target != nil {
			break
		}
	}
	if target == nil {
		return fmt.Errorf("%w: %q", ErrHostNotFound, alias)
	}

	// 备份（已在锁内，使用 backupUnlocked 避免死锁）。
	if _, err := backupUnlocked(path); err != nil {
		return err
	}

	// 读取原始文件行。
	raw, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("config: read %q: %w", path, err)
	}
	lines := strings.Split(string(raw), "\n")

	hostLine := target.LineStart - 1 // 转为 0-based

	// 向前扫描：收集 Host 块上方的注释与空行（属于该块的 remark）。
	remarkStart := hostLine
	for remarkStart > 0 {
		prev := strings.TrimSpace(lines[remarkStart-1])
		if prev == "" || strings.HasPrefix(prev, "#") {
			remarkStart--
		} else {
			break
		}
	}

	// 向后扫描：找到该 Host 块的最后一行配置。
	lastConfigLine := hostLine
	for i := hostLine + 1; i < len(lines); i++ {
		trimmed := strings.TrimSpace(lines[i])
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		if strings.HasPrefix(strings.ToLower(trimmed), "host ") {
			break
		}
		lastConfigLine = i
	}

	// 删除目标行范围（remark 到块的最后一行配置）。
	newLines := append(lines[:remarkStart], lines[lastConfigLine+1:]...)
	content := strings.Join(newLines, "\n")
	content = strings.TrimRight(content, "\n")
	if content != "" {
		content += "\n"
	}

	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		return fmt.Errorf("config: write %q: %w", path, err)
	}
	return nil
}
