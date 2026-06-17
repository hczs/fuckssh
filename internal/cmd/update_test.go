package cmd

import (
	"bytes"
	"errors"
	"io"
	"testing"
)

func TestRunUpdate_alreadyLatestMessage(t *testing.T) {
	prev := runUpdateFn
	runUpdateFn = func(stdout, stderr io.Writer) error {
		_, _ = stdout.Write([]byte("already latest"))
		return nil
	}
	t.Cleanup(func() { runUpdateFn = prev })

	var stdout, stderr bytes.Buffer
	rootCmd.SetOut(&stdout)
	rootCmd.SetErr(&stderr)

	if err := ExecuteWithArgs([]string{"update", "--check"}); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if !bytes.Contains(stdout.Bytes(), []byte("already latest")) {
		t.Fatalf("stdout = %q", stdout.String())
	}
}

func TestRunUpdate_checkAvailable(t *testing.T) {
	prev := runUpdateFn
	runUpdateFn = func(stdout, _ io.Writer) error {
		_, err := stdout.Write([]byte("update available"))
		return err
	}
	t.Cleanup(func() { runUpdateFn = prev })

	var stdout, stderr bytes.Buffer
	rootCmd.SetOut(&stdout)
	rootCmd.SetErr(&stderr)

	if err := ExecuteWithArgs([]string{"update", "-c"}); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if !bytes.Contains(stdout.Bytes(), []byte("update available")) {
		t.Fatalf("stdout = %q", stdout.String())
	}
}

func TestRunUpdate_propagatesError(t *testing.T) {
	want := errors.New("network down")
	prev := runUpdateFn
	runUpdateFn = func(_, _ io.Writer) error { return want }
	t.Cleanup(func() { runUpdateFn = prev })

	var stdout, stderr bytes.Buffer
	rootCmd.SetOut(&stdout)
	rootCmd.SetErr(&stderr)

	err := ExecuteWithArgs([]string{"update"})
	if !errors.Is(err, want) {
		t.Fatalf("err = %v, want %v", err, want)
	}
}
