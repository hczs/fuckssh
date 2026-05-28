package wizard

// connection_test_field.go 合并密码与密钥两种测连字段为统一的 connectionTestField。
//
// 设计说明：
// - passwordTestField 和 keyIdentityField 有 60%+ 的重复代码
// - 通过 testFieldStrategy 接口封装差异行为（验证、测连、清空凭证）
// - 统一状态机 testState 取代原来的 pwTestState / keyIDState
// - 构造函数 NewPasswordTestField / NewKeyIdentityField 保持向后兼容

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

// testFieldStrategy 封装密码/密钥两种测连字段的行为差异。
type testFieldStrategy interface {
	// validateInput 验证用户输入（密码：trimSpace；密钥：expandPath + stat）
	// 返回处理后的值和可能的错误。
	validateInput(raw string) (string, error)

	// startTestCmd 根据已验证的输入发起测连命令。
	startTestCmd(ctx context.Context, input string) tea.Cmd

	// clearCredential 测连失败后清空对应凭证字段。
	clearCredential()

	// failureDefaultMsg 测连失败且无具体错误时的默认行内文案。
	failureDefaultMsg() string

	// hasDescription View 时是否在 edit 状态显示 description 区。
	hasDescription() bool
}

// testState 统一的测连状态机。
type testState int

const (
	testStateEdit    testState = iota // 正在编辑
	testStateTesting                  // 正在测连
	testStateOK                       // 测连成功
	testStateFail                     // 测连失败
)

// testDoneMsg 统一的测连完成消息。
type testDoneMsg struct {
	err     error
	elapsed time.Duration
}

// connectionTestField 统一的测连字段，密码与密钥共用。
type connectionTestField struct {
	baseField   // 嵌入共有逻辑
	strategy    testFieldStrategy
	ctx         context.Context
	onOK        func()
	onFail      func()
	spinner     spinner.Model
	state       testState
	elapsed     time.Duration
	inlineMsg   string
	description string
}

// --- 构造函数（保持向后兼容） ---

// NewPasswordTestField 创建密码测连字段。
func NewPasswordTestField(ctx context.Context, in *PasswordModeInput, testAuth passwordAuthTestFn, onOK, onFail func()) *connectionTestField {
	ti := textinput.New()
	ti.EchoMode = textinput.EchoPassword
	ti.CharLimit = 256
	ti.Placeholder = i18n.T(i18n.KeyWizardPasswordTestDesc)

	sp := spinner.New()
	sp.Spinner = spinner.Dot

	return &connectionTestField{
		baseField: baseField{
			accessor:  &huh.EmbeddedAccessor[string]{},
			id:        nextBaseFieldID(),
			textinput: ti,
			keymap:    wizardCredentialKeyMap(),
		},
		strategy: &passwordTestStrategy{in: in, testAuth: testAuth},
		ctx:      ctx,
		onOK:     onOK,
		onFail:   onFail,
		spinner:  sp,
		state:    testStateEdit,
	}
}

// NewKeyIdentityField 创建密钥测连字段。
func NewKeyIdentityField(ctx context.Context, in *KeyModeInput, testAuth keyAuthTestFn, onOK, onFail func()) *connectionTestField {
	ti := textinput.New()
	ti.CharLimit = 512
	ti.Placeholder = i18n.T(i18n.KeyWizardPasswordTestDesc)

	sp := spinner.New()
	sp.Spinner = spinner.Dot

	return &connectionTestField{
		baseField: baseField{
			accessor:  &huh.EmbeddedAccessor[string]{},
			id:        nextBaseFieldID(),
			textinput: ti,
			keymap:    wizardCredentialKeyMap(),
		},
		strategy:    &keyIdentityTestStrategy{in: in, testAuth: testAuth},
		ctx:         ctx,
		onOK:        onOK,
		onFail:      onFail,
		spinner:     sp,
		state:       testStateEdit,
		description: i18n.T(i18n.KeyWizardIdentityDesc),
	}
}

// --- Builder 方法 ---

func (f *connectionTestField) Value(v *string) *connectionTestField {
	f.setValue(v)
	return f
}

func (f *connectionTestField) Key(k string) *connectionTestField {
	f.setKey(k)
	return f
}

func (f *connectionTestField) Title(t string) *connectionTestField {
	f.setTitle(t)
	return f
}

func (f *connectionTestField) Description(d string) *connectionTestField {
	f.description = d
	return f
}

// --- 委托给 baseField 的通用方法 ---

func (f *connectionTestField) Init() tea.Cmd { return f.baseField.Init() }
func (f *connectionTestField) Error() error  { return f.baseField.Error() }
func (f *connectionTestField) Skip() bool    { return f.baseField.Skip() }
func (f *connectionTestField) Zoom() bool    { return f.baseField.Zoom() }
func (f *connectionTestField) Run() error    { return huh.Run(f) }
func (f *connectionTestField) RunAccessible(w io.Writer, r io.Reader) error {
	return f.baseField.RunAccessible(w, r)
}

func (f *connectionTestField) WithTheme(theme *huh.Theme) huh.Field {
	f.baseField.WithTheme(theme)
	return f
}

func (f *connectionTestField) WithAccessible(a bool) huh.Field {
	f.baseField.WithAccessible(a)
	return f
}

func (f *connectionTestField) WithKeyMap(k *huh.KeyMap) huh.Field {
	f.baseField.WithKeyMap(k)
	return f
}

func (f *connectionTestField) WithWidth(w int) huh.Field {
	f.baseField.WithWidth(w)
	return f
}

func (f *connectionTestField) WithHeight(h int) huh.Field {
	f.baseField.WithHeight(h)
	return f
}

func (f *connectionTestField) WithPosition(p huh.FieldPosition) huh.Field {
	applyCredentialNavPosition(&f.keymap, p)
	return f
}

func (f *connectionTestField) GetKey() string { return f.baseField.GetKey() }
func (f *connectionTestField) GetValue() any  { return f.baseField.GetValue() }

// --- connectionTestField 自己的差异逻辑 ---

func (f *connectionTestField) activeStyles() *huh.FieldStyles {
	if f.theme == nil {
		f.theme = WizardTheme()
	}
	// 测连成功或测试中时保持 blurred 样式
	if f.focused && f.state != testStateTesting && f.state != testStateOK {
		return &f.theme.Focused
	}
	return &f.theme.Blurred
}

func (f *connectionTestField) statusStyle() lipgloss.Style {
	if f.theme == nil {
		return lipgloss.NewStyle()
	}
	switch f.state {
	case testStateOK:
		return wizardStatusOKStyle(f.theme)
	case testStateFail:
		return f.theme.Focused.ErrorMessage
	default:
		return f.theme.Blurred.Description
	}
}

func (f *connectionTestField) statusText() string {
	switch f.state {
	case testStateTesting:
		return i18n.T(i18n.KeyWizardTestingConnInline)
	case testStateOK:
		return i18n.T(i18n.KeyWizardConnOKMs, f.elapsed.Milliseconds())
	case testStateFail:
		if f.inlineMsg != "" {
			return f.inlineMsg
		}
		return f.strategy.failureDefaultMsg()
	default:
		return ""
	}
}

func (f *connectionTestField) renderInputRow(styles *huh.FieldStyles) string {
	line := f.textinput.View()
	status := f.statusText()
	if status == "" {
		return line
	}
	var suffix string
	if f.state == testStateTesting {
		suffix = f.spinner.View() + " " + f.statusStyle().Render(status)
	} else {
		suffix = f.statusStyle().Render(status)
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, line, " ", suffix)
}

func (f *connectionTestField) startTest(input string) tea.Cmd {
	validated, err := f.strategy.validateInput(input)
	if err != nil {
		f.state = testStateFail
		f.inlineMsg = err.Error()
		if f.onFail != nil {
			f.onFail()
		}
		f.textinput.Focus()
		f.focused = true
		return textinput.Blink
	}
	f.state = testStateTesting
	f.inlineMsg = ""
	f.textinput.Blur()

	testCmd := func() tea.Msg {
		// 使用策略的 startTestCmd，但需要同步返回 testDoneMsg
		return f.strategy.startTestCmd(f.ctx, validated)()
	}
	return tea.Batch(f.spinner.Tick, testCmd)
}

func (f *connectionTestField) handleTestDone(msg testDoneMsg) (tea.Model, tea.Cmd) {
	// huh Group 会对同一消息 Update 两次，避免重复 NextField 直接提交表单。
	if f.state == testStateOK {
		return f, nil
	}
	if msg.err != nil {
		f.state = testStateFail
		f.inlineMsg = connectionTestFailureMessage(msg.err)
		if f.onFail != nil {
			f.onFail()
		}
		f.textinput.SetValue("")
		f.accessor.Set("")
		f.strategy.clearCredential()
		f.textinput.Focus()
		f.focused = true
		return f, textinput.Blink
	}
	f.state = testStateOK
	f.elapsed = msg.elapsed
	f.inlineMsg = ""
	if f.onOK != nil {
		f.onOK()
	}
	// 测连成功后自动进入备注步，无需再按 Enter。
	return f, huh.NextField
}

// resumeEditFromSuccess 在测连已通过时回到可编辑态，便于用户改密码/密钥后重新测试。
func (f *connectionTestField) resumeEditFromSuccess() tea.Cmd {
	f.state = testStateEdit
	f.elapsed = 0
	f.inlineMsg = ""
	if f.onFail != nil {
		f.onFail()
	}
	f.focused = true
	return f.textinput.Focus()
}

// Focus 覆盖 baseField.Focus，根据状态决定是否聚焦 textinput。
func (f *connectionTestField) Focus() tea.Cmd {
	f.focused = true
	if f.state == testStateEdit || f.state == testStateFail {
		return f.textinput.Focus()
	}
	return nil
}

// Blur 覆盖 baseField.Blur。
func (f *connectionTestField) Blur() tea.Cmd {
	f.focused = false
	f.textinput.Blur()
	f.accessor.Set(f.textinput.Value())
	return nil
}

// KeyBinds 覆盖 baseField.KeyBinds，根据状态返回不同的绑定。
func (f *connectionTestField) KeyBinds() []key.Binding {
	if f.state == testStateTesting {
		return nil
	}
	return []key.Binding{f.keymap.Prev, f.keymap.Submit, f.keymap.Next}
}

func (f *connectionTestField) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case testDoneMsg:
		return f.handleTestDone(msg)
	case spinner.TickMsg:
		if f.state == testStateTesting {
			var cmd tea.Cmd
			f.spinner, cmd = f.spinner.Update(msg)
			return f, cmd
		}
	case tea.KeyMsg:
		if f.state == testStateTesting {
			return f, nil
		}
		if f.state == testStateOK {
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
			raw := strings.TrimSpace(f.textinput.Value())
			if raw == "" {
				f.state = testStateFail
				f.inlineMsg = i18n.T(i18n.KeyWizardErrEmpty)
				return f, nil
			}
			return f, f.startTest(raw)
		case key.Matches(msg, f.keymap.Next):
			return f, huh.NextField
		}
	}

	if f.state == testStateEdit || f.state == testStateFail {
		var cmd tea.Cmd
		f.textinput, cmd = f.textinput.Update(msg)
		f.accessor.Set(f.textinput.Value())
		cmds = append(cmds, cmd)
	}

	return f, tea.Batch(cmds...)
}

func (f *connectionTestField) View() string {
	styles := f.activeStyles()
	f.applyTextInputStyles(styles)

	var below []string
	if f.strategy.hasDescription() && f.state == testStateEdit && f.description != "" {
		below = append(below, styles.Description.Render(f.description))
	}
	return renderInlineField(f.width, f.height, styles, f.title, f.renderInputRow(styles), below...)
}

// --- 密码策略实现 ---

type passwordTestStrategy struct {
	in       *PasswordModeInput
	testAuth passwordAuthTestFn
}

func (s *passwordTestStrategy) validateInput(raw string) (string, error) {
	pwd := strings.TrimSpace(raw)
	if pwd == "" {
		return "", errors.New(i18n.T(i18n.KeyWizardErrEmpty))
	}
	return pwd, nil
}

func (s *passwordTestStrategy) startTestCmd(ctx context.Context, input string) tea.Cmd {
	s.in.Password = input
	return func() tea.Msg {
		elapsed, err := testPasswordConnection(ctx, s.in, input, s.testAuth)
		return testDoneMsg{err: err, elapsed: elapsed}
	}
}

func (s *passwordTestStrategy) clearCredential() {
	s.in.Password = ""
}

func (s *passwordTestStrategy) failureDefaultMsg() string {
	return i18n.T(i18n.KeyWizardConnFailInline)
}

func (s *passwordTestStrategy) hasDescription() bool { return false }

// --- 密钥策略实现 ---

type keyIdentityTestStrategy struct {
	in       *KeyModeInput
	testAuth keyAuthTestFn
}

func (s *keyIdentityTestStrategy) validateInput(raw string) (string, error) {
	path := strings.TrimSpace(raw)
	if path == "" {
		return "", errors.New(i18n.T(i18n.KeyWizardErrEmpty))
	}
	expanded, err := platform.ExpandPath(path)
	if err != nil {
		return "", err
	}
	if _, err := os.Stat(expanded); err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("%s: %s", i18n.T(i18n.KeyWizardErrKeyMissing), expanded)
		}
		return "", fmt.Errorf("%s: %v", i18n.T(i18n.KeyWizardErrKeyRead), err)
	}
	s.in.IdentityFile = expanded
	return expanded, nil
}

func (s *keyIdentityTestStrategy) startTestCmd(ctx context.Context, _ string) tea.Cmd {
	return func() tea.Msg {
		err := s.testAuth(ctx, *s.in)
		return testDoneMsg{err: err}
	}
}

func (s *keyIdentityTestStrategy) clearCredential() {
	s.in.IdentityFile = ""
}

func (s *keyIdentityTestStrategy) failureDefaultMsg() string {
	return i18n.T(i18n.KeyWizardKeyAuthFailed)
}

func (s *keyIdentityTestStrategy) hasDescription() bool { return true }
