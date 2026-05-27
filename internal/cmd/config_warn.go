package cmd

import (
	"fmt"
	"io"

	"github.com/fuckssh/fuckssh/internal/config"
	"github.com/fuckssh/fuckssh/internal/i18n"
)

// maybeWarnInclude 在 list/search 前提示 config 含 Include 时 MVP 不会展开。
func maybeWarnInclude(stderr io.Writer, configPath string) {
	has, err := config.HasIncludeDirective(configPath)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "warning: %v\n", err)
		return
	}
	if !has {
		return
	}
	_, _ = fmt.Fprintf(stderr, "%s\n", i18n.T(i18n.KeyConfigIncludeSkipped))
}
