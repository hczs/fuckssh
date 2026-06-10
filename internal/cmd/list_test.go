package cmd

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fuckssh/fuckssh/internal/config"
	"github.com/fuckssh/fuckssh/internal/i18n"
)

func fixtureConfig(name string) string {
	return filepath.Join("..", "..", "testdata", "config", name)
}

// assertContainsAny 断言 s 至少包含 subs 之一（便于中英双语测试）。
func assertContainsAny(t *testing.T, s, label string, subs ...string) {
	t.Helper()
	for _, sub := range subs {
		if strings.Contains(s, sub) {
			return
		}
	}
	t.Errorf("%s should contain one of %v, got: %q", label, subs, s)
}

func Test_ListCmd_printsTableFromFixture(t *testing.T) {
	path := fixtureConfig("multiple.conf")
	var stdout, stderr bytes.Buffer
	if err := runList(path, &stdout, &stderr); err != nil {
		t.Fatalf("runList: %v", err)
	}

	assertContainsAny(t, stderr.String(), "stderr host count", "共", "host(s)")
	assertContainsAny(t, stderr.String(), "stderr reading meta", "读取", "Reading")
	if !strings.Contains(stderr.String(), path) {
		t.Errorf("stderr should mention config path %q, got: %q", path, stderr.String())
	}

	out := stdout.String()
	hasZH := strings.Contains(out, "别名") && strings.Contains(out, "地址")
	hasEN := strings.Contains(out, "ALIAS") && strings.Contains(out, "HOSTNAME")
	if !hasZH && !hasEN {
		t.Errorf("output missing table headers in:\n%s", out)
	}
	for _, row := range []string{
		"srv1", "10.0.0.1", "2222", "admin",
		"srv2", "srv2-alt", "example.com", "deploy",
	} {
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
	assertContainsAny(t, stdout.String(), "stdout empty hint",
		"未找到 Host 条目", "No Host entries found")
	assertContainsAny(t, stdout.String(), "stdout CTA", "fuckssh add")
}

func Test_ListCmd_respectsConfigFlag(t *testing.T) {
	testConfigPath = fixtureConfig("single.conf")
	t.Cleanup(func() { testConfigPath = "" })

	var stdout, stderr bytes.Buffer
	if err := runListCmd(&stdout, &stderr); err != nil {
		t.Fatalf("runListCmd: %v", err)
	}

	if !strings.Contains(stdout.String(), "myserver") {
		t.Errorf("stdout = %q, want myserver from single.conf", stdout.String())
	}
}

func Test_ListCmd_invalidConfig_returnsParseError(t *testing.T) {
	var stdout, stderr bytes.Buffer
	err := runList(fixtureConfig("invalid.conf"), &stdout, &stderr)
	if err == nil {
		t.Fatal("expected parse error")
	}
	var pe *config.ParseError
	if !errors.As(err, &pe) {
		t.Fatalf("error type = %T, want *config.ParseError", err)
	}
	if got := ExitCode(err); got != 2 {
		t.Errorf("ExitCode = %d, want 2 for parse error", got)
	}
}

func Test_ListCmd_missingConfig_returnsExitCode3(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "no-such-config.conf")
	var stdout, stderr bytes.Buffer
	err := runList(missing, &stdout, &stderr)
	if err == nil {
		t.Fatal("expected error for missing config")
	}
	if got := ExitCode(err); got != 3 {
		t.Errorf("ExitCode = %d, want 3 for missing file", got)
	}
}

func Test_ListCmd_rejectsExtraArgs(t *testing.T) {
	resetHelpFlags(rootCmd)
	testConfigPath = fixtureConfig("single.conf")
	t.Cleanup(func() {
		testConfigPath = ""
		resetHelpFlags(rootCmd)
	})

	var stdout, stderr bytes.Buffer
	rootCmd.SetOut(&stdout)
	rootCmd.SetErr(&stderr)
	i18n.SetInteractiveOverrideForTest(func(io.Writer) bool { return false })

	err := ExecuteWithArgs([]string{"list", "extra"})
	if err == nil {
		t.Fatal("expected error when extra args passed to list")
	}
}
