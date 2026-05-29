package wizard

import (
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
)

// aliasField 在别名输入行内联展示校验错误；留空时根据主机地址实时预览并将生成别名。
type aliasField struct {
	baseField  // 嵌入共有逻辑
	configPath string
	hostName   *string
	inlineMsg  string
	onAdvance  func()
}

// NewAliasField 创建带内联冲突提示的 Host 别名字段。
func NewAliasField(configPath string, hostName *string) *aliasField {
	ti := newTextInput(128, "")

	return &aliasField{
		baseField: baseField{
			accessor:  &huh.EmbeddedAccessor[string]{},
			id:        nextBaseFieldID(),
			textinput: ti,
			keymap:    wizardInputKeyMap(),
		},
		configPath: configPath,
		hostName:   hostName,
	}
}

// --- Builder 方法（返回 *aliasField 支持链式调用） ---

func (f *aliasField) Value(v *string) *aliasField { f.setValue(v); return f }
func (f *aliasField) Key(k string) *aliasField    { f.setKey(k); return f }
func (f *aliasField) Title(t string) *aliasField  { f.setTitle(t); return f }

// OnAdvance 在别名校验通过并进入下一步时调用。
func (f *aliasField) OnAdvance(fn func()) *aliasField {
	f.onAdvance = fn
	return f
}

// --- 委托给 baseField 的通用方法 ---

func (f *aliasField) Init() tea.Cmd           { return f.baseField.Init() }
func (f *aliasField) Error() error            { return f.baseField.Error() }
func (f *aliasField) Skip() bool              { return f.baseField.Skip() }
func (f *aliasField) Zoom() bool              { return f.baseField.Zoom() }
func (f *aliasField) KeyBinds() []key.Binding { return f.baseField.KeyBinds() }
func (f *aliasField) Run() error              { return huh.Run(f) }
func (f *aliasField) RunAccessible(w io.Writer, r io.Reader) error {
	return f.baseField.RunAccessible(w, r)
}

func (f *aliasField) WithTheme(theme *huh.Theme) huh.Field { f.baseField.WithTheme(theme); return f }
func (f *aliasField) WithAccessible(a bool) huh.Field      { f.baseField.WithAccessible(a); return f }
func (f *aliasField) WithKeyMap(k *huh.KeyMap) huh.Field   { f.baseField.WithKeyMap(k); return f }
func (f *aliasField) WithWidth(w int) huh.Field            { f.baseField.WithWidth(w); return f }
func (f *aliasField) WithHeight(h int) huh.Field           { f.baseField.WithHeight(h); return f }
func (f *aliasField) WithPosition(p huh.FieldPosition) huh.Field {
	f.baseField.WithPosition(p)
	return f
}
func (f *aliasField) GetKey() string { return f.baseField.GetKey() }
func (f *aliasField) GetValue() any  { return f.baseField.GetValue() }

// --- aliasField 自己的差异逻辑 ---

// Focus 覆盖 baseField.Focus，同步 placeholder。
func (f *aliasField) Focus() tea.Cmd {
	f.focused = true
	f.syncPlaceholder()
	return f.textinput.Focus()
}

// Blur 覆盖 baseField.Blur，空值时自动填充别名。
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
	f.applyTextInputStyles(styles)

	var below []string
	if f.inlineMsg != "" {
		below = append(below, styles.ErrorMessage.Width(maxWidth).Render(f.inlineMsg))
	}
	return renderInlineField(f.width, f.height, styles, f.title, f.textinput.View(), below...)
}
