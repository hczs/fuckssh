package wizard

import (
	"sync"

	catppuccin "github.com/catppuccin/go"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

var (
	wizardTheme     *huh.Theme
	wizardThemeOnce sync.Once
)

// WizardTheme 返回 add 向导的自适应主题（亮/暗终端均可读）。
// 基于 Catppuccin，并单独强化 Blurred，使 LayoutStack 中已填行更易辨认。
func WizardTheme() *huh.Theme {
	wizardThemeOnce.Do(func() {
		wizardTheme = buildWizardTheme()
	})
	return wizardTheme
}

func buildWizardTheme() *huh.Theme {
	t := huh.ThemeCatppuccin()

	light := catppuccin.Latte
	dark := catppuccin.Mocha
	var (
		subtext1 = lipgloss.AdaptiveColor{Light: light.Subtext1().Hex, Dark: dark.Subtext1().Hex}
		subtext0 = lipgloss.AdaptiveColor{Light: light.Subtext0().Hex, Dark: dark.Subtext0().Hex}
		overlay1 = lipgloss.AdaptiveColor{Light: light.Overlay1().Hex, Dark: dark.Overlay1().Hex}
		overlay0 = lipgloss.AdaptiveColor{Light: light.Overlay0().Hex, Dark: dark.Overlay0().Hex}
		green    = lipgloss.AdaptiveColor{Light: light.Green().Hex, Dark: dark.Green().Hex}
	)

	// 已填行（Blurred）：标题与正文分层变灰，比 ThemeCatppuccin 默认更明显。
	t.Blurred.Title = t.Blurred.Title.Foreground(subtext0).Bold(false)
	t.Blurred.NoteTitle = t.Blurred.NoteTitle.Foreground(subtext0)
	t.Blurred.Description = t.Blurred.Description.Foreground(overlay1)
	t.Blurred.TextInput.Text = t.Blurred.TextInput.Text.Foreground(subtext1)
	t.Blurred.TextInput.Prompt = t.Blurred.TextInput.Prompt.Foreground(overlay1)
	t.Blurred.TextInput.Placeholder = t.Blurred.TextInput.Placeholder.Foreground(overlay1)
	t.Blurred.SelectedOption = t.Blurred.SelectedOption.Foreground(subtext1)
	t.Blurred.UnselectedOption = t.Blurred.UnselectedOption.Foreground(overlay1)
	t.Blurred.Option = t.Blurred.Option.Foreground(subtext1)

	// 聚焦行 placeholder 略淡于正文，失焦预览再淡一档。
	t.Focused.TextInput.Placeholder = t.Focused.TextInput.Placeholder.Foreground(overlay0)

	// 测连成功态复用主题绿色。
	t.Focused.SelectedOption = t.Focused.SelectedOption.Foreground(green)

	return t
}

// wizardStatusOKStyle 测连成功等正向状态文案样式。
func wizardStatusOKStyle(t *huh.Theme) lipgloss.Style {
	if t == nil {
		return lipgloss.NewStyle()
	}
	return t.Focused.SelectedOption
}
