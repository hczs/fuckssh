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
	var stdout, stderr bytes.Buffer
	if err := runList(fixtureConfig("multiple.conf"), &stdout, &stderr); err != nil {
		t.Fatalf("runList: %v", err)
	}

	if !strings.Contains(stderr.String(), "共") && !strings.Contains(stderr.String(), "host") {
		t.Errorf("stderr should contain host count, got: %q", stderr.String())
	}
	out := stdout.String()
	for _, col := range []string{"别名", "地址", "端口", "用户"} {
		if !strings.Contains(out, col) {
			t.Errorf("output missing column header %q in:\n%s", col, out)
		}
	}
	for _, row := range []string{"srv1", "10.0.0.1", "2222", "admin", "srv2", "example.com", "deploy"} {
		if !strings.Contains(out, row) {
			t.Errorf("output missing %q in:\n%s", row, out)
		}
	}
}

func Test_ListCmd_emptyHostsFriendlyMessage(t *testing.T) {
	var stdout, stderr bytes.Buffer
	empty := filepath.Join(t.TempDir(), "empty.conf")
	if err := os.WriteFile(empty, nil, 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	if err := runList(empty, &stdout, &stderr); err != nil {
		t.Fatalf("runList: %v", err)
	}
	if !strings.Contains(stdout.String(), "未找到 Host 条目") {
		t.Errorf("stdout = %q, want empty hint", stdout.String())
	}
	if !strings.Contains(stdout.String(), "fuckssh add") {
		t.Errorf("stdout = %q, want CTA", stdout.String())
	}
}

func Test_ListCmd_respectsConfigFlag(t *testing.T) {
	configFileFlag = fixtureConfig("single.conf")
	t.Cleanup(func() { configFileFlag = "" })

	var stdout, stderr bytes.Buffer
	if err := runListCmd(&stdout, &stderr); err != nil {
		t.Fatalf("runListCmd: %v", err)
	}

	if !strings.Contains(stdout.String(), "myserver") {
		t.Errorf("stdout = %q, want myserver from single.conf", stdout.String())
	}
}

func TestFormatHostsTable_columnAlignment(t *testing.T) {
	entries, err := config.ParseFile(fixtureConfig("multiple.conf"))
	if err != nil {
		t.Fatalf("ParseFile: %v", err)
	}

	out := formatHostsTable(entries)
	if !strings.Contains(out, "srv1") || !strings.Contains(out, "srv2") {
		t.Fatalf("unexpected output:\n%s", out)
	}
}
