package update

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestAssetName(t *testing.T) {
	tests := []struct {
		goos, goarch, want string
		wantErr            bool
	}{
		{"linux", "amd64", "fuckssh_linux_x86_64.tar.gz", false},
		{"darwin", "arm64", "fuckssh_macos_arm64.tar.gz", false},
		{"windows", "amd64", "fuckssh_windows_x86_64.zip", false},
		{"plan9", "amd64", "", true},
		{"linux", "386", "", true},
	}
	for _, tt := range tests {
		got, err := AssetName(tt.goos, tt.goarch)
		if tt.wantErr {
			if err == nil {
				t.Errorf("AssetName(%q, %q) expected error", tt.goos, tt.goarch)
			}
			continue
		}
		if err != nil {
			t.Fatalf("AssetName(%q, %q): %v", tt.goos, tt.goarch, err)
		}
		if got != tt.want {
			t.Errorf("AssetName(%q, %q) = %q, want %q", tt.goos, tt.goarch, got, tt.want)
		}
	}
}

func TestAssetFallback(t *testing.T) {
	got, ok := AssetFallback("fuckssh_macos_arm64.tar.gz")
	if !ok || got != "fuckssh_macos_all.tar.gz" {
		t.Fatalf("fallback = %q, ok = %v", got, ok)
	}
	_, ok = AssetFallback("fuckssh_linux_x86_64.tar.gz")
	if ok {
		t.Fatal("linux asset should not have fallback")
	}
}

func TestRun_checkOnlyDetectsUpdate(t *testing.T) {
	srv := newTestReleaseServer(t, "v9.9.9", mustTarGzBinary(t, []byte("#!/bin/sh\n")))
	t.Cleanup(srv.Close)

	result, err := Run(Options{
		Owner:      "test",
		Repo:       "fuckssh",
		BaseURL:    srv.URL,
		CheckOnly:  true,
		CurrentVer: "v0.1.0",
		DestPath:   filepath.Join(t.TempDir(), "fuckssh"),
		HTTPClient: srv.Client(),
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if result.TargetVersion != "v9.9.9" {
		t.Fatalf("target = %q, want v9.9.9", result.TargetVersion)
	}
	if result.Updated {
		t.Fatal("check-only should not update")
	}
}

func TestRun_alreadyLatest(t *testing.T) {
	srv := newTestReleaseServer(t, "v0.6.0", mustTarGzBinary(t, []byte("bin")))
	t.Cleanup(srv.Close)

	result, err := Run(Options{
		Owner:      "test",
		Repo:       "fuckssh",
		BaseURL:    srv.URL,
		CurrentVer: "v0.6.0",
		DestPath:   filepath.Join(t.TempDir(), "fuckssh"),
		HTTPClient: srv.Client(),
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !result.AlreadyLatest || result.Updated {
		t.Fatalf("result = %+v", result)
	}
}

func TestRun_installsBinary_versionWithoutVPrefix(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("unix install path covered separately on windows CI")
	}

	payload := []byte("#!/bin/sh\necho updated\n")
	srv := newTestReleaseServer(t, "v9.9.9", mustTarGzBinary(t, payload))
	t.Cleanup(srv.Close)

	dest := filepath.Join(t.TempDir(), "fuckssh")
	if err := os.WriteFile(dest, []byte("#!/bin/sh\necho old\n"), 0o755); err != nil {
		t.Fatal(err)
	}

	result, err := Run(Options{
		Owner:      "test",
		Repo:       "fuckssh",
		BaseURL:    srv.URL,
		Version:    "9.9.9", // 无 v 前缀，下载 URL 应对齐 GitHub tag v9.9.9
		CurrentVer: "devel",
		DestPath:   dest,
		HTTPClient: srv.Client(),
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if result.TargetVersion != "v9.9.9" {
		t.Fatalf("target = %q, want v9.9.9", result.TargetVersion)
	}
	if !result.Updated {
		t.Fatal("expected Updated=true")
	}

	got, err := os.ReadFile(dest)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, payload) {
		t.Fatalf("binary content = %q, want %q", got, payload)
	}
}

func TestRun_installsBinary(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("unix install path covered separately on windows CI")
	}

	payload := []byte("#!/bin/sh\necho updated\n")
	srv := newTestReleaseServer(t, "v9.9.9", mustTarGzBinary(t, payload))
	t.Cleanup(srv.Close)

	dest := filepath.Join(t.TempDir(), "fuckssh")
	if err := os.WriteFile(dest, []byte("#!/bin/sh\necho old\n"), 0o755); err != nil {
		t.Fatal(err)
	}

	result, err := Run(Options{
		Owner:      "test",
		Repo:       "fuckssh",
		BaseURL:    srv.URL,
		CurrentVer: "devel",
		DestPath:   dest,
		HTTPClient: srv.Client(),
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !result.Updated {
		t.Fatal("expected Updated=true")
	}

	got, err := os.ReadFile(dest)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, payload) {
		t.Fatalf("binary content = %q, want %q", got, payload)
	}
}

func newTestReleaseServer(t *testing.T, tag string, archive []byte) *httptest.Server {
	t.Helper()
	asset, err := AssetName(runtime.GOOS, runtime.GOARCH)
	if err != nil {
		t.Fatal(err)
	}
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/test/fuckssh/releases/latest":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"tag_name":"` + tag + `"}`))
		case "/test/fuckssh/releases/download/" + tag + "/" + asset:
			_, _ = w.Write(archive)
		default:
			http.NotFound(w, r)
		}
	}))
}

func mustTarGzBinary(t *testing.T, content []byte) []byte {
	t.Helper()
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gw)
	name := binaryName
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	if err := tw.WriteHeader(&tar.Header{
		Name: name,
		Mode: 0o755,
		Size: int64(len(content)),
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := tw.Write(content); err != nil {
		t.Fatal(err)
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gw.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}
