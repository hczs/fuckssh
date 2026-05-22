package config

import (
	"fmt"
	"os"
	"strings"
)

// ErrHostExists 表示 config 中已有同名 Host 别名。
var ErrHostExists = fmt.Errorf("config: Host 别名已存在")

// AppendHost 在 config 文件末尾追加一个 Host 块（不覆盖已有内容）。
func AppendHost(path string, entry HostEntry) error {
	if strings.TrimSpace(entry.Alias) == "" {
		return fmt.Errorf("config: Host 别名不能为空")
	}
	if strings.TrimSpace(entry.HostName) == "" {
		return fmt.Errorf("config: HostName 不能为空")
	}
	if strings.TrimSpace(entry.User) == "" {
		return fmt.Errorf("config: User 不能为空")
	}
	if strings.TrimSpace(entry.IdentityFile) == "" {
		return fmt.Errorf("config: IdentityFile 不能为空")
	}

	exists, err := hostAliasExists(path, entry.Alias)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("%w: %q（请换别名或手动编辑 config）", ErrHostExists, entry.Alias)
	}

	port := entry.Port
	if port == "" {
		port = "22"
	}

	block := formatHostBlock(entry.Alias, entry.HostName, entry.User, port, entry.IdentityFile)

	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0o600)
	if err != nil {
		return fmt.Errorf("config: open %q: %w", path, err)
	}
	defer f.Close()

	if _, err := f.WriteString(block); err != nil {
		return fmt.Errorf("config: append Host block: %w", err)
	}
	return nil
}

func hostAliasExists(path, alias string) (bool, error) {
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
func formatHostBlock(alias, hostName, user, port, identityFile string) string {
	var b strings.Builder
	b.WriteString("\n")
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

func formatIdentityFile(path string) string {
	if strings.ContainsAny(path, " \t") {
		return `"` + path + `"`
	}
	return path
}
