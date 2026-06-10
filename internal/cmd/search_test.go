package cmd

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fuckssh/fuckssh/internal/config"
)

func Test_Search_matchesAliasViaCmd(t *testing.T) {
	testConfigPath = fixtureConfig("multiple.conf")
	t.Cleanup(func() { testConfigPath = "" })

	var stdout, stderr bytes.Buffer
	if err := runSearchCmd([]string{"srv1"}, &stdout, &stderr); err != nil {
		t.Fatalf("runSearchCmd: %v", err)
	}
	if !strings.Contains(stderr.String(), "搜索") && !strings.Contains(stderr.String(), "Search") {
		t.Errorf("stderr = %q, want search meta", stderr.String())
	}
	if !strings.Contains(stdout.String(), "srv1") || strings.Contains(stdout.String(), "srv2") {
		t.Fatalf("stdout should only contain srv1:\n%s", stdout.String())
	}
}

func Test_Search_multiKeywordOR(t *testing.T) {
	testConfigPath = fixtureConfig("multiple.conf")
	t.Cleanup(func() { testConfigPath = "" })

	var stdout, stderr bytes.Buffer
	if err := runSearchCmd([]string{"srv1", "example"}, &stdout, &stderr); err != nil {
		t.Fatalf("runSearchCmd: %v", err)
	}
	out := stdout.String()
	if !strings.Contains(out, "srv1") || !strings.Contains(out, "example.com") {
		t.Fatalf("stdout should contain both srv1 and example.com:\n%s", out)
	}
}

func Test_Search_noMatch_returnsEmptyWithHint(t *testing.T) {
	testConfigPath = fixtureConfig("multiple.conf")
	t.Cleanup(func() { testConfigPath = "" })

	var stdout, stderr bytes.Buffer
	if err := runSearchCmd([]string{"nomatch-xyz"}, &stdout, &stderr); err != nil {
		t.Fatalf("runSearchCmd: %v", err)
	}
	out := stdout.String()
	if !strings.Contains(out, "未找到匹配") && !strings.Contains(out, "No Host entries match") {
		t.Errorf("stdout = %q, want no-match hint", out)
	}
	if !strings.Contains(out, "fuckssh list") {
		t.Errorf("stdout = %q, want hint to list", out)
	}
}

func Test_Search_invalidConfig_returnsParseError(t *testing.T) {
	testConfigPath = fixtureConfig("invalid.conf")
	t.Cleanup(func() { testConfigPath = "" })

	var stdout, stderr bytes.Buffer
	err := runSearchCmd([]string{"foo"}, &stdout, &stderr)
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

func Test_Search_missingConfig_returnsExitCode3(t *testing.T) {
	testConfigPath = filepath.Join(t.TempDir(), "no-such-config.conf")
	t.Cleanup(func() { testConfigPath = "" })

	var stdout, stderr bytes.Buffer
	err := runSearchCmd([]string{"foo"}, &stdout, &stderr)
	if err == nil {
		t.Fatal("expected error for missing config")
	}
	if !errors.Is(err, os.ErrNotExist) {
		t.Errorf("err = %v, want os.ErrNotExist", err)
	}
	if got := ExitCode(err); got != 3 {
		t.Errorf("ExitCode = %d, want 3 for missing file", got)
	}
}

func Test_Search_withUserFlag(t *testing.T) {
	testConfigPath = fixtureConfig("multiple.conf")
	t.Cleanup(func() {
		testConfigPath = ""
		resetSearchFlags()
	})
	searchUserFlag = "admin"

	var stdout, stderr bytes.Buffer
	if err := runSearchCmd([]string{"srv"}, &stdout, &stderr); err != nil {
		t.Fatalf("runSearchCmd: %v", err)
	}
	out := stdout.String()
	if !strings.Contains(out, "srv1") || strings.Contains(out, "srv2") {
		t.Fatalf("stdout should only contain srv1 (user=admin):\n%s", out)
	}
}

func Test_Search_withPortFlag(t *testing.T) {
	testConfigPath = fixtureConfig("multiple.conf")
	t.Cleanup(func() {
		testConfigPath = ""
		resetSearchFlags()
	})
	searchPortFlag = "2222"

	var stdout, stderr bytes.Buffer
	if err := runSearchCmd([]string{"srv"}, &stdout, &stderr); err != nil {
		t.Fatalf("runSearchCmd: %v", err)
	}
	out := stdout.String()
	if !strings.Contains(out, "srv1") || strings.Contains(out, "srv2") {
		t.Fatalf("stdout should only contain srv1 (port=2222):\n%s", out)
	}
}

func Test_Search_highlightDisabled_pipedOutput(t *testing.T) {
	// stdout 为 bytes.Buffer（非 TTY），不应包含 ANSI 转义码。
	testConfigPath = fixtureConfig("multiple.conf")
	t.Cleanup(func() { testConfigPath = "" })

	var stdout, stderr bytes.Buffer
	if err := runSearchCmd([]string{"srv1"}, &stdout, &stderr); err != nil {
		t.Fatalf("runSearchCmd: %v", err)
	}
	if strings.Contains(stdout.String(), "\033[") {
		t.Errorf("piped stdout should not contain ANSI codes:\n%s", stdout.String())
	}
}
