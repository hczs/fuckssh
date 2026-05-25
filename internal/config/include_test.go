package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestHasIncludeDirective(t *testing.T) {
	dir := t.TempDir()
	withInclude := filepath.Join(dir, "with.conf")
	if err := os.WriteFile(withInclude, []byte("Include ~/.ssh/conf.d/*\nHost x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	has, err := HasIncludeDirective(withInclude)
	if err != nil {
		t.Fatal(err)
	}
	if !has {
		t.Error("want true for Include line")
	}

	plain := filepath.Join(dir, "plain.conf")
	if err := os.WriteFile(plain, []byte("Host y\n    HostName 1.2.3.4\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	has, err = HasIncludeDirective(plain)
	if err != nil {
		t.Fatal(err)
	}
	if has {
		t.Error("want false without Include")
	}

	has, err = HasIncludeDirective(filepath.Join(dir, "missing.conf"))
	if err != nil {
		t.Fatal(err)
	}
	if has {
		t.Error("missing file should return false")
	}
}
