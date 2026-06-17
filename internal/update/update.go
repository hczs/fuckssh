// Package update 提供 fuckssh 自更新：从 GitHub Releases 下载并替换当前二进制。
package update

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const (
	defaultOwner = "hczs"
	defaultRepo  = "fuckssh"
	binaryName   = "fuckssh"
)

// Options 控制一次自更新流程。
type Options struct {
	Owner      string
	Repo       string
	BaseURL    string // 默认 https://github.com，测试时可指向 httptest
	Version    string // 空表示 latest
	CheckOnly  bool
	CurrentVer string
	DestPath   string
	HTTPClient *http.Client
	Out        io.Writer
	ErrOut     io.Writer
}

// Result 描述检查或更新结果。
type Result struct {
	CurrentVersion string
	TargetVersion  string
	Updated        bool
	AlreadyLatest  bool
}

// ErrAlreadyLatest 表示当前已是最新版本（仅 --check 时作为错误返回，便于区分退出语义）。
var ErrAlreadyLatest = errors.New("update: already on latest version")

// Run 检查或执行自更新。
func Run(opts Options) (*Result, error) {
	opts = normalizeOptions(opts)

	target, err := resolveTargetVersion(opts)
	if err != nil {
		return nil, err
	}

	result := &Result{
		CurrentVersion: opts.CurrentVer,
		TargetVersion:  target,
	}

	if isReleaseVersion(opts.CurrentVer) && compareVersion(opts.CurrentVer, target) >= 0 && opts.Version == "" {
		result.AlreadyLatest = true
		if opts.CheckOnly {
			return result, ErrAlreadyLatest
		}
		return result, nil
	}

	if opts.CheckOnly {
		return result, nil
	}

	asset, err := AssetName(runtime.GOOS, runtime.GOARCH)
	if err != nil {
		return nil, err
	}

	data, usedAsset, err := downloadAsset(opts, target, asset)
	if err != nil {
		return nil, err
	}

	binary, err := extractBinary(data, usedAsset)
	if err != nil {
		return nil, err
	}

	if err := installBinary(opts.DestPath, binary); err != nil {
		return nil, err
	}

	result.Updated = true
	return result, nil
}

func normalizeOptions(opts Options) Options {
	if opts.Owner == "" {
		opts.Owner = envOr("FUCKSSH_INSTALL_OWNER", defaultOwner)
	}
	if opts.Repo == "" {
		opts.Repo = envOr("FUCKSSH_INSTALL_REPO", defaultRepo)
	}
	if opts.BaseURL == "" {
		opts.BaseURL = "https://github.com"
	}
	if opts.HTTPClient == nil {
		opts.HTTPClient = &http.Client{Timeout: 2 * time.Minute}
	}
	if opts.Out == nil {
		opts.Out = io.Discard
	}
	if opts.ErrOut == nil {
		opts.ErrOut = io.Discard
	}
	return opts
}

func envOr(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}

func resolveTargetVersion(opts Options) (string, error) {
	if opts.Version != "" {
		return normalizeTag(opts.Version), nil
	}
	return fetchLatestTag(opts.HTTPClient, opts.BaseURL, opts.Owner, opts.Repo)
}

// FetchLatestTag 从 GitHub Releases 获取 latest 标签。
func FetchLatestTag(client *http.Client, owner, repo string) (string, error) {
	return fetchLatestTag(client, "https://github.com", owner, repo)
}

func fetchLatestTag(client *http.Client, baseURL, owner, repo string) (string, error) {
	url := fmt.Sprintf("%s/%s/%s/releases/latest", strings.TrimRight(baseURL, "/"), owner, repo)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("update: fetch latest release: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("update: fetch latest release: HTTP %d", resp.StatusCode)
	}

	var payload struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return "", fmt.Errorf("update: parse latest release: %w", err)
	}
	if payload.TagName == "" {
		return "", errors.New("update: latest release has empty tag_name")
	}
	return payload.TagName, nil
}

// AssetName 返回与 GoReleaser / install.sh 一致的压缩包名。
func AssetName(goos, goarch string) (string, error) {
	arch, err := normalizeArch(goarch)
	if err != nil {
		return "", err
	}
	switch goos {
	case "linux":
		return fmt.Sprintf("%s_linux_%s.tar.gz", binaryName, arch), nil
	case "darwin":
		return fmt.Sprintf("%s_macos_%s.tar.gz", binaryName, arch), nil
	case "windows":
		return fmt.Sprintf("%s_windows_%s.zip", binaryName, arch), nil
	default:
		return "", fmt.Errorf("update: unsupported GOOS %q", goos)
	}
}

// AssetFallback 在分架构 macOS 包不存在时回退到通用包。
func AssetFallback(primary string) (string, bool) {
	switch primary {
	case "fuckssh_macos_arm64.tar.gz", "fuckssh_macos_x86_64.tar.gz":
		return "fuckssh_macos_all.tar.gz", true
	default:
		return "", false
	}
}

func normalizeArch(goarch string) (string, error) {
	switch goarch {
	case "amd64":
		return "x86_64", nil
	case "arm64":
		return "arm64", nil
	default:
		return "", fmt.Errorf("update: unsupported GOARCH %q", goarch)
	}
}

func downloadAsset(opts Options, tag, asset string) ([]byte, string, error) {
	data, err := fetchReleaseAsset(opts.HTTPClient, opts.BaseURL, opts.Owner, opts.Repo, tag, asset)
	if err == nil {
		return data, asset, nil
	}

	fallback, ok := AssetFallback(asset)
	if !ok {
		return nil, "", err
	}
	_, _ = fmt.Fprintf(opts.ErrOut, "warning: %s not found, trying %s\n", asset, fallback)
	data, err = fetchReleaseAsset(opts.HTTPClient, opts.BaseURL, opts.Owner, opts.Repo, tag, fallback)
	if err != nil {
		return nil, "", err
	}
	return data, fallback, nil
}

func fetchReleaseAsset(client *http.Client, baseURL, owner, repo, tag, name string) ([]byte, error) {
	url := fmt.Sprintf("%s/%s/%s/releases/download/%s/%s", strings.TrimRight(baseURL, "/"), owner, repo, tag, name)
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("update: download %s: %w", name, err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("update: download %s: HTTP %d", name, resp.StatusCode)
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("update: read %s: %w", name, err)
	}
	return data, nil
}

func extractBinary(data []byte, assetName string) ([]byte, error) {
	switch {
	case strings.HasSuffix(assetName, ".tar.gz"), strings.HasSuffix(assetName, ".tgz"):
		return extractFromTarGz(data)
	case strings.HasSuffix(assetName, ".zip"):
		return extractFromZip(data)
	default:
		return nil, fmt.Errorf("update: unsupported archive %q", assetName)
	}
}

func extractFromTarGz(data []byte) ([]byte, error) {
	gzr, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("update: open tar.gz: %w", err)
	}
	defer func() { _ = gzr.Close() }()
	return readBinaryFromTar(gzr)
}

func readBinaryFromTar(r io.Reader) ([]byte, error) {
	tr := tar.NewReader(r)
	for {
		hdr, err := tr.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("update: read tar: %w", err)
		}
		if hdr.Typeflag != tar.TypeReg {
			continue
		}
		base := filepath.Base(hdr.Name)
		if base != binaryName && base != binaryName+".exe" {
			continue
		}
		data, err := io.ReadAll(tr)
		if err != nil {
			return nil, fmt.Errorf("update: read binary from tar: %w", err)
		}
		return data, nil
	}
	return nil, fmt.Errorf("update: %s not found in archive", binaryName)
}

func extractFromZip(data []byte) ([]byte, error) {
	readerAt := bytes.NewReader(data)
	zr, err := zip.NewReader(readerAt, int64(len(data)))
	if err != nil {
		return nil, fmt.Errorf("update: open zip: %w", err)
	}
	for _, f := range zr.File {
		base := filepath.Base(f.Name)
		if base != binaryName && base != binaryName+".exe" {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			return nil, fmt.Errorf("update: open zip entry: %w", err)
		}
		content, err := io.ReadAll(rc)
		_ = rc.Close()
		if err != nil {
			return nil, fmt.Errorf("update: read zip entry: %w", err)
		}
		return content, nil
	}
	return nil, fmt.Errorf("update: %s not found in archive", binaryName)
}

func installBinary(dest string, data []byte) error {
	dir := filepath.Dir(dest)
	tmp, err := os.CreateTemp(dir, binaryName+"-update-*")
	if err != nil {
		return fmt.Errorf("update: create temp file: %w", err)
	}
	tmpName := tmp.Name()
	cleanup := func() { _ = os.Remove(tmpName) }

	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		cleanup()
		return fmt.Errorf("update: write temp binary: %w", err)
	}
	if err := tmp.Close(); err != nil {
		cleanup()
		return fmt.Errorf("update: close temp binary: %w", err)
	}
	if runtime.GOOS != "windows" {
		if err := os.Chmod(tmpName, 0o755); err != nil {
			cleanup()
			return fmt.Errorf("update: chmod temp binary: %w", err)
		}
	}

	if runtime.GOOS == "windows" {
		return installBinaryWindows(dest, tmpName, cleanup)
	}
	if err := os.Rename(tmpName, dest); err != nil {
		cleanup()
		return fmt.Errorf("update: replace binary: %w", err)
	}
	return nil
}

func installBinaryWindows(dest, tmpName string, cleanup func()) error {
	oldName := dest + ".old"
	_ = os.Remove(oldName)
	if err := os.Rename(dest, oldName); err != nil {
		cleanup()
		return fmt.Errorf("update: backup current binary: %w", err)
	}
	if err := os.Rename(tmpName, dest); err != nil {
		_ = os.Rename(oldName, dest)
		cleanup()
		return fmt.Errorf("update: replace binary: %w", err)
	}
	cleanup()
	_ = os.Remove(oldName)
	return nil
}

// ResolveExecutable 返回当前可执行文件的绝对路径（解析符号链接）。
func ResolveExecutable() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("update: locate executable: %w", err)
	}
	resolved, err := filepath.EvalSymlinks(exe)
	if err != nil {
		resolved = exe
	}
	abs, err := filepath.Abs(resolved)
	if err != nil {
		return resolved, nil
	}
	return abs, nil
}
