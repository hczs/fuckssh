package config

import (
	"bufio"
	"os"
	"strings"
)

// HasIncludeDirective 扫描 config 是否含 Include 指令（MVP 不展开，仅用于提示用户）。
func HasIncludeDirective(path string) (bool, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	defer func() { _ = f.Close() }()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		trimmed := strings.TrimSpace(scanner.Text())
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		key, _, err := splitDirective(trimmed)
		if err != nil {
			continue
		}
		if strings.EqualFold(key, "include") {
			return true, nil
		}
	}
	return false, scanner.Err()
}
