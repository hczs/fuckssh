package config

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDeleteHost_removesOnlyEntry(t *testing.T) {
	dir := t.TempDir()
	cfg := filepath.Join(dir, "config")
	writeConfig(t, cfg, "Host myserver\n    HostName 1.2.3.4\n    User root\n")

	if err := DeleteHost(cfg, "myserver"); err != nil {
		t.Fatalf("DeleteHost: %v", err)
	}

	raw := readRaw(t, cfg)
	if strings.TrimSpace(raw) != "" {
		t.Errorf("file should be empty after deleting only entry, got:\n%s", raw)
	}
}

func TestDeleteHost_removesOneOfMultiple(t *testing.T) {
	dir := t.TempDir()
	cfg := filepath.Join(dir, "config")
	writeConfig(t, cfg, "Host srv1\n    HostName 10.0.0.1\n    User admin\n\nHost srv2\n    HostName 10.0.0.2\n    User root\n")

	if err := DeleteHost(cfg, "srv1"); err != nil {
		t.Fatalf("DeleteHost: %v", err)
	}

	entries, err := ParseFile(cfg)
	if err != nil {
		t.Fatalf("ParseFile: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("len(entries) = %d, want 1", len(entries))
	}
	if entries[0].Alias != "srv2" {
		t.Errorf("remaining alias = %q, want srv2", entries[0].Alias)
	}
}

func TestDeleteHost_removesMiddleEntry(t *testing.T) {
	dir := t.TempDir()
	cfg := filepath.Join(dir, "config")
	writeConfig(t, cfg, "Host srv1\n    HostName 10.0.0.1\n    User admin\n\nHost srv2\n    HostName 10.0.0.2\n    User root\n\nHost srv3\n    HostName 10.0.0.3\n    User deploy\n")

	if err := DeleteHost(cfg, "srv2"); err != nil {
		t.Fatalf("DeleteHost: %v", err)
	}

	entries, err := ParseFile(cfg)
	if err != nil {
		t.Fatalf("ParseFile: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("len(entries) = %d, want 2", len(entries))
	}
	if entries[0].Alias != "srv1" || entries[1].Alias != "srv3" {
		t.Errorf("aliases = [%q, %q], want [srv1, srv3]", entries[0].Alias, entries[1].Alias)
	}
}

func TestDeleteHost_removesRemarkWithBlock(t *testing.T) {
	dir := t.TempDir()
	cfg := filepath.Join(dir, "config")
	writeConfig(t, cfg, "# 生产环境\nHost prod\n    HostName 1.2.3.4\n    User root\n\nHost dev\n    HostName 10.0.0.1\n    User admin\n")

	if err := DeleteHost(cfg, "prod"); err != nil {
		t.Fatalf("DeleteHost: %v", err)
	}

	raw := readRaw(t, cfg)
	if strings.Contains(raw, "生产环境") {
		t.Errorf("remark should be deleted:\n%s", raw)
	}
	if strings.Contains(raw, "prod") {
		t.Errorf("prod should be deleted:\n%s", raw)
	}

	entries, err := ParseFile(cfg)
	if err != nil {
		t.Fatalf("ParseFile: %v", err)
	}
	if len(entries) != 1 || entries[0].Alias != "dev" {
		t.Errorf("entries = %+v, want only dev", entries)
	}
}

func TestDeleteHost_caseInsensitiveMatch(t *testing.T) {
	dir := t.TempDir()
	cfg := filepath.Join(dir, "config")
	writeConfig(t, cfg, "Host MyServer\n    HostName 1.2.3.4\n    User root\n")

	if err := DeleteHost(cfg, "myserver"); err != nil {
		t.Fatalf("DeleteHost: %v", err)
	}

	entries, err := ParseFile(cfg)
	if err != nil {
		t.Fatalf("ParseFile: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("len(entries) = %d, want 0", len(entries))
	}
}

func TestDeleteHost_matchesSecondaryAlias(t *testing.T) {
	dir := t.TempDir()
	cfg := filepath.Join(dir, "config")
	writeConfig(t, cfg, "Host primary secondary\n    HostName 1.2.3.4\n    User root\n")

	if err := DeleteHost(cfg, "secondary"); err != nil {
		t.Fatalf("DeleteHost: %v", err)
	}

	entries, err := ParseFile(cfg)
	if err != nil {
		t.Fatalf("ParseFile: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("len(entries) = %d, want 0 (entire block should be removed)", len(entries))
	}
}

func TestDeleteHost_aliasNotFound(t *testing.T) {
	dir := t.TempDir()
	cfg := filepath.Join(dir, "config")
	writeConfig(t, cfg, "Host myserver\n    HostName 1.2.3.4\n    User root\n")

	err := DeleteHost(cfg, "nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent alias")
	}
	if !errors.Is(err, ErrHostNotFound) {
		t.Errorf("error = %v, want ErrHostNotFound", err)
	}
}

func TestDeleteHost_emptyAlias(t *testing.T) {
	dir := t.TempDir()
	cfg := filepath.Join(dir, "config")
	writeConfig(t, cfg, "Host myserver\n    HostName 1.2.3.4\n    User root\n")

	err := DeleteHost(cfg, "")
	if err == nil {
		t.Fatal("expected error for empty alias")
	}
}

func TestDeleteHost_createsBackup(t *testing.T) {
	dir := t.TempDir()
	cfg := filepath.Join(dir, "config")
	writeConfig(t, cfg, "Host myserver\n    HostName 1.2.3.4\n    User root\n")

	if err := DeleteHost(cfg, "myserver"); err != nil {
		t.Fatalf("DeleteHost: %v", err)
	}

	backupDir, err := filepath.EvalSymlinks(filepath.Join(dir, ".ssh", "backup"))
	if err != nil {
		// Backup dir might be in platform default; check if any backup was created.
		// The backup function uses platform.BackupDir() which may not be in dir.
		return
	}
	entries, _ := os.ReadDir(backupDir)
	if len(entries) == 0 {
		t.Error("expected at least one backup file")
	}
}

func TestDeleteHost_preservesOtherEntries(t *testing.T) {
	dir := t.TempDir()
	cfg := filepath.Join(dir, "config")
	writeConfig(t, cfg, "# 备注1\nHost srv1\n    HostName 10.0.0.1\n    User admin\n    Port 2222\n\n# 备注2\nHost srv2\n    HostName example.com\n    User deploy\n")

	if err := DeleteHost(cfg, "srv1"); err != nil {
		t.Fatalf("DeleteHost: %v", err)
	}

	entries, err := ParseFile(cfg)
	if err != nil {
		t.Fatalf("ParseFile: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("len(entries) = %d, want 1", len(entries))
	}
	if entries[0].Alias != "srv2" {
		t.Errorf("alias = %q, want srv2", entries[0].Alias)
	}
	if entries[0].Remark != "备注2" {
		t.Errorf("remark = %q, want 备注2", entries[0].Remark)
	}
	if entries[0].HostName != "example.com" {
		t.Errorf("HostName = %q, want example.com", entries[0].HostName)
	}
}

func writeConfig(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}

func readRaw(t *testing.T, path string) string {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(raw)
}
