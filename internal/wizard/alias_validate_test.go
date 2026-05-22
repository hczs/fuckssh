package wizard

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAliasFieldValidate_rejectsConflict(t *testing.T) {
	dir := t.TempDir()
	cfg := filepath.Join(dir, "config")
	if err := os.WriteFile(cfg, []byte("Host existing\n  HostName 1.2.3.4\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	host := "10.0.0.1"
	validate := aliasFieldValidate(cfg, &host)
	if err := validate("existing"); err == nil {
		t.Fatal("want conflict error")
	}
}

func TestAliasFieldValidate_allowsNewAlias(t *testing.T) {
	dir := t.TempDir()
	cfg := filepath.Join(dir, "config")
	host := "10.0.0.2"
	validate := aliasFieldValidate(cfg, &host)
	if err := validate("my-vps"); err != nil {
		t.Fatalf("validate: %v", err)
	}
}
