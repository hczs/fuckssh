package wizard

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/fuckssh/fuckssh/internal/config"
	"github.com/fuckssh/fuckssh/internal/keys"
)

func TestRunKeyFlow_callsDepsInOrder(t *testing.T) {
	dir := t.TempDir()
	cfg := filepath.Join(dir, "config")
	key := filepath.Join(dir, "id_ed25519")
	if err := os.WriteFile(key, []byte("fake-key"), 0o600); err != nil {
		t.Fatal(err)
	}

	var calls []string
	deps := keyFlowDeps{
		backup: func(string) (string, error) {
			calls = append(calls, "backup")
			return cfg + ".bak", nil
		},
		stageKey: func(alias, src string) (string, bool, error) {
			calls = append(calls, "stageKey")
			return src, false, nil
		},
		appendHost: func(string, config.HostEntry) error {
			calls = append(calls, "appendHost")
			return nil
		},
		onProgress: func(int, int, string) {},
	}

	result := &WizardResult{
		Alias:        "test-vps",
		HostName:     "203.0.113.50",
		User:         "ubuntu",
		Port:         "22",
		IdentityFile: key,
	}
	if err := runKeyFlow(cfg, result, deps); err != nil {
		t.Fatalf("runKeyFlow: %v", err)
	}
	want := []string{"backup", "stageKey", "appendHost"}
	if len(calls) != len(want) {
		t.Fatalf("calls = %v, want %v", calls, want)
	}
	for i, w := range want {
		if calls[i] != w {
			t.Fatalf("calls[%d] = %q, want %q; all=%v", i, calls[i], w, calls)
		}
	}
}

func TestRunKeyFlow_integrationCopiesKeyAndAppendsConfig(t *testing.T) {
	dir := t.TempDir()
	if runtime.GOOS == "windows" {
		t.Setenv("USERPROFILE", dir)
	} else {
		t.Setenv("HOME", dir)
	}

	cfg := filepath.Join(dir, "config")
	key := filepath.Join(dir, "id_ed25519")
	if err := os.WriteFile(key, []byte("fake-key"), 0o600); err != nil {
		t.Fatal(err)
	}

	result := &WizardResult{
		Alias:        "test-vps",
		HostName:     "203.0.113.50",
		User:         "ubuntu",
		Port:         "22",
		IdentityFile: key,
	}
	if err := RunKeyFlow(cfg, result); err != nil {
		t.Fatalf("RunKeyFlow: %v", err)
	}

	entries, err := config.ParseFile(cfg)
	if err != nil {
		t.Fatalf("ParseFile: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("len(entries) = %d, want 1", len(entries))
	}
	e := entries[0]
	if e.Alias != "test-vps" || e.HostName != "203.0.113.50" || e.User != "ubuntu" {
		t.Errorf("entry = %+v", e)
	}
	privName, _ := keys.KeyPaths("test-vps")
	wantIdentity := "~/.ssh/keys/" + privName
	if e.IdentityFile != wantIdentity {
		t.Errorf("IdentityFile = %q, want %q", e.IdentityFile, wantIdentity)
	}
	if result.IdentityFile != wantIdentity {
		t.Errorf("result.IdentityFile = %q, want %q", result.IdentityFile, wantIdentity)
	}
	copied := filepath.Join(dir, ".ssh", "keys", privName)
	if _, err := os.Stat(copied); err != nil {
		t.Errorf("copied key missing at %q: %v", copied, err)
	}
}
