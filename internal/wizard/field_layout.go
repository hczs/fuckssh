package wizard

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

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
