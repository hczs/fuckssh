package config

import (
	"fmt"
	"os"
	"strings"
)

// EditHost 通过行级编辑更新 ssh config 中指定 Host 条目的字段。
// 只替换被修改字段对应的行，保留文件中的未知配置项（如 ProxyJump、ForwardAgent 等）。
// 用户确认后、写文件前自动备份。使用文件锁保证并发安全。
func EditHost(path, alias string, newEntry HostEntry) error {
	return withConfigLock(path, func() error {
		return editHostUnlocked(path, alias, newEntry)
	})
}

func editHostUnlocked(path, alias string, newEntry HostEntry) error {
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

	// 如果新别名与当前别名不同，检查新别名是否已存在。
	newAlias := strings.TrimSpace(newEntry.Alias)
	if newAlias == "" {
		newAlias = target.Alias
	}
	if !strings.EqualFold(newAlias, alias) {
		exists, err := HostAliasExists(path, newAlias)
		if err != nil {
			return err
		}
		if exists {
			return fmt.Errorf("%w: %q (choose a different alias)", ErrHostExists, newAlias)
		}
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

	// --- 处理 Remark ---
	// 找到 Host 行上方的连续注释行（属于该块的 remark）。
	remarkStart := hostLine
	for remarkStart > 0 {
		prev := strings.TrimSpace(lines[remarkStart-1])
		if prev == "" || strings.HasPrefix(prev, "#") {
			remarkStart--
		} else {
			break
		}
	}

	// 生成新的 Remark 注释行。
	newRemark := strings.TrimSpace(newEntry.Remark)
	var remarkLines []string
	if newRemark != "" {
		remarkLines = buildRemarkLines(newRemark)
	}

	// --- 处理 Host 行 ---
	// 如果别名被修改，替换 Host 行。
	if !strings.EqualFold(newAlias, alias) {
		// 保留原有的多个别名，只替换第一个。
		newAliases := make([]string, len(target.Aliases))
		copy(newAliases, target.Aliases)
		newAliases[0] = newAlias
		lines[hostLine] = "Host " + strings.Join(newAliases, " ")
	}

	// --- 处理块内配置项 ---
	// 从 Host 行下一行开始，扫描到下一个 Host 行或文件末尾。
	blockEnd := hostLine + 1
	for blockEnd < len(lines) {
		trimmed := strings.TrimSpace(lines[blockEnd])
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			blockEnd++
			continue
		}
		if strings.HasPrefix(strings.ToLower(trimmed), "host ") {
			break
		}
		blockEnd++
	}

	// 跟踪哪些字段已在块中找到，用于后续插入缺失字段。
	foundPort := false

	// 在块范围内，按行匹配已知指令并替换。
	for i := hostLine + 1; i < blockEnd; i++ {
		trimmed := strings.TrimSpace(lines[i])
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		key, _, _ := splitDirective(trimmed)
		switch strings.ToLower(key) {
		case "hostname":
			if v := strings.TrimSpace(newEntry.HostName); v != "" {
				lines[i] = "    HostName " + v
			}
		case "user":
			if v := strings.TrimSpace(newEntry.User); v != "" {
				lines[i] = "    User " + v
			}
		case "port":
			foundPort = true
			if v := strings.TrimSpace(newEntry.Port); v != "" && v != "22" {
				lines[i] = "    Port " + v
			} else {
				// Port 为 22 或空时，删除该行（SSH 默认就是 22）。
				lines[i] = ""
			}
		case "identityfile":
			if v := strings.TrimSpace(newEntry.IdentityFile); v != "" {
				lines[i] = "    IdentityFile " + formatIdentityFile(v)
			}
		}
	}

	// 如果原来没有 Port 行，但新值非空且非默认 22，需要在 Host 行后插入。
	needInsertPort := !foundPort && strings.TrimSpace(newEntry.Port) != "" && newEntry.Port != "22"

	// --- 组装新文件内容 ---
	var result []string

	// Host 行之前的行（不含旧 remark）。
	result = append(result, lines[:remarkStart]...)

	// 新 remark。
	result = append(result, remarkLines...)

	// Host 行到块末尾（过滤掉空行）。
	for i := hostLine; i < blockEnd; i++ {
		if lines[i] != "" {
			result = append(result, lines[i])
		}
		// 在 Host 行后插入 Port（如果原来没有）。
		if i == hostLine && needInsertPort {
			result = append(result, "    Port "+strings.TrimSpace(newEntry.Port))
		}
	}

	// 块之后的行。
	result = append(result, lines[blockEnd:]...)

	content := strings.Join(result, "\n")
	content = strings.TrimRight(content, "\n")
	if content != "" {
		content += "\n"
	}

	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		return fmt.Errorf("config: write %q: %w", path, err)
	}
	return nil
}

// buildRemarkLines 将备注文本转换为 # 注释行列表。
func buildRemarkLines(remark string) []string {
	var lines []string
	for _, line := range strings.Split(remark, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		lines = append(lines, "# "+line)
	}
	return lines
}
