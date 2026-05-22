package wizard

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestKeyMode_validateRequiresHostUserIdentity(t *testing.T) {
	_, err := finalizeKeyModeInput(KeyModeInput{}, os.Stat)
	if err == nil {
		t.Fatal("expected error for empty input")
	}

	cases := []KeyModeInput{
		{User: "root", IdentityFile: "/tmp/key"},
		{HostName: "1.2.3.4", IdentityFile: "/tmp/key"},
		{HostName: "1.2.3.4", User: "root"},
	}
	for _, in := range cases {
		if _, err := finalizeKeyModeInput(in, os.Stat); err == nil {
			t.Errorf("finalizeKeyModeInput(%+v): want error", in)
		}
	}
}

func TestKeyMode_defaultPort22(t *testing.T) {
	key := filepath.Join(t.TempDir(), "id_ed25519")
	if err := os.WriteFile(key, []byte("fake"), 0o600); err != nil {
		t.Fatal(err)
	}

	out, err := finalizeKeyModeInput(KeyModeInput{
		HostName:     "203.0.113.1",
		User:         "ubuntu",
		IdentityFile: key,
	}, os.Stat)
	if err != nil {
		t.Fatalf("finalizeKeyModeInput: %v", err)
	}
	if out.Port != "22" {
		t.Errorf("Port = %q, want 22", out.Port)
	}
}

func TestKeyMode_generatesAliasWhenEmpty(t *testing.T) {
	key := filepath.Join(t.TempDir(), "id_ed25519")
	if err := os.WriteFile(key, []byte("fake"), 0o600); err != nil {
		t.Fatal(err)
	}

	out, err := finalizeKeyModeInput(KeyModeInput{
		HostName:     "203.0.113.10",
		User:         "root",
		IdentityFile: key,
	}, os.Stat)
	if err != nil {
		t.Fatalf("finalizeKeyModeInput: %v", err)
	}
	if out.Alias == "" {
		t.Fatal("Alias should be generated from HostName")
	}
	if out.Alias != "203_0_113_10" {
		t.Errorf("Alias = %q, want 203_0_113_10", out.Alias)
	}
}

func TestKeyMode_rejectsMissingPrivateKey(t *testing.T) {
	_, err := finalizeKeyModeInput(KeyModeInput{
		HostName:     "1.2.3.4",
		User:         "root",
		Alias:        "vps",
		IdentityFile: filepath.Join(t.TempDir(), "missing"),
	}, os.Stat)
	if err == nil {
		t.Fatal("expected error for missing key")
	}
	if !errors.Is(err, ErrInvalidInput) {
		t.Errorf("error = %v, want ErrInvalidInput", err)
	}
}
