package cmd

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fuckssh/fuckssh/internal/i18n"
)

func TestMain(m *testing.M) {
	// 固定中文界面；勿用 FUCKSSH_LANG 环境变量，以免干扰「从 settings.json 加载语言」等单测。
	i18n.ResetForTest()
	i18n.SetCurrent(i18n.LangZH)
	code := m.Run()
	os.Exit(code)
}

func TestRootHelp(t *testing.T) {
	rootCmd.SetArgs([]string{"--help"})
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}

	out := buf.String()
	for _, sub := range []string{"add", "list", "search"} {
		if !strings.Contains(out, sub) {
			t.Errorf("help output missing subcommand %q", sub)
		}
	}
	t.Cleanup(func() {
		resetHelpFlags(rootCmd)
		clearCommandArgs(rootCmd)
	})
}

func TestExecute_printsElapsedOnList(t *testing.T) {
	configFileFlag = fixtureConfig("multiple.conf")
	t.Cleanup(func() {
		configFileFlag = ""
		resetHelpFlags(rootCmd)
	})

	rootCmd.SetArgs([]string{"list"})
	var stdout, stderr bytes.Buffer
	rootCmd.SetOut(&stdout)
	rootCmd.SetErr(&stderr)
	i18n.SetInteractiveOverrideForTest(func(io.Writer) bool { return false })

	if err := Execute(); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	assertContainsAny(t, stderr.String(), "stderr elapsed", "执行耗时", "Elapsed")
	if !strings.Contains(stderr.String(), "ms") {
		t.Errorf("stderr should contain ms, got: %q", stderr.String())
	}
}

func TestExecute_skipsElapsedOnHelp(t *testing.T) {
	rootCmd.SetArgs([]string{"list", "--help"})
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)

	if err := Execute(); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	for _, sub := range []string{"执行耗时", "Elapsed"} {
		if strings.Contains(buf.String(), sub) {
			t.Errorf("help output should not contain elapsed hint %q, got: %q", sub, buf.String())
		}
	}
	t.Cleanup(func() {
		resetHelpFlags(rootCmd)
		clearCommandArgs(rootCmd)
	})
}

func TestEnsureLanguageFromSettings(t *testing.T) {
	resetHelpFlags(rootCmd)
	i18n.ResetForTest()
	dir := t.TempDir()
	path := filepath.Join(dir, "settings.json")
	if err := os.WriteFile(path, []byte(`{"lang":"en"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	i18n.SetSettingsPathForTest(path)
	if err := i18n.EnsureLoaded(); err != nil {
		t.Fatal(err)
	}
	if i18n.Current() != i18n.LangEN {
		t.Fatalf("after EnsureLoaded Current() = %q, want en", i18n.Current())
	}

	rootCmd.SetArgs([]string{"list", "--config", filepath.Join(t.TempDir(), "missing")})
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	i18n.SetInteractiveOverrideForTest(func(io.Writer) bool { return false })

	// list 会因文件不存在失败，但语言应已保持为 settings 中的 en
	_ = rootCmd.Execute()
	if i18n.Current() != i18n.LangEN {
		t.Errorf("Current() = %q, want en", i18n.Current())
	}
	i18n.ResetForTest()
	i18n.SetCurrent(i18n.LangZH)
}
