package wizard

import (
	"fmt"
	"io"
	"os"

	"github.com/fuckssh/fuckssh/internal/i18n"
)

// progressOut 为步骤进度输出目标（默认 stderr，测试可替换）。
var progressOut io.Writer = os.Stderr

// reportProgressStep 输出带序号的步骤进度，例如 [2/4] 正在生成密钥…
func reportProgressStep(step, total int, msg string) {
	if progressOut == nil || msg == "" {
		return
	}
	_, _ = fmt.Fprintf(progressOut, "%s\n", i18n.T(i18n.KeyWizardProgressStep, step, total, msg))
}

// ReportProgressStep 供 cmd 等包在确认后输出与密码模式一致的执行进度。
func ReportProgressStep(step, total int, msg string) {
	reportProgressStep(step, total, msg)
}

// KeyFlowProgressTotal 为密钥模式确认后的本地写盘步骤数。
const KeyFlowProgressTotal = 3
