package wizard

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/huh"
)

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
