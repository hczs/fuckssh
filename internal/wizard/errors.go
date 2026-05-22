package wizard

import "errors"

// ErrWizardRetryForm 表示用户在确认屏选择「返回修改」，应重新收集表单。
var ErrWizardRetryForm = errors.New("wizard: retry form")
