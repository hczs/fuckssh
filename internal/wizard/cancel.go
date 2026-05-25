package wizard

import (
	"errors"

	"github.com/charmbracelet/huh"
	"github.com/fuckssh/fuckssh/internal/i18n"
)

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
