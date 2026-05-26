package wizard

import (
	"errors"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/fuckssh/fuckssh/internal/i18n"
)

// hostField 主机地址输入，空值时内联展示「不能为空」。
type hostField struct {
	accessor huh.Accessor[string]
	onValid  func()

	key   string
	id    int
	title string

	textinput textinput.Model
	inlineMsg string

	focused    bool
	accessible bool
	width      int
	height     int
	theme      *huh.Theme
	keymap     huh.InputKeyMap
}

var hostFieldIDSeq int

func nextHostFieldID() int {
	hostFieldIDSeq++
	return hostFieldIDSeq
}

// NewHostField 创建主机地址字段。
func NewHostField(onValid func()) *hostField {
	ti := textinput.New()
	ti.CharLimit = 256
	ti.Placeholder = hostInputPlaceholder()

	return &hostField{
		onValid:  onValid,
		accessor: &huh.EmbeddedAccessor[string]{},
		id:       nextHostFieldID(),
		textinput: ti,
		keymap:    wizardInputKeyMap(),
	}
}

func (f *hostField) Value(v *string) *hostField {
	f.accessor = huh.NewPointerAccessor(v)
	f.textinput.SetValue(f.accessor.Get())
	return f
}

func (f *hostField) Key(k string) *hostField {
	f.key = k
	return f
}

func (f *hostField) Title(title string) *hostField {
	f.title = title
	return f
}

func (f *hostField) activeStyles() *huh.FieldStyles {
	if f.theme == nil {
		f.theme = WizardTheme()
	}
	if f.focused {
		return &f.theme.Focused
	}
	return &f.theme.Blurred
}

func (f *hostField) validate(raw string) error {
	return nonEmpty(i18n.T(i18n.KeyWizardErrEmpty))(raw)
}

func (f *hostField) Init() tea.Cmd { return nil }

func (f *hostField) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, f.keymap.Prev):
			return f, huh.PrevField
		case key.Matches(msg, f.keymap.Next, f.keymap.Submit):
			raw := strings.TrimSpace(f.textinput.Value())
			if err := f.validate(raw); err != nil {
				f.inlineMsg = err.Error()
				return f, nil
			}
			f.inlineMsg = ""
			f.accessor.Set(raw)
			if f.onValid != nil {
				f.onValid()
			}
			return f, huh.NextField
		default:
			f.inlineMsg = ""
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

func (f *hostField) View() string {
	styles := f.activeStyles()
	frame := styles.Base.GetHorizontalFrameSize()
	maxWidth := f.width - frame

	f.textinput.PlaceholderStyle = styles.TextInput.Placeholder
	f.textinput.PromptStyle = styles.TextInput.Prompt
	f.textinput.Cursor.Style = styles.TextInput.Cursor
	f.textinput.Cursor.TextStyle = styles.TextInput.CursorText
	f.textinput.TextStyle = styles.TextInput.Text

	var below []string
	if f.inlineMsg != "" {
		below = append(below, styles.ErrorMessage.Width(maxWidth).Render(f.inlineMsg))
	}
	return renderInlineField(f.width, f.height, styles, f.title, f.textinput.View(), below...)
}

func (f *hostField) Error() error { return nil }
func (f *hostField) Skip() bool   { return false }
func (f *hostField) Zoom() bool   { return false }

func (f *hostField) Focus() tea.Cmd {
	f.focused = true
	return f.textinput.Focus()
}

func (f *hostField) Blur() tea.Cmd {
	f.focused = false
	f.textinput.Blur()
	f.accessor.Set(strings.TrimSpace(f.textinput.Value()))
	return nil
}

func (f *hostField) KeyBinds() []key.Binding {
	return []key.Binding{f.keymap.Prev, f.keymap.Submit, f.keymap.Next}
}

func (f *hostField) Run() error { return huh.Run(f) }

func (f *hostField) RunAccessible(w io.Writer, r io.Reader) error {
	_ = w
	_ = r
	return errors.New("host field: accessible mode not supported")
}

func (f *hostField) WithTheme(theme *huh.Theme) huh.Field {
	if theme != nil && f.theme == nil {
		f.theme = theme
	}
	return f
}

func (f *hostField) WithAccessible(accessible bool) huh.Field {
	f.accessible = accessible
	return f
}

func (f *hostField) WithKeyMap(k *huh.KeyMap) huh.Field {
	if k != nil {
		f.keymap = k.Input
		f.textinput.KeyMap.AcceptSuggestion = f.keymap.AcceptSuggestion
	}
	return f
}

func (f *hostField) WithWidth(width int) huh.Field {
	f.width = width
	setInlineInputWidth(width, f.activeStyles(), f.title, &f.textinput)
	return f
}

func (f *hostField) WithHeight(height int) huh.Field {
	f.height = height
	return f
}

func (f *hostField) WithPosition(p huh.FieldPosition) huh.Field {
	applyInputNavPosition(&f.keymap, p)
	return f
}

func (f *hostField) GetKey() string { return f.key }
func (f *hostField) GetValue() any  { return f.accessor.Get() }
