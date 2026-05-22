package wizard

import (
	"context"
	"errors"
	"io"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/fuckssh/fuckssh/internal/i18n"
)

type pwTestState int

const (
	pwStateEdit pwTestState = iota
	pwStateTesting
	pwStateOK
	pwStateFail
)

type pwTestDoneMsg struct {
	err     error
	elapsed time.Duration
}

// passwordTestField 在密码输入行尾内联展示测连 spinner / 成功 / 失败状态。
type passwordTestField struct {
	ctx      context.Context
	in       *PasswordModeInput
	testAuth passwordAuthTestFn
	onOK     func()
	onFail   func()

	accessor huh.Accessor[string]
	key      string
	id       int

	title       string
	description string

	textinput textinput.Model
	spinner   spinner.Model

	state     pwTestState
	elapsed   time.Duration
	inlineMsg string // 行内状态文案（不通过 Error() 交给 huh，避免底部重复）

	focused  bool
	inline   bool
	accessible bool
	width    int
	height   int
	theme    *huh.Theme
	keymap   huh.InputKeyMap
}

var passwordFieldIDSeq int

func nextPasswordFieldID() int {
	passwordFieldIDSeq++
	return passwordFieldIDSeq
}

// NewPasswordTestField 创建带内联测连反馈的密码字段。
func NewPasswordTestField(ctx context.Context, in *PasswordModeInput, testAuth passwordAuthTestFn, onOK, onFail func()) *passwordTestField {
	ti := textinput.New()
	ti.EchoMode = textinput.EchoPassword
	ti.CharLimit = 256

	sp := spinner.New()
	sp.Spinner = spinner.Dot

	return &passwordTestField{
		ctx:         ctx,
		in:          in,
		testAuth:    testAuth,
		onOK:        onOK,
		onFail:      onFail,
		accessor:    &huh.EmbeddedAccessor[string]{},
		id:          nextPasswordFieldID(),
		textinput:   ti,
		spinner:     sp,
		state:       pwStateEdit,
		description: i18n.T(i18n.KeyWizardHostKeyHint),
		keymap:      huh.NewDefaultKeyMap().Input,
	}
}

func (f *passwordTestField) Value(v *string) *passwordTestField {
	f.accessor = huh.NewPointerAccessor(v)
	f.textinput.SetValue(f.accessor.Get())
	return f
}

func (f *passwordTestField) Key(k string) *passwordTestField {
	f.key = k
	return f
}

func (f *passwordTestField) Title(title string) *passwordTestField {
	f.title = title
	return f
}

func (f *passwordTestField) activeStyles() *huh.FieldStyles {
	if f.theme == nil {
		f.theme = huh.ThemeCharm()
	}
	if f.focused && f.state != pwStateTesting && f.state != pwStateOK {
		return &f.theme.Focused
	}
	return &f.theme.Blurred
}

func (f *passwordTestField) statusStyle() lipgloss.Style {
	if f.theme == nil {
		return lipgloss.NewStyle()
	}
	switch f.state {
	case pwStateOK:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	case pwStateFail:
		return f.theme.Focused.ErrorMessage
	default:
		return f.theme.Blurred.Description
	}
}

func (f *passwordTestField) statusText() string {
	switch f.state {
	case pwStateTesting:
		return i18n.T(i18n.KeyWizardTestingConnInline)
	case pwStateOK:
		return i18n.T(i18n.KeyWizardConnOKMs, f.elapsed.Milliseconds())
	case pwStateFail:
		if f.inlineMsg != "" {
			return f.inlineMsg
		}
		return i18n.T(i18n.KeyWizardConnFailInline)
	default:
		return ""
	}
}

func (f *passwordTestField) renderInputRow(styles *huh.FieldStyles) string {
	line := f.textinput.View()
	status := f.statusText()
	if status == "" {
		return line
	}
	var suffix string
	if f.state == pwStateTesting {
		suffix = f.spinner.View() + " " + f.statusStyle().Render(status)
	} else {
		suffix = f.statusStyle().Render(status)
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, line, " ", suffix)
}

func (f *passwordTestField) startTest(password string) tea.Cmd {
	f.in.Password = strings.TrimSpace(password)
	f.accessor.Set(f.in.Password)
	f.state = pwStateTesting
	f.inlineMsg = ""
	f.textinput.Blur()

	testCmd := func() tea.Msg {
		elapsed, err := testPasswordConnection(f.ctx, f.in, password, f.testAuth)
		return pwTestDoneMsg{err: err, elapsed: elapsed}
	}
	return tea.Batch(f.spinner.Tick, testCmd)
}

func (f *passwordTestField) handleTestDone(msg pwTestDoneMsg) (tea.Model, tea.Cmd) {
	if msg.err != nil {
		f.state = pwStateFail
		f.inlineMsg = connectionTestFailureMessage(msg.err)
		if f.onFail != nil {
			f.onFail()
		}
		f.textinput.SetValue("")
		f.accessor.Set("")
		f.in.Password = ""
		f.textinput.Focus()
		f.focused = true
		return f, textinput.Blink
	}
	f.state = pwStateOK
	f.elapsed = msg.elapsed
	f.inlineMsg = ""
	if f.onOK != nil {
		f.onOK()
	}
	return f, nil
}

// Init implements huh.Field.
func (f *passwordTestField) Init() tea.Cmd {
	return nil
}

// Update implements huh.Field.
func (f *passwordTestField) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case pwTestDoneMsg:
		return f.handleTestDone(msg)
	case spinner.TickMsg:
		if f.state == pwStateTesting {
			var cmd tea.Cmd
			f.spinner, cmd = f.spinner.Update(msg)
			return f, cmd
		}
	case tea.KeyMsg:
		if f.state == pwStateTesting {
			return f, nil
		}
		if f.state == pwStateOK {
			switch {
			case key.Matches(msg, f.keymap.Prev):
				return f, huh.PrevField
			case key.Matches(msg, f.keymap.Next, f.keymap.Submit):
				return f, huh.NextField
			}
			return f, nil
		}
		f.inlineMsg = ""
		switch {
		case key.Matches(msg, f.keymap.Prev):
			return f, huh.PrevField
		case key.Matches(msg, f.keymap.Next, f.keymap.Submit):
			pwd := strings.TrimSpace(f.textinput.Value())
			if pwd == "" {
				f.state = pwStateFail
				f.inlineMsg = i18n.T(i18n.KeyWizardErrEmpty)
				return f, nil
			}
			return f, f.startTest(pwd)
		}
	}

	if f.state == pwStateEdit || f.state == pwStateFail {
		var cmd tea.Cmd
		f.textinput, cmd = f.textinput.Update(msg)
		f.accessor.Set(f.textinput.Value())
		cmds = append(cmds, cmd)
	}

	return f, tea.Batch(cmds...)
}

// View implements huh.Field.
func (f *passwordTestField) View() string {
	styles := f.activeStyles()
	maxWidth := f.width - styles.Base.GetHorizontalFrameSize()

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
	switch f.state {
	case pwStateEdit:
		if f.description != "" {
			sb.WriteString(styles.Description.Render(f.description))
			sb.WriteString("\n")
		}
	case pwStateOK:
		sb.WriteString(styles.Description.Render(i18n.T(i18n.KeyWizardConnOKContinue)))
		sb.WriteString("\n")
	}
	sb.WriteString(f.renderInputRow(styles))

	_ = maxWidth
	return styles.Base.Width(f.width).Height(f.height).Render(sb.String())
}

func (f *passwordTestField) Error() error {
	// 失败文案仅在输入行尾展示，避免 huh 在表单底部再渲染一行 * 错误。
	return nil
}

func (f *passwordTestField) Skip() bool  { return false }
func (f *passwordTestField) Zoom() bool { return false }

func (f *passwordTestField) Focus() tea.Cmd {
	f.focused = true
	if f.state == pwStateEdit || f.state == pwStateFail {
		return f.textinput.Focus()
	}
	// 测连成功后本字段仍保持焦点，等待 Enter 进入别名步。
	return nil
}

func (f *passwordTestField) Blur() tea.Cmd {
	f.focused = false
	f.textinput.Blur()
	f.accessor.Set(f.textinput.Value())
	return nil
}

func (f *passwordTestField) KeyBinds() []key.Binding {
	if f.state == pwStateTesting {
		return nil
	}
	if f.state == pwStateOK {
		return []key.Binding{f.keymap.Prev, f.keymap.Submit, f.keymap.Next}
	}
	return []key.Binding{f.keymap.Prev, f.keymap.Submit, f.keymap.Next}
}

func (f *passwordTestField) Run() error {
	return huh.Run(f)
}

func (f *passwordTestField) RunAccessible(w io.Writer, r io.Reader) error {
	_ = w
	_ = r
	return errors.New("password test field: accessible mode not supported")
}

func (f *passwordTestField) WithTheme(theme *huh.Theme) huh.Field {
	if theme != nil && f.theme == nil {
		f.theme = theme
	}
	return f
}

func (f *passwordTestField) WithAccessible(accessible bool) huh.Field {
	f.accessible = accessible
	return f
}

func (f *passwordTestField) WithKeyMap(k *huh.KeyMap) huh.Field {
	if k != nil {
		f.keymap = k.Input
		f.textinput.KeyMap.AcceptSuggestion = f.keymap.AcceptSuggestion
	}
	return f
}

func (f *passwordTestField) WithWidth(width int) huh.Field {
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

func (f *passwordTestField) WithHeight(height int) huh.Field {
	f.height = height
	return f
}

func (f *passwordTestField) WithPosition(p huh.FieldPosition) huh.Field {
	f.keymap.Prev.SetEnabled(!p.IsFirst())
	f.keymap.Next.SetEnabled(!p.IsLast())
	f.keymap.Submit.SetEnabled(p.IsLast())
	return f
}

func (f *passwordTestField) GetKey() string   { return f.key }
func (f *passwordTestField) GetValue() any    { return f.accessor.Get() }
