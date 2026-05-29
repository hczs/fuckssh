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
		entries       []HostEntry
		current       *HostEntry
		pendingRemark []string
		lineNum       int
	)

	for scanner.Scan() {
		lineNum++
		raw := scanner.Text()
		trimmed := strings.TrimSpace(raw)
		if trimmed == "" {
			// 空行不结束 Host 块（标准 ssh config 允许块内空行）。
			continue
		}
		if strings.HasPrefix(trimmed, "#") {
			// 始终收集注释，Host 指令到来时清空；块内注释会在遇到配置项时被丢弃。
			pendingRemark = append(pendingRemark, stripCommentLine(trimmed))
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
					Msg:     "Host requires at least one alias",
				}
			}
			aliases := strings.Fields(value)
			entry := HostEntry{
				Alias:     aliases[0],
				Aliases:   aliases,
				Port:      "22",
				Remark:    strings.Join(pendingRemark, "\n"),
				LineStart: lineNum,
			}
			pendingRemark = nil
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
					Msg:     "directive outside Host block",
				}
			}
			// 块内遇到配置项，清空 pendingRemark，防止块内注释泄漏到下一个 Host。
			pendingRemark = nil
			applyOption(current, key, value)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return entries, nil
}

// stripCommentLine 去掉行首 # 及紧随的一个空格。
func stripCommentLine(trimmed string) string {
	s := strings.TrimPrefix(trimmed, "#")
	return strings.TrimSpace(s)
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
			return "", "", fmt.Errorf("empty directive")
		}
		key = parts[0]
		if len(parts) > 1 {
			value = strings.Join(parts[1:], " ")
		}
		return key, value, nil
	}

	return line, "", nil
}

func applyOption(entry *HostEntry, key, value string) {
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
		// 静默忽略不支持的配置项（ssh config 选项众多，MVP 只提取关心的字段）。
	}
}
