package sshclient

import (
	"errors"
	"testing"
)

func TestCheckSSH_found(t *testing.T) {
	restore := stubLookPath(func(name string) (string, error) {
		if name != "ssh" {
			t.Fatalf("LookPath name = %q, want ssh", name)
		}
		return "/usr/bin/ssh", nil
	})
	defer restore()

	path, err := CheckSSH()
	if err != nil {
		t.Fatalf("CheckSSH: %v", err)
	}
	if path != "/usr/bin/ssh" {
		t.Errorf("path = %q, want /usr/bin/ssh", path)
	}
}

func TestCheckSSH_notFound(t *testing.T) {
	restore := stubLookPath(func(string) (string, error) {
		return "", errors.New("executable file not found in %PATH%")
	})
	defer restore()

	_, err := CheckSSH()
	if err == nil {
		t.Fatal("CheckSSH: want error")
	}
	if !errors.Is(err, ErrSSHNotFound) {
		t.Fatalf("errors.Is(err, ErrSSHNotFound) = false, err = %v", err)
	}
}

func stubLookPath(fn func(string) (string, error)) func() {
	prev := lookPath
	lookPath = fn
	return func() { lookPath = prev }
}
