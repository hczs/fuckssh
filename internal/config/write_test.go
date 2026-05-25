package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAppendHost_appendsBlockToFile(t *testing.T) {
	dir := t.TempDir()
	cfg := filepath.Join(dir, "config")
	entry := HostEntry{
		Alias:        "my-vps",
		HostName:     "203.0.113.10",
		User:         "ubuntu",
		Port:         "22",
		IdentityFile: filepath.Join(dir, "id_ed25519"),
	}

	if err := AppendHost(cfg, entry); err != nil {
		t.Fatalf("AppendHost: %v", err)
	}

	entries, err := ParseFile(cfg)
	if err != nil {
		t.Fatalf("ParseFile: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("len(entries) = %d, want 1", len(entries))
	}
	if entries[0].Alias != "my-vps" || entries[0].HostName != "203.0.113.10" {
		t.Errorf("parsed entry = %+v", entries[0])
	}
}

func TestAppendHost_preservesExistingContent(t *testing.T) {
	dir := t.TempDir()
	cfg := filepath.Join(dir, "config")
	existing := "Host existing\n    HostName 10.0.0.1\n    User root\n\n"
	if err := os.WriteFile(cfg, []byte(existing), 0o600); err != nil {
		t.Fatal(err)
	}

	entry := HostEntry{
		Alias:        "newhost",
		HostName:     "10.0.0.2",
		User:         "admin",
		Port:         "2222",
		IdentityFile: filepath.Join(dir, "key"),
	}
	if err := AppendHost(cfg, entry); err != nil {
		t.Fatalf("AppendHost: %v", err)
	}

	raw, err := os.ReadFile(cfg)
	if err != nil {
		t.Fatal(err)
	}
	body := string(raw)
	if !strings.HasPrefix(body, existing) {
		t.Errorf("existing content lost:\n%s", body)
	}

	entries, err := ParseFile(cfg)
	if err != nil {
		t.Fatalf("ParseFile: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("len(entries) = %d, want 2", len(entries))
	}
}

func TestAppendHost_formatsIdentityFileWithQuotesWhenNeeded(t *testing.T) {
	dir := t.TempDir()
	cfg := filepath.Join(dir, "config")
	keyPath := filepath.Join(dir, "my keys", "id_ed25519")
	entry := HostEntry{
		Alias:        "vps",
		HostName:     "1.2.3.4",
		User:         "root",
		Port:         "22",
		IdentityFile: keyPath,
	}

	if err := AppendHost(cfg, entry); err != nil {
		t.Fatalf("AppendHost: %v", err)
	}

	raw, err := os.ReadFile(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(raw), `IdentityFile "`+keyPath+`"`) {
		t.Errorf("config = %q, want quoted IdentityFile", string(raw))
	}
}

func TestAppendHost_writesRemarkComment(t *testing.T) {
	dir := t.TempDir()
	cfg := filepath.Join(dir, "config")
	entry := HostEntry{
		Alias:        "my-vps",
		HostName:     "203.0.113.10",
		User:         "ubuntu",
		Port:         "22",
		IdentityFile: filepath.Join(dir, "id_ed25519"),
		Remark:       "生产环境主站",
	}

	if err := AppendHost(cfg, entry); err != nil {
		t.Fatalf("AppendHost: %v", err)
	}

	raw, err := os.ReadFile(cfg)
	if err != nil {
		t.Fatal(err)
	}
	body := string(raw)
	if !strings.Contains(body, "# 生产环境主站\nHost my-vps") {
		t.Errorf("config should have remark above Host:\n%s", body)
	}

	entries, err := ParseFile(cfg)
	if err != nil {
		t.Fatalf("ParseFile: %v", err)
	}
	if len(entries) != 1 || entries[0].Remark != "生产环境主站" {
		t.Errorf("round-trip Remark = %+v", entries)
	}
}

func TestFormatRemarkComments_emptySkipped(t *testing.T) {
	if got := formatRemarkComments(""); got != "" {
		t.Errorf("formatRemarkComments(\"\") = %q, want empty", got)
	}
	if got := formatRemarkComments("  "); got != "" {
		t.Errorf("formatRemarkComments(\"  \") = %q, want empty", got)
	}
}
