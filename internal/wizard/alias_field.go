package wizard

import (
	"errors"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/fuckssh/fuckssh/internal/keys"
)

// aliasField 在别名输入行下方内联展示校验错误（与密码字段一致，避免 WithShowErrors(false) 吞掉提示）。
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
		keymap:     huh.NewDefaultKeyMap().Input,
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

func (f *aliasField) activeStyles() *huh.FieldStyles {
	if f.theme == nil {
		f.theme = huh.ThemeCharm()
	}
	if f.focused {
		return &f.theme.Focused
	}
	return &f.theme.Blurred
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
			normalized := keys.NormalizeHostAlias(raw)
			f.textinput.SetValue(normalized)
			f.accessor.Set(normalized)
			return f, huh.NextField
		default:
			// 用户修改输入时清除旧错误。
			f.inlineMsg = ""
			var cmd tea.Cmd
			f.textinput, cmd = f.textinput.Update(msg)
			f.accessor.Set(f.textinput.Value())
			return f, cmd
		}
	default:
		// huh 会在每次 Update 后发送 refresh 消息；不可在此清除 inlineMsg，否则冲突提示闪退。
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

	f.textinput.PlaceholderStyle = styles.TextInput.Placeholder
	f.textinput.PromptStyle = styles.TextInput.Prompt
	f.textinput.Cursor.Style = styles.TextInput.Cursor
	f.textinput.Cursor.TextStyle = styles.TextInput.CursorText
	f.textinput.TextStyle = styles.TextInput.Text

	var sb strings.Builder
	if f.title != "" {
		sb.WriteString(styles.Title.Render(f.title))
		sb.WriteString("\n")
	}
	if desc := aliasDescription(f.hostName); desc != "" {
		sb.WriteString(styles.Description.Render(desc))
		sb.WriteString("\n")
	}
	sb.WriteString(f.textinput.View())
	if f.inlineMsg != "" {
		sb.WriteString("\n")
		sb.WriteString(styles.ErrorMessage.Width(maxWidth).Render(f.inlineMsg))
	}

	return styles.Base.Width(f.width).Height(f.height).Render(sb.String())
}

func (f *aliasField) Error() error { return nil }

func (f *aliasField) Skip() bool { return false }
func (f *aliasField) Zoom() bool { return false }

func (f *aliasField) Focus() tea.Cmd {
	f.focused = true
	return f.textinput.Focus()
}

func (f *aliasField) Blur() tea.Cmd {
	f.focused = false
	f.textinput.Blur()
	f.accessor.Set(f.textinput.Value())
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
	styles := f.activeStyles()
	f.width = width
	frame := styles.Base.GetHorizontalFrameSize()
	promptW := lipgloss.Width(f.textinput.PromptStyle.Render(f.textinput.Prompt))
	f.textinput.Width = width - frame - promptW - 1
	if f.textinput.Width < 20 {
		f.textinput.Width = 20
	}
	return f
}

func (f *aliasField) WithHeight(height int) huh.Field {
	f.height = height
	return f
}

func (f *aliasField) WithPosition(p huh.FieldPosition) huh.Field {
	f.keymap.Prev.SetEnabled(!p.IsFirst())
	// 别名为表单最后一步；LayoutStack + reveal 时 IsLast 可能滞后，始终允许 Enter 提交校验。
	f.keymap.Next.SetEnabled(true)
	f.keymap.Submit.SetEnabled(true)
	return f
}

func (f *aliasField) GetKey() string { return f.key }
func (f *aliasField) GetValue() any  { return f.accessor.Get() }
