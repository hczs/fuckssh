package wizard

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/fuckssh/fuckssh/internal/i18n"
	"github.com/fuckssh/fuckssh/internal/keys"
)

// 向导总步数：模式(1) + 表单字段(5) + 确认(1)。
const wizardTotalSteps = 7

// connFeedback 供密码/私钥字段 DescriptionFunc 展示测连结果（堆叠布局下保留在已填项旁）。
type connFeedback struct {
	message string
}

func (c *connFeedback) setTesting() {
	c.message = i18n.T(i18n.KeyWizardTestingConnInline)
}

func (c *connFeedback) setOK() {
	c.message = "✓ " + i18n.T(i18n.KeyWizardConnOK)
}

func (c *connFeedback) setIdleHint() {
	c.message = i18n.T(i18n.KeyWizardHostKeyHint)
}

// stepTitle 生成「步骤 n/6 · 标签」标题。
func stepTitle(step int, labelKey string) string {
	return i18n.T(i18n.KeyWizardStepTitle, step, wizardTotalSteps, i18n.T(labelKey))
}

// aliasDescription 根据已填 HostName 生成别名说明（含预览）。
func aliasDescription(hostName *string) string {
	base := i18n.T(i18n.KeyWizardAliasDesc)
	if preview := keys.SanitizeAlias(strings.TrimSpace(*hostName)); preview != "" {
		return base + "\n" + i18n.T(i18n.KeyWizardAliasPreview, preview)
	}
	return base
}

// effectiveAliasForDisplay 用于确认屏展示（空则按 HostName 推导）。
func effectiveAliasForDisplay(alias, hostName string) string {
	if a := strings.TrimSpace(alias); a != "" {
		return keys.SanitizeAlias(a)
	}
	return keys.SanitizeAlias(hostName)
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

// formatTarget 用于确认摘要中的 user@host:port。
func formatTarget(host, user, port string) string {
	p := strings.TrimSpace(port)
	if p == "" {
		p = "22"
	}
	return fmt.Sprintf("%s@%s:%s", strings.TrimSpace(user), strings.TrimSpace(host), p)
}
