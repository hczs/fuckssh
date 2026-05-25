package wizard

import (
	"path/filepath"
	"strings"

	"github.com/fuckssh/fuckssh/internal/i18n"
	"github.com/fuckssh/fuckssh/internal/keys"
)

// 向导总步数：模式(1) + 表单字段(5) + 确认(1)。
const wizardTotalSteps = 7

// stepTitle 生成「步骤 n/7 · 标签」标题。
func stepTitle(step int, labelKey string) string {
	return i18n.T(i18n.KeyWizardStepTitle, step, wizardTotalSteps, i18n.T(labelKey))
}

// aliasDescription 根据已填 HostName 生成别名说明（含预览）。
func aliasDescription(hostName *string) string {
	base := i18n.T(i18n.KeyWizardAliasDesc)
	if preview := keys.NormalizeHostAlias(strings.TrimSpace(*hostName)); preview != "" {
		return base + "\n" + i18n.T(i18n.KeyWizardAliasPreview, preview)
	}
	return base
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
