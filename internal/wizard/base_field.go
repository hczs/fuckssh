package wizard

// base_field.go 提取所有自定义字段的共有状态与 huh.Field 方法。
// 各具体字段通过 struct embedding 复用，只需覆盖少量差异化方法。
//
// 设计说明：
// - baseField 不实现 Update() 和 View()，这两个方法是每个字段的核心差异点
// - baseField 的 WithTheme/WithWidth 等方法返回 *baseField，不满足 huh.Field 接口
//   因此各具体类型需要重新定义这些方法，内部委托给 baseField 的辅助函数
// - 这种"辅助函数"模式比"方法继承"更灵活，允许具体类型覆盖任何行为

import (
	"errors"
	"io"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
)

// baseField 提取所有自定义字段的共有状态与 huh.Field 方法。
type baseField struct {
	accessor   huh.Accessor[string]
	key        string
	id         int
	title      string
	textinput  textinput.Model
	focused    bool
	accessible bool
	width      int
	height     int
	theme      *huh.Theme
	keymap     huh.InputKeyMap
}

// --- 通用 ID 生成器 ---

var baseFieldIDSeq int

func nextBaseFieldID() int {
	baseFieldIDSeq++
	return baseFieldIDSeq
}

// --- Builder 辅助方法（返回值由具体类型包装） ---

func (f *baseField) setKey(k string)   { f.key = k }
func (f *baseField) setTitle(t string) { f.title = t }
func (f *baseField) setValue(v *string) {
	f.accessor = huh.NewPointerAccessor(v)
	f.textinput.SetValue(f.accessor.Get())
}

// --- activeStyles 基础版（无状态机的简单字段用此版本） ---

func (f *baseField) activeStyles() *huh.FieldStyles {
	if f.theme == nil {
		f.theme = WizardTheme()
	}
	if f.focused {
		return &f.theme.Focused
	}
	return &f.theme.Blurred
}

// --- huh.Field 接口的通用实现 ---

func (f *baseField) Init() tea.Cmd { return nil }
func (f *baseField) Error() error  { return nil }
func (f *baseField) Skip() bool    { return false }
func (f *baseField) Zoom() bool    { return false }

func (f *baseField) Focus() tea.Cmd {
	f.focused = true
	return f.textinput.Focus()
}

func (f *baseField) Blur() tea.Cmd {
	f.focused = false
	f.textinput.Blur()
	f.accessor.Set(f.textinput.Value())
	return nil
}

func (f *baseField) KeyBinds() []key.Binding {
	return []key.Binding{f.keymap.Prev, f.keymap.Submit, f.keymap.Next}
}

// Run() 不在 baseField 中实现，因为 baseField 不是完整的 huh.Field。
// 各具体类型需要自行实现 Run()，通常直接调用 huh.Run(f)。

func (f *baseField) RunAccessible(w io.Writer, r io.Reader) error {
	return errors.New("accessible mode not supported")
}

func (f *baseField) WithTheme(theme *huh.Theme) *baseField {
	if theme != nil && f.theme == nil {
		f.theme = theme
	}
	return f
}

func (f *baseField) WithAccessible(accessible bool) *baseField {
	f.accessible = accessible
	return f
}

func (f *baseField) WithKeyMap(k *huh.KeyMap) *baseField {
	if k != nil {
		f.keymap = k.Input
		f.textinput.KeyMap.AcceptSuggestion = f.keymap.AcceptSuggestion
	}
	return f
}

func (f *baseField) WithWidth(width int) *baseField {
	f.width = width
	setInlineInputWidth(width, f.activeStyles(), f.title, &f.textinput)
	return f
}

func (f *baseField) WithHeight(height int) *baseField {
	f.height = height
	return f
}

func (f *baseField) WithPosition(p huh.FieldPosition) *baseField {
	applyInputNavPosition(&f.keymap, p)
	return f
}

func (f *baseField) GetKey() string { return f.key }
func (f *baseField) GetValue() any  { return f.accessor.Get() }

// --- 通用渲染辅助 ---

func (f *baseField) applyTextInputStyles(styles *huh.FieldStyles) {
	f.textinput.PlaceholderStyle = styles.TextInput.Placeholder
	f.textinput.PromptStyle = styles.TextInput.Prompt
	f.textinput.Cursor.Style = styles.TextInput.Cursor
	f.textinput.Cursor.TextStyle = styles.TextInput.CursorText
	f.textinput.TextStyle = styles.TextInput.Text
}

// newTextInput 创建标准 textinput.Model，统一设置 CharLimit 和 Placeholder。
func newTextInput(charLimit int, placeholder string) textinput.Model {
	ti := textinput.New()
	ti.CharLimit = charLimit
	ti.Placeholder = placeholder
	return ti
}
