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
//   - 有别名：id_ed25519_fuckssh_<normalized_alias>
//   - 别名为空：id_ed25519_fuckssh_<hostname 的 SHA256 前 4 字节十六进制>
func KeyPaths(alias string) (priv, pub string) {
	base := NormalizeHostAlias(alias)
	if base == "" {
		base = defaultAliasSuffix()
	}
	priv = keyNamePrefix + base
	pub = priv + ".pub"
	return priv, pub
}

// NormalizeHostAlias 将用户输入规范为 SSH Host 别名：小写字母、数字与连字符。
// 点号与下划线会转为连字符；非法字符同样转为连字符；连续连字符合并并去掉首尾连字符。
func NormalizeHostAlias(alias string) string {
	alias = strings.TrimSpace(alias)
	if alias == "" {
		return ""
	}

	var b strings.Builder
	for _, r := range alias {
		switch {
		case unicode.IsLetter(r):
			b.WriteRune(unicode.ToLower(r))
		case unicode.IsDigit(r), r == '-':
			b.WriteRune(r)
		case r == '.', r == '_':
			b.WriteRune('-')
		default:
			b.WriteRune('-')
		}
	}
	s := b.String()
	for strings.Contains(s, "--") {
		s = strings.ReplaceAll(s, "--", "-")
	}
	return strings.Trim(s, "-")
}

// SanitizeAlias 与 NormalizeHostAlias 相同，保留旧名供调用方过渡。
func SanitizeAlias(alias string) string {
	return NormalizeHostAlias(alias)
}

func defaultAliasSuffix() string {
	host, err := os.Hostname()
	if err != nil || host == "" {
		host = "localhost"
	}
	sum := sha256.Sum256([]byte(host))
	return hex.EncodeToString(sum[:4])
}
