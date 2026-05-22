package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestPruneBackups_keepsNewest20(t *testing.T) {
	dir := t.TempDir()
	base := time.Now()

	for i := 0; i < 25; i++ {
		name := filepath.Join(dir, "config.fuckssh.bak."+base.Add(time.Duration(i)*time.Second).UTC().Format("20060102T150405Z"))
		if err := os.WriteFile(name, []byte("x"), 0o600); err != nil {
			t.Fatal(err)
		}
		// 让修改时间递增，便于排序
		ts := base.Add(time.Duration(i) * time.Second)
		if err := os.Chtimes(name, ts, ts); err != nil {
			t.Fatal(err)
		}
	}

	if err := PruneBackups(dir, 20); err != nil {
		t.Fatalf("PruneBackups: %v", err)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 20 {
		t.Fatalf("len(backups) = %d, want 20", len(entries))
	}
}
