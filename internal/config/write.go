package config

import (
	"fmt"
	"os"
	"strings"
)

// ErrHostExists 表示 config 中已有同名 Host 别名。
var ErrHostExists = fmt.Errorf("config: host alias already exists")

// AppendHost 在 config 文件末尾追加一个 Host 块（不覆盖已有内容）。
func AppendHost(path string, entry HostEntry) error {
	return withConfigLock(path, func() error {
		return appendHostUnlocked(path, entry)
	})
}

func appendHostUnlocked(path string, entry HostEntry) error {
	if strings.TrimSpace(entry.Alias) == "" {
		return fmt.Errorf("config: host alias must not be empty")
	}
	if strings.TrimSpace(entry.HostName) == "" {
		return fmt.Errorf("config: HostName must not be empty")
	}
	if strings.TrimSpace(entry.User) == "" {
		return fmt.Errorf("config: User must not be empty")
	}
	if strings.TrimSpace(entry.IdentityFile) == "" {
		return fmt.Errorf("config: IdentityFile must not be empty")
	}

	exists, err := HostAliasExists(path, entry.Alias)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("%w: %q (choose a different alias or edit config manually)", ErrHostExists, entry.Alias)
	}

	port := entry.Port
	if port == "" {
		port = "22"
	}

	block := formatHostBlock(entry.Alias, entry.HostName, entry.User, port, entry.IdentityFile, entry.Remark)

	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0o600)
	if err != nil {
		return fmt.Errorf("config: open %q: %w", path, err)
	}
	defer func() { _ = f.Close() }()

	if _, err := f.WriteString(block); err != nil {
		return fmt.Errorf("config: append Host block: %w", err)
	}
	return nil
}

// HostAliasExists 检查 config 中是否已有同名 Host 别名（不区分大小写）。
func HostAliasExists(path, alias string) (bool, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false, nil
	}
	entries, err := ParseFile(path)
	if err != nil {
		return false, err
	}
	alias = strings.TrimSpace(alias)
	for _, e := range entries {
		for _, a := range e.Aliases {
			if strings.EqualFold(a, alias) {
				return true, nil
			}
		}
	}
	return false, nil
}

// formatHostBlock 按 OpenSSH 常用顺序生成 Host 块文本。
func formatHostBlock(alias, hostName, user, port, identityFile, remark string) string {
	var b strings.Builder
	b.WriteString("\n")
	if s := strings.TrimSpace(remark); s != "" {
		b.WriteString(formatRemarkComments(s))
	}
	b.WriteString("Host ")
	b.WriteString(alias)
	b.WriteString("\n")
	b.WriteString("    HostName ")
	b.WriteString(hostName)
	b.WriteString("\n")
	b.WriteString("    User ")
	b.WriteString(user)
	b.WriteString("\n")
	if port != "" && port != "22" {
		b.WriteString("    Port ")
		b.WriteString(port)
		b.WriteString("\n")
	}
	b.WriteString("    IdentityFile ")
	b.WriteString(formatIdentityFile(identityFile))
	b.WriteString("\n")
	return b.String()
}

// formatRemarkComments 将备注写成 Host 块上方的 # 注释行。
func formatRemarkComments(remark string) string {
	var b strings.Builder
	for _, line := range strings.Split(remark, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		b.WriteString("# ")
		b.WriteString(line)
		b.WriteString("\n")
	}
	return b.String()
}

func formatIdentityFile(path string) string {
	if strings.ContainsAny(path, " \t") {
		return `"` + path + `"`
	}
	return path
}
