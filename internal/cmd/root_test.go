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
	_ = os.Setenv("FUCKSSH_LANG", "zh")
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
}

func TestEnsureLanguageFromSettings(t *testing.T) {
	t.Setenv("FUCKSSH_LANG", "")
	i18n.ResetForTest()
	dir := t.TempDir()
	path := filepath.Join(dir, "settings.json")
	i18n.SetSettingsPathForTest(path)
	if err := i18n.Save(i18n.LangEN); err != nil {
		t.Fatal(err)
	}
	i18n.ResetForTest()
	i18n.SetSettingsPathForTest(path)

	rootCmd.SetArgs([]string{"list", "--config", filepath.Join(t.TempDir(), "missing")})
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	i18n.SetInteractiveOverrideForTest(func(io.Writer) bool { return false })

	// list 会因文件不存在失败，但语言应已加载
	_ = rootCmd.Execute()
	if i18n.Current() != i18n.LangEN {
		t.Errorf("Current() = %q, want en", i18n.Current())
	}
	i18n.ResetForTest()
	i18n.SetCurrent(i18n.LangZH)
}
