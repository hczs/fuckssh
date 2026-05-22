package i18n

import (
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadSave_roundTrip(t *testing.T) {
	ResetForTest()
	dir := t.TempDir()
	path := filepath.Join(dir, "settings.json")
	SetSettingsPathForTest(path)

	if err := Save(LangEN); err != nil {
		t.Fatal(err)
	}
	ResetForTest()
	SetSettingsPathForTest(path)

	ok, err := Load()
	if err != nil || !ok {
		t.Fatalf("Load() = %v, %v", ok, err)
	}
	if Current() != LangEN {
		t.Errorf("Current() = %q, want en", Current())
	}
}

func TestEnvLangOverridesFile(t *testing.T) {
	ResetForTest()
	t.Setenv(envLangKey, "en")
	dir := t.TempDir()
	SetSettingsPathForTest(filepath.Join(dir, "settings.json"))
	_ = Save(LangZH)

	ok, err := Load()
	if err != nil || !ok {
		t.Fatal(err)
	}
	if Current() != LangEN {
		t.Errorf("Current() = %q, want en from env", Current())
	}
}

func TestT_fallbackUnknownKey(t *testing.T) {
	ResetForTest()
	SetCurrent(LangZH)
	if got := T("unknown.key"); got != "unknown.key" {
		t.Errorf("T() = %q", got)
	}
}

func TestEnsureInteractive_nonTTYDefaultsZh(t *testing.T) {
	ResetForTest()
	dir := t.TempDir()
	SetSettingsPathForTest(filepath.Join(dir, "settings.json"))
	isInteractiveOverride = func(io.Writer) bool { return false }

	if err := EnsureInteractive(os.Stderr); err != nil {
		t.Fatal(err)
	}
	if Current() != LangZH {
		t.Errorf("Current() = %q, want zh", Current())
	}
}
