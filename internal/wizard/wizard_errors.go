package wizard

// wizard_errors.go 合并错误定义、取消处理与表单提示。
//
// 设计说明：
// - ErrWizardRetryForm / cancelError / formRetryHint 都属于向导流程控制
// - 合并后减少 3 个碎片文件，阅读时一处即可理解全部流程控制机制

import (
	"errors"

	"github.com/charmbracelet/huh"
	"github.com/fuckssh/fuckssh/internal/i18n"
)

// ErrWizardRetryForm 表示用户在确认屏选择「返回修改」，应重新收集表单。
var ErrWizardRetryForm = errors.New("wizard: retry form")

// --- 取消处理 ---

// cancelError 表示用户在向导中主动取消（Ctrl+C）；对外只展示友好中文/英文一句。
type cancelError struct {
	msg string
}

func (e *cancelError) Error() string { return e.msg }

func (e *cancelError) Is(target error) bool {
	return target == huh.ErrUserAborted
}

// UserCancelled 返回供 CLI 展示的单行取消错误（与 mapWizardAbort 一致）。
func UserCancelled() error {
	return &cancelError{msg: i18n.T(i18n.KeyWizardCancelled)}
}

// CancelMessage 返回应展示给用户的取消文案。
func CancelMessage(err error) string {
	var ce *cancelError
	if errors.As(err, &ce) {
		return ce.msg
	}
	if errors.Is(err, huh.ErrUserAborted) {
		return i18n.T(i18n.KeyWizardCancelled)
	}
	return err.Error()
}

// IsCancelled 判断是否为向导用户取消（含 cancelError 与 huh.ErrUserAborted）。
func IsCancelled(err error) bool {
	if err == nil {
		return false
	}
	var ce *cancelError
	if errors.As(err, &ce) {
		return true
	}
	return errors.Is(err, huh.ErrUserAborted)
}

// mapWizardAbort 将 Ctrl+C 映射为单行取消提示；其它错误原样返回。
func mapWizardAbort(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, huh.ErrUserAborted) {
		return &cancelError{msg: i18n.T(i18n.KeyWizardCancelled)}
	}
	return err
}

// --- 表单提示 ---

// formRetryHint 在用户从确认页返回修改时，展示在表单首字段 placeholder（仅消费一次）。
var formRetryHint string

// SetFormRetryHint 设置返回修改提示（由 confirm 调用）。
func SetFormRetryHint(msg string) {
	formRetryHint = msg
}

// consumeFormRetryHint 读取并清除返回修改提示。
func consumeFormRetryHint() string {
	h := formRetryHint
	formRetryHint = ""
	return h
}

// hostInputPlaceholder 主机地址输入框占位（含返回修改提示）。
func hostInputPlaceholder() string {
	if h := consumeFormRetryHint(); h != "" {
		return h
	}
	return i18n.T(i18n.KeyWizardHostPlaceholder)
}
