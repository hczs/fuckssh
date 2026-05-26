package wizard

import (
	"errors"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
)

// aliasField 在别名输入行内联展示校验错误；留空时根据主机地址实时预览并将生成别名。
type aliasField struct {
	configPath string
	hostName   *string

	accessor huh.Accessor[string]
	key      string
	id       int

	title string

	textinput textinput.Model
	inlineMsg string

	focused    bool
	accessible bool
	width      int
	height     int
	theme      *huh.Theme
	keymap     huh.InputKeyMap

	onAdvance func()
}

var aliasFieldIDSeq int

func nextAliasFieldID() int {
	aliasFieldIDSeq++
	return aliasFieldIDSeq
}

// NewAliasField 创建带内联冲突提示的 Host 别名字段。
func NewAliasField(configPath string, hostName *string) *aliasField {
	ti := textinput.New()
	ti.CharLimit = 128

	return &aliasField{
		configPath: configPath,
		hostName:   hostName,
		accessor:   &huh.EmbeddedAccessor[string]{},
		id:         nextAliasFieldID(),
		textinput:  ti,
		keymap:     wizardInputKeyMap(),
	}
}

func (f *aliasField) Value(v *string) *aliasField {
	f.accessor = huh.NewPointerAccessor(v)
	f.textinput.SetValue(f.accessor.Get())
	return f
}

func (f *aliasField) Key(k string) *aliasField {
	f.key = k
	return f
}

func (f *aliasField) Title(title string) *aliasField {
	f.title = title
	return f
}

// OnAdvance 在别名校验通过并进入下一步时调用。
func (f *aliasField) OnAdvance(fn func()) *aliasField {
	f.onAdvance = fn
	return f
}

func (f *aliasField) activeStyles() *huh.FieldStyles {
	if f.theme == nil {
		f.theme = huh.ThemeCharm()
	}
	if f.focused {
		return &f.theme.Focused
	}
	return &f.theme.Blurred
}

func (f *aliasField) syncPlaceholder() {
	host := ""
	if f.hostName != nil {
		host = *f.hostName
	}
	f.textinput.Placeholder = aliasPlaceholder(host)
}

func (f *aliasField) validate(raw string) error {
	return aliasFieldValidate(f.configPath, f.hostName)(raw)
}

func (f *aliasField) Init() tea.Cmd { return nil }

func (f *aliasField) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, f.keymap.Prev):
			return f, huh.PrevField
		case key.Matches(msg, f.keymap.Next, f.keymap.Submit):
			raw := f.textinput.Value()
			if err := f.validate(raw); err != nil {
				f.inlineMsg = err.Error()
				return f, nil
			}
			f.inlineMsg = ""
			host := strings.TrimSpace(*f.hostName)
			normalized := resolveAliasCandidate(raw, host)
			f.textinput.SetValue(normalized)
			f.accessor.Set(normalized)
			if f.onAdvance != nil {
				f.onAdvance()
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

func (f *aliasField) View() string {
	styles := f.activeStyles()
	frame := styles.Base.GetHorizontalFrameSize()
	maxWidth := f.width - frame

	f.syncPlaceholder()
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

func (f *aliasField) Error() error { return nil }

func (f *aliasField) Skip() bool { return false }
func (f *aliasField) Zoom() bool { return false }

func (f *aliasField) Focus() tea.Cmd {
	f.focused = true
	f.syncPlaceholder()
	return f.textinput.Focus()
}

func (f *aliasField) Blur() tea.Cmd {
	f.focused = false
	f.textinput.Blur()
	raw := f.textinput.Value()
	if strings.TrimSpace(raw) == "" {
		host := strings.TrimSpace(*f.hostName)
		if gen := resolveAliasCandidate("", host); gen != "" {
			f.accessor.Set(gen)
		} else {
			f.accessor.Set(raw)
		}
	} else {
		f.accessor.Set(raw)
	}
	return nil
}

func (f *aliasField) KeyBinds() []key.Binding {
	return []key.Binding{f.keymap.Prev, f.keymap.Submit, f.keymap.Next}
}

func (f *aliasField) Run() error { return huh.Run(f) }

func (f *aliasField) RunAccessible(w io.Writer, r io.Reader) error {
	_ = w
	_ = r
	return errors.New("alias field: accessible mode not supported")
}

func (f *aliasField) WithTheme(theme *huh.Theme) huh.Field {
	if theme != nil && f.theme == nil {
		f.theme = theme
	}
	return f
}

func (f *aliasField) WithAccessible(accessible bool) huh.Field {
	f.accessible = accessible
	return f
}

func (f *aliasField) WithKeyMap(k *huh.KeyMap) huh.Field {
	if k != nil {
		f.keymap = k.Input
		f.textinput.KeyMap.AcceptSuggestion = f.keymap.AcceptSuggestion
	}
	return f
}

func (f *aliasField) WithWidth(width int) huh.Field {
	f.width = width
	setInlineInputWidth(width, f.activeStyles(), f.title, &f.textinput)
	return f
}

func (f *aliasField) WithHeight(height int) huh.Field {
	f.height = height
	return f
}

func (f *aliasField) WithPosition(p huh.FieldPosition) huh.Field {
	applyInputNavPosition(&f.keymap, p)
	return f
}

func (f *aliasField) GetKey() string { return f.key }
func (f *aliasField) GetValue() any  { return f.accessor.Get() }
