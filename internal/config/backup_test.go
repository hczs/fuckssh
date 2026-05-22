package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBackup_createsTimestampedCopy(t *testing.T) {
	dir := t.TempDir()
	cfg := filepath.Join(dir, "config")
	content := "Host old\n    HostName 1.2.3.4\n"
	if err := os.WriteFile(cfg, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	bakPath, err := Backup(cfg)
	if err != nil {
		t.Fatalf("Backup: %v", err)
	}

	if !strings.Contains(filepath.Base(bakPath), "config.fuckssh.bak.") {
		t.Errorf("bak basename = %q, want config.fuckssh.bak.<timestamp>", filepath.Base(bakPath))
	}
	if filepath.Dir(bakPath) != dir {
		t.Errorf("bak dir = %q, want %q", filepath.Dir(bakPath), dir)
	}

	got, err := os.ReadFile(bakPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != content {
		t.Errorf("backup content = %q, want %q", got, content)
	}
}
