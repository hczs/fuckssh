package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/fuckssh/fuckssh/internal/config"
	"github.com/fuckssh/fuckssh/internal/vault"
)

func TestPrintKeyImportStats_showsRenameAndSkip(t *testing.T) {
	var buf bytes.Buffer
	result := &vault.ImportResult{KeysImported: 1, KeysSkipped: 2}
	plan := &config.KeyImportPlan{KeysRenamed: []string{"shared→id_ed25519_fuckssh_staging"}}

	printKeyImportStats(&buf, result, plan)
	out := buf.String()

	if !strings.Contains(out, "密钥已按别名重命名") {
		t.Errorf("输出应包含密钥重命名信息: %q", out)
	}
	if !strings.Contains(out, "私钥已存在且内容相同") {
		t.Errorf("输出应包含跳过覆盖信息: %q", out)
	}
}
