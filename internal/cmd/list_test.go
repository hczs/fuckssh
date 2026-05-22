package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fuckssh/fuckssh/internal/config"
)

func fixtureConfig(name string) string {
	return filepath.Join("..", "..", "testdata", "config", name)
}

func Test_ListCmd_printsTableFromFixture(t *testing.T) {
	var buf bytes.Buffer
	if err := runList(fixtureConfig("multiple.conf"), &buf); err != nil {
		t.Fatalf("runList: %v", err)
	}

	out := buf.String()
	for _, col := range []string{"ALIAS", "HOSTNAME", "PORT", "USER"} {
		if !strings.Contains(out, col) {
			t.Errorf("output missing column header %q", col)
		}
	}
	for _, row := range []string{"srv1", "10.0.0.1", "2222", "admin", "srv2", "example.com", "deploy"} {
		if !strings.Contains(out, row) {
			t.Errorf("output missing %q in:\n%s", row, out)
		}
	}
}

func Test_ListCmd_emptyHostsFriendlyMessage(t *testing.T) {
	var buf bytes.Buffer
	empty := filepath.Join(t.TempDir(), "empty.conf")
	if err := os.WriteFile(empty, nil, 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	if err := runList(empty, &buf); err != nil {
		t.Fatalf("runList: %v", err)
	}
	if !strings.Contains(buf.String(), "未找到 Host 条目") {
		t.Errorf("output = %q, want empty hint", buf.String())
	}
}

func Test_ListCmd_respectsConfigFlag(t *testing.T) {
	configFileFlag = fixtureConfig("single.conf")
	t.Cleanup(func() { configFileFlag = "" })

	var buf bytes.Buffer
	if err := runListCmd(&buf, &buf); err != nil {
		t.Fatalf("runListCmd: %v", err)
	}

	if !strings.Contains(buf.String(), "myserver") {
		t.Errorf("output = %q, want myserver from single.conf", buf.String())
	}
}

func TestFormatHosts_columnAlignment(t *testing.T) {
	entries, err := config.ParseFile(fixtureConfig("multiple.conf"))
	if err != nil {
		t.Fatalf("ParseFile: %v", err)
	}

	out := FormatHosts(entries)
	if !strings.Contains(out, "srv1") || !strings.Contains(out, "srv2") {
		t.Fatalf("unexpected output:\n%s", out)
	}
}
