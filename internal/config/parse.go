package config

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

// ParseError 描述 config 解析失败的位置与原因。
type ParseError struct {
	File    string
	Line    int
	Snippet string
	Msg     string
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("%s:%d: %s: %s", e.File, e.Line, e.Msg, strings.TrimSpace(e.Snippet))
}

// ParseFile 从路径读取并解析 ssh config。
func ParseFile(path string) ([]HostEntry, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()

	return Parse(f, path)
}

// Parse 从 reader 解析 ssh config（filename 仅用于错误信息）。
func Parse(r io.Reader, filename string) ([]HostEntry, error) {
	scanner := bufio.NewScanner(r)
	var (
		entries []HostEntry
		current *HostEntry
		lineNum int
	)

	for scanner.Scan() {
		lineNum++
		raw := scanner.Text()
		trimmed := strings.TrimSpace(raw)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		key, value, err := splitDirective(trimmed)
		if err != nil {
			return nil, &ParseError{
				File:    filename,
				Line:    lineNum,
				Snippet: raw,
				Msg:     err.Error(),
			}
		}

		switch strings.ToLower(key) {
		case "host":
			if value == "" {
				return nil, &ParseError{
					File:    filename,
					Line:    lineNum,
					Snippet: raw,
					Msg:     "Host 需要至少一个别名",
				}
			}
			aliases := strings.Fields(value)
			entry := HostEntry{
				Alias:     aliases[0],
				Aliases:   aliases,
				Port:      "22",
				LineStart: lineNum,
			}
			entries = append(entries, entry)
			current = &entries[len(entries)-1]
		case "include":
			// MVP 不展开 Include；顶层或块内均跳过。
			continue
		default:
			if current == nil {
				return nil, &ParseError{
					File:    filename,
					Line:    lineNum,
					Snippet: raw,
					Msg:     "Host 块之外的指令",
				}
			}
			if err := applyOption(current, key, value); err != nil {
				return nil, &ParseError{
					File:    filename,
					Line:    lineNum,
					Snippet: raw,
					Msg:     err.Error(),
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return entries, nil
}

func splitDirective(line string) (key, value string, err error) {
	if idx := strings.IndexAny(line, "= \t"); idx >= 0 {
		key = strings.TrimSpace(line[:idx])
		value = strings.TrimSpace(line[idx+1:])
		if line[idx] == '=' {
			return key, value, nil
		}
		// 空格分隔：value 为第一个 token 之后的部分或整段 fields
		parts := strings.Fields(line)
		if len(parts) == 0 {
			return "", "", fmt.Errorf("空指令")
		}
		key = parts[0]
		if len(parts) > 1 {
			value = strings.Join(parts[1:], " ")
		}
		return key, value, nil
	}

	return line, "", nil
}

func applyOption(entry *HostEntry, key, value string) error {
	switch strings.ToLower(key) {
	case "hostname":
		entry.HostName = value
	case "user":
		entry.User = value
	case "port":
		entry.Port = value
	case "identityfile":
		entry.IdentityFile = value
	default:
		return fmt.Errorf("不支持的配置项 %q", key)
	}
	return nil
}
