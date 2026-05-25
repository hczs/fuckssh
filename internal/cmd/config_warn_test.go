package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func Test_maybeWarnInclude_printsWhenPresent(t *testing.T) {
	dir := t.TempDir()
	cfg := filepath.Join(dir, "config")
	if err := os.WriteFile(cfg, []byte("Include extra.conf\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	var stderr bytes.Buffer
	maybeWarnInclude(&stderr, cfg)
	out := stderr.String()
	if !strings.Contains(out, "Include") {
		t.Errorf("stderr = %q, want Include warning", out)
	}
}

func Test_maybeWarnInclude_silentWithoutInclude(t *testing.T) {
	dir := t.TempDir()
	cfg := filepath.Join(dir, "config")
	if err := os.WriteFile(cfg, []byte("Host x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	var stderr bytes.Buffer
	maybeWarnInclude(&stderr, cfg)
	if stderr.Len() != 0 {
		t.Errorf("stderr = %q, want empty", stderr.String())
	}
}
