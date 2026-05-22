package keys

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRemoveKeyPair_deletesBothFiles(t *testing.T) {
	dir := t.TempDir()
	priv := filepath.Join(dir, "id_test")
	pub := priv + ".pub"
	if err := os.WriteFile(priv, []byte("p"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(pub, []byte("P"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := RemoveKeyPair(priv); err != nil {
		t.Fatalf("RemoveKeyPair: %v", err)
	}
	for _, p := range []string{priv, pub} {
		if _, err := os.Stat(p); !os.IsNotExist(err) {
			t.Errorf("%s should be removed", p)
		}
	}
}
