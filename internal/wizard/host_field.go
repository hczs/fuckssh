package wizard

import (
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/fuckssh/fuckssh/internal/i18n"
)

// hostField 主机地址输入，空值时内联展示「不能为空」。
type hostField struct {
	baseField // 嵌入共有逻辑
	onValid   func()
	inlineMsg string
}

// NewHostField 创建主机地址字段。
func NewHostField(onValid func()) *hostField {
	ti := newTextInput(256, hostInputPlaceholder())

	return &hostField{
		baseField: baseField{
			accessor:  &huh.EmbeddedAccessor[string]{},
			id:        nextBaseFieldID(),
			textinput: ti,
			keymap:    wizardInputKeyMap(),
		},
		onValid: onValid,
	}
}

// --- Builder 方法（返回 *hostField 支持链式调用） ---

func (f *hostField) Value(v *string) *hostField { f.setValue(v); return f }
func (f *hostField) Key(k string) *hostField    { f.setKey(k); return f }
func (f *hostField) Title(t string) *hostField  { f.setTitle(t); return f }

// --- 委托给 baseField 的通用方法 ---

func (f *hostField) Init() tea.Cmd           { return f.baseField.Init() }
func (f *hostField) Error() error            { return f.baseField.Error() }
func (f *hostField) Skip() bool              { return f.baseField.Skip() }
func (f *hostField) Zoom() bool              { return f.baseField.Zoom() }
func (f *hostField) Focus() tea.Cmd          { return f.baseField.Focus() }
func (f *hostField) KeyBinds() []key.Binding { return f.baseField.KeyBinds() }
func (f *hostField) Run() error              { return huh.Run(f) }
func (f *hostField) RunAccessible(w io.Writer, r io.Reader) error {
	return f.baseField.RunAccessible(w, r)
}

func (f *hostField) WithTheme(theme *huh.Theme) huh.Field { f.baseField.WithTheme(theme); return f }
func (f *hostField) WithAccessible(a bool) huh.Field      { f.baseField.WithAccessible(a); return f }
func (f *hostField) WithKeyMap(k *huh.KeyMap) huh.Field   { f.baseField.WithKeyMap(k); return f }
func (f *hostField) WithWidth(w int) huh.Field            { f.baseField.WithWidth(w); return f }
func (f *hostField) WithHeight(h int) huh.Field           { f.baseField.WithHeight(h); return f }
func (f *hostField) WithPosition(p huh.FieldPosition) huh.Field {
	f.baseField.WithPosition(p)
	return f
}
func (f *hostField) GetKey() string { return f.baseField.GetKey() }
func (f *hostField) GetValue() any  { return f.baseField.GetValue() }

// --- hostField 自己的差异逻辑 ---

// Blur 覆盖 baseField.Blur，对输入值做 TrimSpace 处理。
func (f *hostField) Blur() tea.Cmd {
	f.focused = false
	f.textinput.Blur()
	f.accessor.Set(strings.TrimSpace(f.textinput.Value()))
	return nil
}

func (f *hostField) validate(raw string) error {
	return nonEmpty(i18n.T(i18n.KeyWizardErrEmpty))(raw)
}

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

	f.applyTextInputStyles(styles)

	var below []string
	if f.inlineMsg != "" {
		below = append(below, styles.ErrorMessage.Width(maxWidth).Render(f.inlineMsg))
	}
	return renderInlineField(f.width, f.height, styles, f.title, f.textinput.View(), below...)
}
