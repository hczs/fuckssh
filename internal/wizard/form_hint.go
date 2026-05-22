package wizard

// formRetryHint 在用户从确认页返回修改时，展示在表单首字段上方（仅消费一次）。
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

// firstFieldDescription 合并返回修改提示与字段原有说明。
func firstFieldDescription(base string) string {
	if h := consumeFormRetryHint(); h != "" {
		if base == "" {
			return h
		}
		return h + "\n\n" + base
	}
	return base
}
