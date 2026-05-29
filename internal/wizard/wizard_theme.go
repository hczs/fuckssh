package wizard

// wizard_theme.go 合并主题配置与按键映射。
//
// 设计说明：
// - theme 和 keymap 都属于 TUI 外观/交互配置，内聚度高
// - 合并后减少 2 个碎片文件

import (
	"sync"

	catppuccin "github.com/catppuccin/go"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

// --- 主题配置 ---

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

// --- 按键映射 ---

// 向导统一导航：↑ 上一项；↓、Enter 下一项；Tab 仅用于单选切换（见 wizardAuthKeyMap）。

// wizardInputKeyMap 普通输入项。
// Enter 绑在 Next 上：huh 在非末项会禁用 Submit，否则 Enter 无法下一项。
func wizardInputKeyMap() huh.InputKeyMap {
	return huh.InputKeyMap{
		AcceptSuggestion: key.NewBinding(key.WithKeys("ctrl+e"), key.WithHelp("ctrl+e", "complete")),
		Prev:             key.NewBinding(key.WithKeys("up"), key.WithHelp("↑", "back")),
		Next:             key.NewBinding(key.WithKeys("down", "enter"), key.WithHelp("↓/enter", "next")),
		Submit:           key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "next")),
	}
}

// wizardCredentialKeyMap 凭证项：Enter 为测连（非下一项），↓ 为下一项。
func wizardCredentialKeyMap() huh.InputKeyMap {
	return huh.InputKeyMap{
		AcceptSuggestion: key.NewBinding(key.WithKeys("ctrl+e"), key.WithHelp("ctrl+e", "complete")),
		Prev:             key.NewBinding(key.WithKeys("up"), key.WithHelp("↑", "back")),
		Next:             key.NewBinding(key.WithKeys("down"), key.WithHelp("↓", "next")),
		Submit:           key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "test")),
	}
}

// wizardAuthKeyMap 认证单选：Tab 在选项间切换。
func wizardAuthKeyMap() huh.SelectKeyMap {
	tabToggle := key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "switch"))
	return huh.SelectKeyMap{
		Prev:   key.NewBinding(key.WithKeys("up"), key.WithHelp("↑", "back")),
		Next:   key.NewBinding(key.WithKeys("down", "enter"), key.WithHelp("↓/enter", "next")),
		Submit: key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "next")),
		Up:     key.NewBinding(key.WithKeys("up"), key.WithHelp("↑", "back")),
		Down:   key.NewBinding(key.WithKeys("down"), key.WithHelp("↓", "next")),
		Left:   tabToggle,
		Right:  tabToggle,
	}
}

// wizardFormKeyMap 供 collectAddInput 表单使用的完整按键表。
func wizardFormKeyMap() *huh.KeyMap {
	km := huh.NewDefaultKeyMap()
	km.Input = wizardInputKeyMap()
	km.Select = wizardAuthKeyMap()
	km.Confirm.Prev = key.NewBinding(key.WithKeys("up"), key.WithHelp("↑", "back"))
	km.Confirm.Next = key.NewBinding(key.WithKeys("down", "enter"), key.WithHelp("↓", "next"))
	km.Confirm.Submit = key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "submit"))
	km.Note.Prev = key.NewBinding(key.WithKeys("up"), key.WithHelp("↑", "back"))
	km.Note.Next = key.NewBinding(key.WithKeys("down", "enter"), key.WithHelp("↓", "next"))
	km.Note.Submit = key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "next"))
	return km
}

// applyInputNavPosition 与 huh.Input.WithPosition 一致；Enter 走 Next 绑定。
func applyInputNavPosition(km *huh.InputKeyMap, p huh.FieldPosition) {
	km.Prev.SetEnabled(!p.IsFirst())
	km.Next.SetEnabled(!p.IsLast())
	km.Submit.SetEnabled(p.IsLast())
}

// applyAuthNavPosition 认证单选前进键；Enter 走 Next。
func applyAuthNavPosition(km *huh.SelectKeyMap, p huh.FieldPosition) {
	km.Prev.SetEnabled(!p.IsFirst())
	km.Next.SetEnabled(!p.IsLast())
	km.Submit.SetEnabled(p.IsLast())
	km.Up.SetEnabled(!p.IsFirst())
	km.Down.SetEnabled(!p.IsLast())
}

// applyCredentialNavPosition 凭证项 Enter 始终为测连，不受 huh 末项逻辑影响。
func applyCredentialNavPosition(km *huh.InputKeyMap, p huh.FieldPosition) {
	km.Prev.SetEnabled(!p.IsFirst())
	km.Next.SetEnabled(!p.IsLast())
	km.Submit.SetEnabled(true)
}
