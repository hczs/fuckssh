package wizard

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/fuckssh/fuckssh/internal/i18n"
	"github.com/fuckssh/fuckssh/internal/keys"
)

// fieldLabel 表单项标签，形如「1. Host 别名」（无步骤 x/y 噪音）。
func fieldLabel(n int, labelKey string) string {
	return fmt.Sprintf("%d. %s", n, i18n.T(labelKey))
}

// confirmTitle 确认页标题。
func confirmTitle() string {
	return i18n.T(i18n.KeyWizardConfirmStep)
}

// aliasPlaceholder 别名留空时，根据主机地址展示将写入的别名（实时预览）。
func aliasPlaceholder(hostName string) string {
	if gen := keys.NormalizeHostAlias(strings.TrimSpace(hostName)); gen != "" {
		return gen
	}
	return i18n.T(i18n.KeyWizardAliasEmptyHint)
}

// safeTTYString 将路径等字符串转为 TUI 安全展示（反斜杠在 lipgloss 中会被当成转义吃掉）。
func safeTTYString(s string) string {
	return filepath.ToSlash(s)
}

// effectivePort 空端口视为 22。
func effectivePort(port string) string {
	if strings.TrimSpace(port) == "" {
		return "22"
	}
	return strings.TrimSpace(port)
}
