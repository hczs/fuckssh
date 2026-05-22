package platform

import (
	"path/filepath"
	"runtime"
	"testing"
)

func TestSSHDir_windows_usesUserProfile(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("windows only")
	}

	t.Setenv("USERPROFILE", `C:\Users\testuser`)
	t.Setenv("HOME", `C:\should\not\use`)

	got, err := SSHDir()
	if err != nil {
		t.Fatalf("SSHDir: %v", err)
	}

	want := filepath.Join(`C:\Users\testuser`, ".ssh")
	if got != want {
		t.Errorf("SSHDir() = %q, want %q", got, want)
	}
}

func TestSSHDir_unix_usesHome(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("unix only")
	}

	t.Setenv("HOME", "/home/testuser")
	t.Setenv("USERPROFILE", "/should/not/use")

	got, err := SSHDir()
	if err != nil {
		t.Fatalf("SSHDir: %v", err)
	}

	want := filepath.Join("/home/testuser", ".ssh")
	if got != want {
		t.Errorf("SSHDir() = %q, want %q", got, want)
	}
}

func TestDefaultConfigPath_joinsConfig(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Setenv("USERPROFILE", `C:\Users\testuser`)
	} else {
		t.Setenv("HOME", "/home/testuser")
	}

	got, err := DefaultConfigPath()
	if err != nil {
		t.Fatalf("DefaultConfigPath: %v", err)
	}

	sshDir, err := SSHDir()
	if err != nil {
		t.Fatalf("SSHDir: %v", err)
	}

	want := filepath.Join(sshDir, "config")
	if got != want {
		t.Errorf("DefaultConfigPath() = %q, want %q", got, want)
	}
}

func TestExpandPath_tildePrefix(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Setenv("USERPROFILE", `C:\Users\testuser`)
	} else {
		t.Setenv("HOME", "/home/testuser")
	}

	got, err := ExpandPath("~/keys/id_ed25519")
	if err != nil {
		t.Fatalf("ExpandPath: %v", err)
	}

	home, err := userHomeDir()
	if err != nil {
		t.Fatalf("userHomeDir: %v", err)
	}

	want := filepath.Join(home, "keys", "id_ed25519")
	if got != want {
		t.Errorf("ExpandPath() = %q, want %q", got, want)
	}
}
