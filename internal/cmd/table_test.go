package cmd

import (
	"strings"
	"testing"

	"github.com/fuckssh/fuckssh/internal/config"
)

func TestFormatHostsTable_columnAlignment(t *testing.T) {
	entries, err := config.ParseFile(fixtureConfig("multiple.conf"))
	if err != nil {
		t.Fatalf("ParseFile: %v", err)
	}

	out := formatHostsTable(entries, nil)
	if !strings.Contains(out, "srv1") || !strings.Contains(out, "srv2") {
		t.Fatalf("unexpected output:\n%s", out)
	}
	if !strings.Contains(out, "srv2-alt") {
		t.Errorf("output should list all aliases for srv2, got:\n%s", out)
	}
	// 无备注时显示占位符。
	if !strings.Contains(out, "-") {
		t.Errorf("output should show dash for empty remark:\n%s", out)
	}
	for _, box := range []string{"┌", "┬", "┐", "├", "┼", "┤", "└", "┴", "┘", "│", "─"} {
		if !strings.Contains(out, box) {
			t.Errorf("output missing box char %q in:\n%s", box, out)
		}
	}
}

func TestFormatHostsTable_showsRemark(t *testing.T) {
	entries, err := config.ParseFile(fixtureConfig("with_remark.conf"))
	if err != nil {
		t.Fatalf("ParseFile: %v", err)
	}
	out := formatHostsTable(entries, nil)
	if !strings.Contains(out, "生产环境主站") || !strings.Contains(out, "测试机") {
		t.Errorf("output should include remarks:\n%s", out)
	}
}
