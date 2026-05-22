package sshclient

import (
	"context"
	"errors"
	"testing"
)

func TestTestKeyAuth_success(t *testing.T) {
	restoreDial := stubDialKeySSH(func(ctx context.Context, opts KeyAuthOpts) (sshClient, error) {
		return &fakeSSHClient{}, nil
	})
	defer restoreDial()

	err := TestKeyAuth(context.Background(), KeyAuthOpts{
		Host:         "203.0.113.1",
		Port:         "22",
		User:         "root",
		IdentityFile: "/tmp/id_ed25519",
	})
	if err != nil {
		t.Fatalf("TestKeyAuth: %v", err)
	}
}

func TestTestKeyAuth_emptyHost(t *testing.T) {
	err := TestKeyAuth(context.Background(), KeyAuthOpts{User: "root", IdentityFile: "/k"})
	if err == nil || !errors.Is(err, ErrDeployFailed) {
		t.Fatalf("err = %v, want ErrDeployFailed", err)
	}
}

func stubDialKeySSH(fn func(context.Context, KeyAuthOpts) (sshClient, error)) func() {
	prev := dialKeySSH
	dialKeySSH = fn
	return func() { dialKeySSH = prev }
}
