package wizard

// wizard_ui.go 合并字段布局渲染辅助与端口校验。

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

// --- 字段布局渲染 ---

// setInlineInputWidth 为「标题 + 输入」同行布局计算 textinput 宽度。
func setInlineInputWidth(width int, styles *huh.FieldStyles, title string, ti *textinput.Model) {
	frame := styles.Base.GetHorizontalFrameSize()
	promptW := lipgloss.Width(ti.PromptStyle.Render(ti.Prompt))
	titleW := lipgloss.Width(styles.Title.Render(title))
	ti.Width = width - frame - promptW - titleW - 2
	if ti.Width < 12 {
		ti.Width = 12
	}
}

// renderInlineField 标题与输入同一行；说明/错误另起一行（更紧凑）。
func renderInlineField(width, height int, styles *huh.FieldStyles, title, inputRow string, below ...string) string {
	titleRendered := ""
	if title != "" {
		titleRendered = styles.Title.Render(title) + " "
	}
	row := lipgloss.JoinHorizontal(lipgloss.Top, titleRendered, inputRow)

	var sb strings.Builder
	sb.WriteString(row)
	for _, line := range below {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		sb.WriteString("\n")
		sb.WriteString(line)
	}
	return styles.Base.Width(width).Height(height).Render(sb.String())
}

// --- 端口校验 ---

// validatePort 校验 SSH 端口为 1–65535；空字符串由调用方补默认值。
func validatePort(port string) error {
	port = strings.TrimSpace(port)
	if port == "" {
		return nil
	}
	n, err := strconv.Atoi(port)
	if err != nil || n < 1 || n > 65535 {
		return fmt.Errorf("%w: port must be between 1 and 65535", ErrInvalidInput)
	}
	return nil
}
