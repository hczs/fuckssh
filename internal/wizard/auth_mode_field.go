package wizard

import (
	"io"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/fuckssh/fuckssh/internal/i18n"
)

// authModeField 水平单选：密码 / SSH 密钥（Tab 在选项间切换，1/2 快捷键）。
type authModeField struct {
	mode *ConnectionMode

	title string
	key   string
	id    int

	onChange func(prev, next ConnectionMode)

	focused    bool
	accessible bool
	width      int
	height     int
	theme      *huh.Theme
	keymap     huh.SelectKeyMap
}

var authModeFieldIDSeq int

func nextAuthModeFieldID() int {
	authModeFieldIDSeq++
	return authModeFieldIDSeq
}

// NewAuthModeField 创建认证方式单选字段。
func NewAuthModeField(mode *ConnectionMode, onChange func(prev, next ConnectionMode)) *authModeField {
	return &authModeField{
		mode:     mode,
		onChange: onChange,
		id:       nextAuthModeFieldID(),
		keymap:   wizardAuthKeyMap(),
	}
}

func (f *authModeField) Value(v *ConnectionMode) *authModeField {
	f.mode = v
	return f
}

func (f *authModeField) Key(k string) *authModeField {
	f.key = k
	return f
}

func (f *authModeField) Title(title string) *authModeField {
	f.title = title
	return f
}

func (f *authModeField) activeStyles() *huh.FieldStyles {
	if f.theme == nil {
		f.theme = WizardTheme()
	}
	if f.focused {
		return &f.theme.Focused
	}
	return &f.theme.Blurred
}

func (f *authModeField) setMode(next ConnectionMode) {
	if f.mode == nil || *f.mode == next {
		return
	}
	prev := *f.mode
	*f.mode = next
	if f.onChange != nil {
		f.onChange(prev, next)
	}
}

func (f *authModeField) toggleMode() {
	if f.mode == nil {
		return
	}
	if *f.mode == ModePassword {
		f.setMode(ModeKey)
	} else {
		f.setMode(ModePassword)
	}
}

func (f *authModeField) optionLabel(m ConnectionMode) string {
	if m == ModePassword {
		return i18n.T(i18n.KeyWizardAuthPasswordShort)
	}
	return i18n.T(i18n.KeyWizardAuthKeyShort)
}

func (f *authModeField) renderOptions(styles *huh.FieldStyles) string {
	selected := ModePassword
	if f.mode != nil {
		selected = *f.mode
	}
	renderOne := func(m ConnectionMode) string {
		marker := "○ "
		style := styles.UnselectedOption
		if m == selected {
			marker = "◉ "
			style = styles.SelectedOption
		}
		return marker + style.Render(f.optionLabel(m))
	}
	return lipgloss.JoinHorizontal(lipgloss.Left,
		renderOne(ModePassword),
		"  ",
		renderOne(ModeKey),
	)
}

func (f *authModeField) Init() tea.Cmd { return nil }

func (f *authModeField) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, f.keymap.Prev), key.Matches(msg, f.keymap.Up):
			return f, huh.PrevField
		case key.Matches(msg, f.keymap.Next), key.Matches(msg, f.keymap.Submit), key.Matches(msg, f.keymap.Down):
			return f, huh.NextField
		case key.Matches(msg, f.keymap.Left), key.Matches(msg, f.keymap.Right):
			f.toggleMode()
		case msg.String() == "1":
			f.setMode(ModePassword)
		case msg.String() == "2":
			f.setMode(ModeKey)
		}
	}
	return f, nil
}

func (f *authModeField) View() string {
	styles := f.activeStyles()
	row := lipgloss.JoinHorizontal(lipgloss.Top, styles.Title.Render(f.title)+" ", f.renderOptions(styles))
	return styles.Base.Width(f.width).Height(f.height).Render(row)
}

func (f *authModeField) Error() error { return nil }
func (f *authModeField) Skip() bool   { return false }
func (f *authModeField) Zoom() bool   { return false }

func (f *authModeField) Focus() tea.Cmd {
	f.focused = true
	return nil
}

func (f *authModeField) Blur() tea.Cmd {
	f.focused = false
	return nil
}

func (f *authModeField) KeyBinds() []key.Binding {
	return []key.Binding{
		f.keymap.Prev, f.keymap.Left, f.keymap.Down, f.keymap.Submit,
	}
}

func (f *authModeField) Run() error { return huh.Run(f) }

func (f *authModeField) RunAccessible(w io.Writer, r io.Reader) error {
	_ = w
	_ = r
	return nil
}

func (f *authModeField) WithTheme(theme *huh.Theme) huh.Field {
	if theme != nil && f.theme == nil {
		f.theme = theme
	}
	return f
}

func (f *authModeField) WithAccessible(accessible bool) huh.Field {
	f.accessible = accessible
	return f
}

func (f *authModeField) WithKeyMap(k *huh.KeyMap) huh.Field {
	if k != nil {
		f.keymap = k.Select
	}
	return f
}

func (f *authModeField) WithWidth(width int) huh.Field {
	f.width = width
	return f
}

func (f *authModeField) WithHeight(height int) huh.Field {
	f.height = height
	return f
}

func (f *authModeField) WithPosition(p huh.FieldPosition) huh.Field {
	applyAuthNavPosition(&f.keymap, p)
	return f
}

func (f *authModeField) GetKey() string { return f.key }
func (f *authModeField) GetValue() any {
	if f.mode == nil {
		return ConnectionMode("")
	}
	return *f.mode
}
