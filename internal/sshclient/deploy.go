package sshclient

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

// ErrDeployFailed 表示远端公钥部署失败（CLI 映射退出码 4）。
var ErrDeployFailed = errors.New("sshclient: deploy failed")

// ErrDeployAuthFailed 表示 SSH 密码认证失败。
var ErrDeployAuthFailed = errors.New("sshclient: authentication failed")

// DeployOpts 为密码登录并部署公钥所需的连接参数。
type DeployOpts struct {
	Host       string
	Port       string
	User       string
	Password   string
	PublicLine string // authorized_keys 单行（可含末尾换行）
}

// sshClient 抽象真实 SSH 客户端，便于单测注入。
type sshClient interface {
	RunSession(cmd string) (stdout, stderr string, exitCode int, err error)
	// WriteAuthorizedKeys 通过会话 stdin 写入远端 ~/.ssh/authorized_keys（不依赖远端 base64）。
	WriteAuthorizedKeys(content []byte) error
	Close() error
}

// dialSSH 可在测试中替换；默认建立密码认证的 SSH 连接。
var dialSSH = defaultDialSSH

const deployMaxAttempts = 3 // 首次 + 最多 2 次重试

// TestPasswordAuth 仅验证能否用密码建立 SSH 连接（向导填完密码后「测试连接」用）。
func TestPasswordAuth(ctx context.Context, opts DeployOpts) error {
	if strings.TrimSpace(opts.Host) == "" || strings.TrimSpace(opts.User) == "" {
		return fmt.Errorf("%w: Host 与 User 不能为空", ErrDeployFailed)
	}
	if strings.TrimSpace(opts.Password) == "" {
		return fmt.Errorf("%w: 密码不能为空", ErrDeployFailed)
	}
	client, err := dialSSH(ctx, opts)
	if err != nil {
		return classifyDialError(err)
	}
	return client.Close()
}

// DeployPublicKey 使用密码登录远端，确保 ~/.ssh 存在并将公钥追加到 authorized_keys。
//
// MVP 使用 InsecureIgnoreHostKey，生产环境应加强 host key 校验（见架构待办）。
func DeployPublicKey(ctx context.Context, opts DeployOpts) error {
	if strings.TrimSpace(opts.Host) == "" || strings.TrimSpace(opts.User) == "" {
		return fmt.Errorf("%w: Host 与 User 不能为空", ErrDeployFailed)
	}
	if strings.TrimSpace(opts.Password) == "" {
		return fmt.Errorf("%w: 密码不能为空", ErrDeployFailed)
	}
	if strings.TrimSpace(opts.PublicLine) == "" {
		return fmt.Errorf("%w: 公钥不能为空", ErrDeployFailed)
	}

	var lastErr error
	for attempt := 0; attempt < deployMaxAttempts; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(time.Duration(attempt) * 500 * time.Millisecond):
			}
		}
		lastErr = deployOnce(ctx, opts)
		if lastErr == nil {
			return nil
		}
		if errors.Is(lastErr, ErrDeployAuthFailed) || !isRetryableDeployError(lastErr) {
			return lastErr
		}
	}
	return lastErr
}

func deployOnce(ctx context.Context, opts DeployOpts) error {
	client, err := dialSSH(ctx, opts)
	if err != nil {
		return classifyDialError(err)
	}
	defer client.Close()

	if err := runRemote(client, "mkdir -p ~/.ssh && chmod 700 ~/.ssh"); err != nil {
		return fmt.Errorf("%w: 创建 ~/.ssh: %v", ErrDeployFailed, err)
	}

	existing, err := runRemoteOutput(client, "cat ~/.ssh/authorized_keys 2>/dev/null || true")
	if err != nil {
		return fmt.Errorf("%w: 读取 authorized_keys: %v", ErrDeployFailed, err)
	}
	newContent := appendAuthorizedKey(existing, opts.PublicLine)
	if err := client.WriteAuthorizedKeys([]byte(newContent)); err != nil {
		return fmt.Errorf("%w: 写入 authorized_keys: %v", ErrDeployFailed, err)
	}
	return nil
}

// appendAuthorizedKey 将公钥行追加到既有 authorized_keys 内容（不重复追加同一行）。
func appendAuthorizedKey(existing, pubLine string) string {
	pubLine = strings.TrimSpace(pubLine)
	if pubLine == "" {
		return existing
	}
	for _, line := range strings.Split(existing, "\n") {
		if strings.TrimSpace(line) == pubLine {
			return existing
		}
	}
	var b strings.Builder
	b.WriteString(existing)
	if existing != "" && !strings.HasSuffix(existing, "\n") {
		b.WriteString("\n")
	}
	b.WriteString(pubLine)
	b.WriteString("\n")
	return b.String()
}

func runRemote(client sshClient, cmd string) error {
	_, _, _, err := client.RunSession(cmd)
	return err
}

func runRemoteOutput(client sshClient, cmd string) (string, error) {
	stdout, _, _, err := client.RunSession(cmd)
	return stdout, err
}

func classifyDialError(err error) error {
	if isAuthError(err) {
		return fmt.Errorf("%w: %w", ErrDeployFailed, ErrDeployAuthFailed)
	}
	if isRetryableDeployError(err) {
		return fmt.Errorf("%w: %v", ErrDeployFailed, err)
	}
	return fmt.Errorf("%w: %v", ErrDeployFailed, err)
}

func isAuthError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "unable to authenticate") ||
		strings.Contains(msg, "authentication failed") ||
		strings.Contains(msg, "no supported methods remain")
}

func isRetryableDeployError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, ErrDeployAuthFailed) {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "connection refused") ||
		strings.Contains(msg, "connection reset") ||
		strings.Contains(msg, "timeout") ||
		strings.Contains(msg, "i/o timeout") ||
		strings.Contains(msg, "temporary failure")
}

func defaultDialSSH(ctx context.Context, opts DeployOpts) (sshClient, error) {
	port := opts.Port
	if port == "" {
		port = "22"
	}
	portNum, err := strconv.Atoi(port)
	if err != nil {
		return nil, fmt.Errorf("invalid port %q: %w", port, err)
	}

	config := &ssh.ClientConfig{
		User: opts.User,
		Auth: []ssh.AuthMethod{
			ssh.Password(opts.Password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), //nolint:gosec // MVP：待加强 host key 校验
		Timeout:         30 * time.Second,
	}

	addr := net.JoinHostPort(opts.Host, strconv.Itoa(portNum))
	var d net.Dialer
	conn, err := d.DialContext(ctx, "tcp", addr)
	if err != nil {
		return nil, err
	}

	c, chans, reqs, err := ssh.NewClientConn(conn, addr, config)
	if err != nil {
		_ = conn.Close()
		return nil, err
	}
	return &realSSHClient{client: ssh.NewClient(c, chans, reqs)}, nil
}

type realSSHClient struct {
	client *ssh.Client
}

func (r *realSSHClient) WriteAuthorizedKeys(content []byte) error {
	sess, err := r.client.NewSession()
	if err != nil {
		return err
	}
	defer sess.Close()

	sess.Stdin = bytes.NewReader(content)
	// 通过 stdin 喂给 cat；使用 $HOME 避免单引号内 ~ 不展开，且不依赖远端 base64。
	cmd := `sh -c "mkdir -p \"$HOME/.ssh\" && chmod 700 \"$HOME/.ssh\" && cat > \"$HOME/.ssh/authorized_keys\" && chmod 600 \"$HOME/.ssh/authorized_keys\""`
	runErr := sess.Run(cmd)
	if runErr != nil {
		return fmt.Errorf("write authorized_keys: %w", runErr)
	}
	return nil
}

func (r *realSSHClient) RunSession(cmd string) (stdout, stderr string, exitCode int, err error) {
	sess, err := r.client.NewSession()
	if err != nil {
		return "", "", -1, err
	}
	defer sess.Close()

	var outBuf, errBuf strings.Builder
	sess.Stdout = &outBuf
	sess.Stderr = &errBuf
	runErr := sess.Run(cmd)
	exitCode = 0
	if runErr != nil {
		var exitErr *ssh.ExitError
		if errors.As(runErr, &exitErr) {
			exitCode = exitErr.ExitStatus()
		} else {
			return outBuf.String(), errBuf.String(), -1, runErr
		}
	}
	if exitCode != 0 {
		return outBuf.String(), errBuf.String(), exitCode,
			fmt.Errorf("remote command failed (exit %d): %s", exitCode, strings.TrimSpace(errBuf.String()))
	}
	return outBuf.String(), errBuf.String(), 0, nil
}

func (r *realSSHClient) Close() error {
	return r.client.Close()
}
