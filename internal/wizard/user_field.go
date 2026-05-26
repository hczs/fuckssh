package wizard

import (
	"errors"
	"io"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
)

const defaultSSHUser = "root"

// userField SSH 用户名；留空回车时使用 root（与占位符提示一致）。
type userField struct {
	accessor  huh.Accessor[string]
	onAdvance func()

	key   string
	id    int
	title string

	textinput textinput.Model

	focused    bool
	accessible bool
	width      int
	height     int
	theme      *huh.Theme
	keymap     huh.InputKeyMap
}

var userFieldIDSeq int

func nextUserFieldID() int {
	userFieldIDSeq++
	return userFieldIDSeq
}

// NewUserField 创建 SSH 用户名字段。
func NewUserField(onAdvance func()) *userField {
	ti := textinput.New()
	ti.CharLimit = 64
	ti.Placeholder = defaultSSHUser

	return &userField{
		onAdvance: onAdvance,
		accessor:  &huh.EmbeddedAccessor[string]{},
		id:        nextUserFieldID(),
		textinput: ti,
		keymap:    wizardInputKeyMap(),
	}
}

func (f *userField) Value(v *string) *userField {
	f.accessor = huh.NewPointerAccessor(v)
	f.textinput.SetValue(f.accessor.Get())
	return f
}

func (f *userField) Key(k string) *userField {
	f.key = k
	return f
}

func (f *userField) Title(title string) *userField {
	f.title = title
	return f
}

func (f *userField) activeStyles() *huh.FieldStyles {
	if f.theme == nil {
		f.theme = WizardTheme()
	}
	if f.focused {
		return &f.theme.Focused
	}
	return &f.theme.Blurred
}

func (f *userField) commit(raw string) {
	user := effectiveUser(raw)
	f.textinput.SetValue(user)
	f.accessor.Set(user)
}

func (f *userField) Init() tea.Cmd { return nil }

func (f *userField) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, f.keymap.Prev):
			return f, huh.PrevField
		case key.Matches(msg, f.keymap.Next, f.keymap.Submit):
			f.commit(f.textinput.Value())
			if f.onAdvance != nil {
				f.onAdvance()
			}
			return f, huh.NextField
		default:
			var cmd tea.Cmd
			f.textinput, cmd = f.textinput.Update(msg)
			f.accessor.Set(f.textinput.Value())
			return f, cmd
		}
	default:
		if !f.focused {
			return f, nil
		}
		var cmd tea.Cmd
		f.textinput, cmd = f.textinput.Update(msg)
		f.accessor.Set(f.textinput.Value())
		return f, cmd
	}
}

func (f *userField) View() string {
	styles := f.activeStyles()
	f.textinput.PlaceholderStyle = styles.TextInput.Placeholder
	f.textinput.PromptStyle = styles.TextInput.Prompt
	f.textinput.Cursor.Style = styles.TextInput.Cursor
	f.textinput.Cursor.TextStyle = styles.TextInput.CursorText
	f.textinput.TextStyle = styles.TextInput.Text
	return renderInlineField(f.width, f.height, styles, f.title, f.textinput.View())
}

func (f *userField) Error() error { return nil }
func (f *userField) Skip() bool   { return false }
func (f *userField) Zoom() bool   { return false }

func (f *userField) Focus() tea.Cmd {
	f.focused = true
	return f.textinput.Focus()
}

func (f *userField) Blur() tea.Cmd {
	f.focused = false
	f.textinput.Blur()
	f.commit(f.textinput.Value())
	return nil
}

func (f *userField) KeyBinds() []key.Binding {
	return []key.Binding{f.keymap.Prev, f.keymap.Submit, f.keymap.Next}
}

func (f *userField) Run() error { return huh.Run(f) }

func (f *userField) RunAccessible(w io.Writer, r io.Reader) error {
	_ = w
	_ = r
	return errors.New("user field: accessible mode not supported")
}

func (f *userField) WithTheme(theme *huh.Theme) huh.Field {
	if theme != nil && f.theme == nil {
		f.theme = theme
	}
	return f
}

func (f *userField) WithAccessible(accessible bool) huh.Field {
	f.accessible = accessible
	return f
}

func (f *userField) WithKeyMap(k *huh.KeyMap) huh.Field {
	if k != nil {
		f.keymap = k.Input
		f.textinput.KeyMap.AcceptSuggestion = f.keymap.AcceptSuggestion
	}
	return f
}

func (f *userField) WithWidth(width int) huh.Field {
	f.width = width
	setInlineInputWidth(width, f.activeStyles(), f.title, &f.textinput)
	return f
}

func (f *userField) WithHeight(height int) huh.Field {
	f.height = height
	return f
}

func (f *userField) WithPosition(p huh.FieldPosition) huh.Field {
	applyInputNavPosition(&f.keymap, p)
	return f
}

func (f *userField) GetKey() string { return f.key }
func (f *userField) GetValue() any  { return f.accessor.Get() }
