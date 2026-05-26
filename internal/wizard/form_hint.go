package wizard

import (
	"github.com/fuckssh/fuckssh/internal/i18n"
)

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
