package platform

import (
	"path/filepath"
	"runtime"
	"testing"
)

func TestBackupDir_and_KeysDir_underSSH(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Setenv("USERPROFILE", `C:\Users\testuser`)
	} else {
		t.Setenv("HOME", "/home/testuser")
	}

	sshDir, err := SSHDir()
	if err != nil {
		t.Fatalf("SSHDir: %v", err)
	}

	backupDir, err := BackupDir()
	if err != nil {
		t.Fatalf("BackupDir: %v", err)
	}
	if backupDir != filepath.Join(sshDir, "backup") {
		t.Errorf("BackupDir() = %q, want %q", backupDir, filepath.Join(sshDir, "backup"))
	}

	keysDir, err := KeysDir()
	if err != nil {
		t.Fatalf("KeysDir: %v", err)
	}
	if keysDir != filepath.Join(sshDir, "keys") {
		t.Errorf("KeysDir() = %q, want %q", keysDir, filepath.Join(sshDir, "keys"))
	}
}

func TestIdentityFileRef_underSSHDir(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Setenv("USERPROFILE", `C:\Users\testuser`)
	} else {
		t.Setenv("HOME", "/home/testuser")
	}

	sshDir, err := SSHDir()
	if err != nil {
		t.Fatalf("SSHDir: %v", err)
	}

	abs := filepath.Join(sshDir, "keys", "id_ed25519_fuckssh_my")
	got, err := IdentityFileRef(abs)
	if err != nil {
		t.Fatalf("IdentityFileRef: %v", err)
	}
	want := "~/.ssh/keys/id_ed25519_fuckssh_my"
	if got != want {
		t.Errorf("IdentityFileRef() = %q, want %q", got, want)
	}
}

func TestIdentityFileRef_outsideSSHDir_fallsBackAbsolute(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Setenv("USERPROFILE", `C:\Users\testuser`)
	} else {
		t.Setenv("HOME", "/home/testuser")
	}

	abs := filepath.Join(t.TempDir(), "custom", "id_rsa")
	got, err := IdentityFileRef(abs)
	if err != nil {
		t.Fatalf("IdentityFileRef: %v", err)
	}
	want := filepath.ToSlash(abs)
	if got != want {
		t.Errorf("IdentityFileRef() = %q, want %q", got, want)
	}
}
