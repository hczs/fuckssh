package keys

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"strings"
	"testing"
)

func TestKeyPaths_sanitizesInvalidChars(t *testing.T) {
	priv, pub := KeyPaths("my/../vps!")

	if strings.Contains(priv, "/") || strings.Contains(priv, "\\") {
		t.Fatalf("priv path contains separator: %q", priv)
	}
	if strings.Contains(priv, "..") {
		t.Fatalf("priv path contains ..: %q", priv)
	}

	wantBase := "id_ed25519_fuckssh_my_vps"
	if priv != wantBase {
		t.Errorf("priv = %q, want %q", priv, wantBase)
	}
	if pub != wantBase+".pub" {
		t.Errorf("pub = %q, want %q", pub, wantBase+".pub")
	}
}

func TestKeyPaths_aliasMyVps(t *testing.T) {
	priv, pub := KeyPaths("my-vps")
	wantBase := "id_ed25519_fuckssh_my-vps"
	if priv != wantBase {
		t.Errorf("priv = %q, want %q", priv, wantBase)
	}
	if pub != wantBase+".pub" {
		t.Errorf("pub = %q, want %q", pub, wantBase+".pub")
	}
}

func TestKeyPaths_defaultWhenAliasEmpty(t *testing.T) {
	host, err := os.Hostname()
	if err != nil || host == "" {
		host = "localhost"
	}
	sum := sha256.Sum256([]byte(host))
	wantSuffix := hex.EncodeToString(sum[:4])

	priv, pub := KeyPaths("")
	wantBase := "id_ed25519_fuckssh_" + wantSuffix
	if priv != wantBase {
		t.Errorf("priv = %q, want %q", priv, wantBase)
	}
	if pub != wantBase+".pub" {
		t.Errorf("pub = %q, want %q", pub, wantBase+".pub")
	}
}
