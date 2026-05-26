package wizard

import (
	"io"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
)

// credentialSwitchField 按认证方式只渲染密码或私钥其一（LayoutStack 下避免两项同屏）。
type credentialSwitchField struct {
	mode *ConnectionMode
	pw   *passwordTestField
	keyF *keyIdentityField

	fieldKey string
}

// NewCredentialSwitchField 创建凭证切换字段。
func NewCredentialSwitchField(mode *ConnectionMode, pw *passwordTestField, keyF *keyIdentityField) *credentialSwitchField {
	return &credentialSwitchField{mode: mode, pw: pw, keyF: keyF}
}

func (f *credentialSwitchField) active() huh.Field {
	if f.mode != nil && *f.mode == ModeKey {
		return f.keyF
	}
	return f.pw
}

func (f *credentialSwitchField) Key(k string) *credentialSwitchField {
	f.fieldKey = k
	f.pw.Key(k + "-pw")
	f.keyF.Key(k + "-key")
	return f
}

func (f *credentialSwitchField) Init() tea.Cmd { return f.active().Init() }

func (f *credentialSwitchField) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	m, cmd := f.active().Update(msg)
	switch a := m.(type) {
	case *passwordTestField:
		f.pw = a
	case *keyIdentityField:
		f.keyF = a
	}
	return f, cmd
}

func (f *credentialSwitchField) View() string { return f.active().View() }

func (f *credentialSwitchField) Error() error { return f.active().Error() }

func (f *credentialSwitchField) Skip() bool { return false }

func (f *credentialSwitchField) Zoom() bool { return false }

func (f *credentialSwitchField) Focus() tea.Cmd { return f.active().Focus() }

func (f *credentialSwitchField) Blur() tea.Cmd { return f.active().Blur() }

func (f *credentialSwitchField) KeyBinds() []key.Binding { return f.active().KeyBinds() }

func (f *credentialSwitchField) Run() error { return f.active().Run() }

func (f *credentialSwitchField) RunAccessible(w io.Writer, r io.Reader) error {
	return f.active().RunAccessible(w, r)
}

func (f *credentialSwitchField) WithTheme(theme *huh.Theme) huh.Field {
	f.pw.WithTheme(theme)
	f.keyF.WithTheme(theme)
	return f
}

func (f *credentialSwitchField) WithAccessible(accessible bool) huh.Field {
	f.pw.WithAccessible(accessible)
	f.keyF.WithAccessible(accessible)
	return f
}

func (f *credentialSwitchField) WithKeyMap(k *huh.KeyMap) huh.Field {
	km := &huh.KeyMap{Input: wizardCredentialKeyMap()}
	if k != nil {
		km.Input.AcceptSuggestion = k.Input.AcceptSuggestion
	}
	f.pw.WithKeyMap(km)
	f.keyF.WithKeyMap(km)
	return f
}

func (f *credentialSwitchField) WithWidth(width int) huh.Field {
	f.pw.WithWidth(width)
	f.keyF.WithWidth(width)
	return f
}

func (f *credentialSwitchField) WithHeight(height int) huh.Field {
	f.pw.WithHeight(height)
	f.keyF.WithHeight(height)
	return f
}

func (f *credentialSwitchField) WithPosition(p huh.FieldPosition) huh.Field {
	f.pw.WithPosition(p)
	f.keyF.WithPosition(p)
	return f
}

func (f *credentialSwitchField) GetKey() string { return f.fieldKey }

func (f *credentialSwitchField) GetValue() any { return f.active().GetValue() }
