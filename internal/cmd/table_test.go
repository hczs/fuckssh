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

	out := formatHostsTable(entries)
	if !strings.Contains(out, "srv1") || !strings.Contains(out, "srv2") {
		t.Fatalf("unexpected output:\n%s", out)
	}
	if !strings.Contains(out, "srv2-alt") {
		t.Errorf("output should list all aliases for srv2, got:\n%s", out)
	}
	for _, box := range []string{"┌", "┬", "┐", "├", "┼", "┤", "└", "┴", "┘", "│", "─"} {
		if !strings.Contains(out, box) {
			t.Errorf("output missing box char %q in:\n%s", box, out)
		}
	}
}
