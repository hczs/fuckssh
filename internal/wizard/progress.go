package wizard

import (
	"fmt"
	"io"
	"os"
)

// progressOut 为步骤进度输出目标（默认 stderr，测试可替换）。
var progressOut io.Writer = os.Stderr

// reportProgress 向用户输出当前正在执行的步骤（架构 §8.1：进度走 stderr）。
func reportProgress(msg string) {
	if progressOut == nil || msg == "" {
		return
	}
	fmt.Fprintf(progressOut, "%s\n", msg)
}
