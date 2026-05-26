package wizard

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/fuckssh/fuckssh/internal/i18n"
	"github.com/fuckssh/fuckssh/internal/platform"
)

type keyIDState int

const (
	keyIDStateEdit keyIDState = iota
	keyIDStateTesting
	keyIDStateOK
	keyIDStateFail
)

type keyIDDoneMsg struct {
	err     error
	elapsed time.Duration
}

// keyIdentityField 在私钥路径输入行尾内联展示测连 spinner / 成功 / 失败状态（与密码字段一致）。
type keyIdentityField struct {
	ctx      context.Context
	in       *KeyModeInput
	testAuth keyAuthTestFn
	onOK     func()
	onFail   func()

	accessor huh.Accessor[string]
	key      string
	id       int

	title       string
	description string

	textinput textinput.Model
	spinner   spinner.Model

	state     keyIDState
	elapsed   time.Duration
	inlineMsg string

	focused    bool
	accessible bool
	width      int
	height     int
	theme      *huh.Theme
	keymap     huh.InputKeyMap
}

var keyIdentityFieldIDSeq int

func nextKeyIdentityFieldID() int {
	keyIdentityFieldIDSeq++
	return keyIdentityFieldIDSeq
}

// NewKeyIdentityField 创建带内联测连反馈的私钥路径字段。
func NewKeyIdentityField(ctx context.Context, in *KeyModeInput, testAuth keyAuthTestFn, onOK, onFail func()) *keyIdentityField {
	ti := textinput.New()
	ti.CharLimit = 512
	ti.Placeholder = i18n.T(i18n.KeyWizardPasswordTestDesc)

	sp := spinner.New()
	sp.Spinner = spinner.Dot

	return &keyIdentityField{
		ctx:       ctx,
		in:        in,
		testAuth:  testAuth,
		onOK:      onOK,
		onFail:    onFail,
		accessor:  &huh.EmbeddedAccessor[string]{},
		id:        nextKeyIdentityFieldID(),
		textinput: ti,
		spinner:   sp,
		state:     keyIDStateEdit,
		keymap:    wizardCredentialKeyMap(),
	}
}

func (f *keyIdentityField) Value(v *string) *keyIdentityField {
	f.accessor = huh.NewPointerAccessor(v)
	f.textinput.SetValue(f.accessor.Get())
	return f
}

func (f *keyIdentityField) Key(k string) *keyIdentityField {
	f.key = k
	return f
}

func (f *keyIdentityField) Title(title string) *keyIdentityField {
	f.title = title
	return f
}

func (f *keyIdentityField) Description(desc string) *keyIdentityField {
	f.description = desc
	return f
}

func (f *keyIdentityField) activeStyles() *huh.FieldStyles {
	if f.theme == nil {
		f.theme = WizardTheme()
	}
	if f.focused && f.state != keyIDStateTesting && f.state != keyIDStateOK {
		return &f.theme.Focused
	}
	return &f.theme.Blurred
}

func (f *keyIdentityField) statusStyle() lipgloss.Style {
	if f.theme == nil {
		return lipgloss.NewStyle()
	}
	switch f.state {
	case keyIDStateOK:
		return wizardStatusOKStyle(f.theme)
	case keyIDStateFail:
		return f.theme.Focused.ErrorMessage
	default:
		return f.theme.Blurred.Description
	}
}

func (f *keyIdentityField) statusText() string {
	switch f.state {
	case keyIDStateTesting:
		return i18n.T(i18n.KeyWizardTestingConnInline)
	case keyIDStateOK:
		return i18n.T(i18n.KeyWizardConnOKMs, f.elapsed.Milliseconds())
	case keyIDStateFail:
		if f.inlineMsg != "" {
			return f.inlineMsg
		}
		return i18n.T(i18n.KeyWizardKeyAuthFailed)
	default:
		return ""
	}
}

func (f *keyIdentityField) renderInputRow(styles *huh.FieldStyles) string {
	line := f.textinput.View()
	status := f.statusText()
	if status == "" {
		return line
	}
	var suffix string
	if f.state == keyIDStateTesting {
		suffix = f.spinner.View() + " " + f.statusStyle().Render(status)
	} else {
		suffix = f.statusStyle().Render(status)
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, line, " ", suffix)
}

func (f *keyIdentityField) preparePath(path string) error {
	path = strings.TrimSpace(path)
	if path == "" {
		return errors.New(i18n.T(i18n.KeyWizardErrEmpty))
	}
	expanded, err := platform.ExpandPath(path)
	if err != nil {
		return err
	}
	if _, err := os.Stat(expanded); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("%s: %s", i18n.T(i18n.KeyWizardErrKeyMissing), expanded)
		}
		return fmt.Errorf("%s: %v", i18n.T(i18n.KeyWizardErrKeyRead), err)
	}
	f.in.IdentityFile = expanded
	f.accessor.Set(expanded)
	f.textinput.SetValue(expanded)
	return nil
}

func (f *keyIdentityField) startTest(path string) tea.Cmd {
	if err := f.preparePath(path); err != nil {
		f.state = keyIDStateFail
		f.inlineMsg = err.Error()
		if f.onFail != nil {
			f.onFail()
		}
		f.textinput.Focus()
		f.focused = true
		return textinput.Blink
	}
	f.state = keyIDStateTesting
	f.inlineMsg = ""
	f.textinput.Blur()

	testCmd := func() tea.Msg {
		start := time.Now()
		err := f.testAuth(f.ctx, *f.in)
		return keyIDDoneMsg{err: err, elapsed: time.Since(start)}
	}
	return tea.Batch(f.spinner.Tick, testCmd)
}

func (f *keyIdentityField) handleTestDone(msg keyIDDoneMsg) (tea.Model, tea.Cmd) {
	// huh Group 会对同一消息 Update 两次，避免重复 NextField 直接提交表单。
	if f.state == keyIDStateOK {
		return f, nil
	}
	if msg.err != nil {
		f.state = keyIDStateFail
		f.inlineMsg = keyConnectionTestFailureMessage(msg.err)
		if f.onFail != nil {
			f.onFail()
		}
		f.textinput.SetValue("")
		f.accessor.Set("")
		f.in.IdentityFile = ""
		f.textinput.Focus()
		f.focused = true
		return f, textinput.Blink
	}
	f.state = keyIDStateOK
	f.elapsed = msg.elapsed
	f.inlineMsg = ""
	if f.onOK != nil {
		f.onOK()
	}
	// 测连成功后自动进入备注步，无需再按 Enter。
	return f, huh.NextField
}

// resumeEditFromSuccess 测连已通过时回到可编辑态，便于修改私钥路径后重新测试。
func (f *keyIdentityField) resumeEditFromSuccess() tea.Cmd {
	f.state = keyIDStateEdit
	f.elapsed = 0
	f.inlineMsg = ""
	if f.onFail != nil {
		f.onFail()
	}
	f.focused = true
	return f.textinput.Focus()
}

func (f *keyIdentityField) Init() tea.Cmd { return nil }

func (f *keyIdentityField) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case keyIDDoneMsg:
		return f.handleTestDone(msg)
	case spinner.TickMsg:
		if f.state == keyIDStateTesting {
			var cmd tea.Cmd
			f.spinner, cmd = f.spinner.Update(msg)
			return f, cmd
		}
	case tea.KeyMsg:
		if f.state == keyIDStateTesting {
			return f, nil
		}
		if f.state == keyIDStateOK {
			switch {
			case key.Matches(msg, f.keymap.Prev):
				return f, huh.PrevField
			case key.Matches(msg, f.keymap.Submit):
				return f, huh.NextField
			case key.Matches(msg, f.keymap.Next):
				return f, huh.NextField
			default:
				var cmds []tea.Cmd
				cmds = append(cmds, f.resumeEditFromSuccess())
				var cmd tea.Cmd
				f.textinput, cmd = f.textinput.Update(msg)
				f.accessor.Set(f.textinput.Value())
				cmds = append(cmds, cmd)
				return f, tea.Batch(cmds...)
			}
		}
		f.inlineMsg = ""
		switch {
		case key.Matches(msg, f.keymap.Prev):
			return f, huh.PrevField
		case key.Matches(msg, f.keymap.Submit):
			path := strings.TrimSpace(f.textinput.Value())
			if path == "" {
				f.state = keyIDStateFail
				f.inlineMsg = i18n.T(i18n.KeyWizardErrEmpty)
				return f, nil
			}
			return f, f.startTest(path)
		case key.Matches(msg, f.keymap.Next):
			return f, huh.NextField
		}
	}

	if f.state == keyIDStateEdit || f.state == keyIDStateFail {
		var cmd tea.Cmd
		f.textinput, cmd = f.textinput.Update(msg)
		f.accessor.Set(f.textinput.Value())
		cmds = append(cmds, cmd)
	}

	return f, tea.Batch(cmds...)
}

func (f *keyIdentityField) View() string {
	styles := f.activeStyles()
	f.textinput.PlaceholderStyle = styles.TextInput.Placeholder
	f.textinput.PromptStyle = styles.TextInput.Prompt
	f.textinput.Cursor.Style = styles.TextInput.Cursor
	f.textinput.Cursor.TextStyle = styles.TextInput.CursorText
	f.textinput.TextStyle = styles.TextInput.Text

	var below []string
	if f.state == keyIDStateEdit && f.description != "" {
		below = append(below, styles.Description.Render(f.description))
	}
	return renderInlineField(f.width, f.height, styles, f.title, f.renderInputRow(styles), below...)
}

func (f *keyIdentityField) Error() error { return nil }

func (f *keyIdentityField) Skip() bool { return false }
func (f *keyIdentityField) Zoom() bool { return false }

func (f *keyIdentityField) Focus() tea.Cmd {
	f.focused = true
	if f.state == keyIDStateEdit || f.state == keyIDStateFail {
		return f.textinput.Focus()
	}
	return nil
}

func (f *keyIdentityField) Blur() tea.Cmd {
	f.focused = false
	f.textinput.Blur()
	f.accessor.Set(f.textinput.Value())
	return nil
}

func (f *keyIdentityField) KeyBinds() []key.Binding {
	if f.state == keyIDStateTesting {
		return nil
	}
	if f.state == keyIDStateOK {
		return []key.Binding{f.keymap.Prev, f.keymap.Submit, f.keymap.Next}
	}
	return []key.Binding{f.keymap.Prev, f.keymap.Submit, f.keymap.Next}
}

func (f *keyIdentityField) Run() error { return huh.Run(f) }

func (f *keyIdentityField) RunAccessible(w io.Writer, r io.Reader) error {
	_ = w
	_ = r
	return errors.New("key identity field: accessible mode not supported")
}

func (f *keyIdentityField) WithTheme(theme *huh.Theme) huh.Field {
	if theme != nil && f.theme == nil {
		f.theme = theme
	}
	return f
}

func (f *keyIdentityField) WithAccessible(accessible bool) huh.Field {
	f.accessible = accessible
	return f
}

func (f *keyIdentityField) WithKeyMap(k *huh.KeyMap) huh.Field {
	if k != nil {
		f.keymap = k.Input
		f.textinput.KeyMap.AcceptSuggestion = f.keymap.AcceptSuggestion
	}
	return f
}

func (f *keyIdentityField) WithWidth(width int) huh.Field {
	f.width = width
	setInlineInputWidth(width, f.activeStyles(), f.title, &f.textinput)
	return f
}

func (f *keyIdentityField) WithHeight(height int) huh.Field {
	f.height = height
	return f
}

func (f *keyIdentityField) WithPosition(p huh.FieldPosition) huh.Field {
	applyCredentialNavPosition(&f.keymap, p)
	return f
}

func (f *keyIdentityField) GetKey() string { return f.key }
func (f *keyIdentityField) GetValue() any  { return f.accessor.Get() }
