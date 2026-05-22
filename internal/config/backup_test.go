package config

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestBackup_createsTimestampedCopy(t *testing.T) {
	dir := t.TempDir()
	if runtime.GOOS == "windows" {
		t.Setenv("USERPROFILE", dir)
	} else {
		t.Setenv("HOME", dir)
	}
	sshDir := filepath.Join(dir, ".ssh")
	if err := os.MkdirAll(sshDir, 0o700); err != nil {
		t.Fatal(err)
	}
	cfg := filepath.Join(sshDir, "config")
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
	backupDir := filepath.Join(sshDir, "backup")
	if filepath.Dir(bakPath) != backupDir {
		t.Errorf("bak dir = %q, want %q", filepath.Dir(bakPath), backupDir)
	}

	got, err := os.ReadFile(bakPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != content {
		t.Errorf("backup content = %q, want %q", got, content)
	}
}
