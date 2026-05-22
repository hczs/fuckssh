package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRestoreFromBackup_overwritesConfig(t *testing.T) {
	dir := t.TempDir()
	cfg := filepath.Join(dir, "config")
	bak := filepath.Join(dir, "config.bak")
	if err := os.WriteFile(bak, []byte("Host old\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(cfg, []byte("Host new\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := RestoreFromBackup(bak, cfg); err != nil {
		t.Fatalf("RestoreFromBackup: %v", err)
	}
	got, err := os.ReadFile(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "Host old\n" {
		t.Errorf("config = %q, want restored content", got)
	}
}

func TestRollbackAfterAddFailure_removesNewConfigFile(t *testing.T) {
	dir := t.TempDir()
	cfg := filepath.Join(dir, "config")
	if err := os.WriteFile(cfg, []byte("Host x\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := RollbackAfterAddFailure(cfg, "", false, true); err != nil {
		t.Fatalf("RollbackAfterAddFailure: %v", err)
	}
	if _, err := os.Stat(cfg); !os.IsNotExist(err) {
		t.Errorf("config should be removed, stat err = %v", err)
	}
}

func TestRollbackAfterAddFailure_noOpWithoutBackupOrModify(t *testing.T) {
	dir := t.TempDir()
	cfg := filepath.Join(dir, "config")
	orig := "Host keep\n    HostName 1.2.3.4\n    User root\n    IdentityFile /tmp/k\n"
	if err := os.WriteFile(cfg, []byte(orig), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := RollbackAfterAddFailure(cfg, "", true, false); err != nil {
		t.Fatalf("RollbackAfterAddFailure: %v", err)
	}
	got, err := os.ReadFile(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != orig {
		t.Errorf("config should be unchanged, got:\n%s", got)
	}
}

func TestRollbackAfterAddFailure_restoresFromBackup(t *testing.T) {
	dir := t.TempDir()
	cfg := filepath.Join(dir, "config")
	bak := filepath.Join(dir, "config.bak")
	orig := "Host old\n"
	if err := os.WriteFile(bak, []byte(orig), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(cfg, []byte("Host new\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := RollbackAfterAddFailure(cfg, bak, true, true); err != nil {
		t.Fatalf("RollbackAfterAddFailure: %v", err)
	}
	got, err := os.ReadFile(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != orig {
		t.Errorf("config = %q, want restored", got)
	}
}

func TestHostAliasExists_caseInsensitive(t *testing.T) {
	dir := t.TempDir()
	cfg := filepath.Join(dir, "config")
	content := "Host My-VPS\n    HostName 1.2.3.4\n    User root\n    IdentityFile /tmp/k\n"
	if err := os.WriteFile(cfg, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	exists, err := HostAliasExists(cfg, "my-vps")
	if err != nil {
		t.Fatalf("HostAliasExists: %v", err)
	}
	if !exists {
		t.Error("want alias to exist")
	}
	exists, err = HostAliasExists(cfg, "other")
	if err != nil {
		t.Fatalf("HostAliasExists: %v", err)
	}
	if exists {
		t.Error("other should not exist")
	}
}

func TestHostAliasExists_missingFile(t *testing.T) {
	dir := t.TempDir()
	cfg := filepath.Join(dir, "missing")
	exists, err := HostAliasExists(cfg, "x")
	if err != nil {
		t.Fatalf("HostAliasExists: %v", err)
	}
	if exists {
		t.Error("want false for missing config")
	}
}

func TestAppendHost_rejectsDuplicateAlias(t *testing.T) {
	dir := t.TempDir()
	cfg := filepath.Join(dir, "config")
	entry := HostEntry{
		Alias: "dup", HostName: "1.2.3.4", User: "root", IdentityFile: "/tmp/k",
	}
	if err := AppendHost(cfg, entry); err != nil {
		t.Fatalf("first AppendHost: %v", err)
	}
	err := AppendHost(cfg, entry)
	if err == nil {
		t.Fatal("expected duplicate error")
	}
	if !strings.Contains(err.Error(), "已存在") {
		t.Errorf("err = %v", err)
	}
}
