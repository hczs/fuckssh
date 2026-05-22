package keys

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"strings"
	"unicode"
)

const keyNamePrefix = "id_ed25519_fuckssh_"

// KeyPaths 根据 Host 别名返回私钥与公钥的**文件名**（不含目录）。
//
// 命名规则（与架构 §2.2.4 一致）：
//   - 有别名：id_ed25519_fuckssh_<sanitized_alias>
//   - 别名为空：id_ed25519_fuckssh_<hostname 的 SHA256 前 4 字节十六进制>
//
// sanitize 仅保留字母、数字、连字符与下划线，其余字符替换为下划线，避免出现路径穿越字符。
func KeyPaths(alias string) (priv, pub string) {
	base := sanitizeAlias(alias)
	if base == "" {
		base = defaultAliasSuffix()
	}
	priv = keyNamePrefix + base
	pub = priv + ".pub"
	return priv, pub
}

// SanitizeAlias 将 Host 别名规范为安全文件名片段（字母数字、连字符、下划线）。
func SanitizeAlias(alias string) string {
	return sanitizeAlias(alias)
}

func sanitizeAlias(alias string) string {
	alias = strings.TrimSpace(alias)
	if alias == "" {
		return ""
	}

	var b strings.Builder
	for _, r := range alias {
		switch {
		case unicode.IsLetter(r), unicode.IsDigit(r), r == '-', r == '_':
			b.WriteRune(r)
		default:
			b.WriteRune('_')
		}
	}
	s := b.String()
	for strings.Contains(s, "__") {
		s = strings.ReplaceAll(s, "__", "_")
	}
	return strings.Trim(s, "_")
}

func defaultAliasSuffix() string {
	host, err := os.Hostname()
	if err != nil || host == "" {
		host = "localhost"
	}
	sum := sha256.Sum256([]byte(host))
	return hex.EncodeToString(sum[:4])
}
