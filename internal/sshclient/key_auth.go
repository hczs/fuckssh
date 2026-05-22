package sshclient

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/fuckssh/fuckssh/internal/keys"
	"golang.org/x/crypto/ssh"
)

// KeyAuthOpts 为私钥连接测试所需参数。
type KeyAuthOpts struct {
	Host         string
	Port         string
	User         string
	IdentityFile string
}

// loadSignerFn 可在测试中注入。
var loadSignerFn = keys.LoadSignerFromFile

// dialKeySSH 可在测试中替换。
var dialKeySSH = defaultDialKeySSH

// TestKeyAuth 验证能否用指定私钥建立 SSH 连接（向导「测试连接」用）。
func TestKeyAuth(ctx context.Context, opts KeyAuthOpts) error {
	if strings.TrimSpace(opts.Host) == "" || strings.TrimSpace(opts.User) == "" {
		return fmt.Errorf("%w: Host 与 User 不能为空", ErrDeployFailed)
	}
	if strings.TrimSpace(opts.IdentityFile) == "" {
		return fmt.Errorf("%w: 私钥路径不能为空", ErrDeployFailed)
	}
	client, err := dialKeySSH(ctx, opts)
	if err != nil {
		return classifyDialError(err)
	}
	return client.Close()
}

func defaultDialKeySSH(ctx context.Context, opts KeyAuthOpts) (sshClient, error) {
	signer, err := loadSignerFn(strings.TrimSpace(opts.IdentityFile))
	if err != nil {
		if errors.Is(err, keys.ErrPassphraseNotSupported) {
			return nil, fmt.Errorf("%w: %w", ErrDeployFailed, err)
		}
		return nil, fmt.Errorf("%w: %w", ErrDeployFailed, err)
	}

	port := opts.Port
	if port == "" {
		port = "22"
	}
	portNum, err := strconv.Atoi(port)
	if err != nil {
		return nil, fmt.Errorf("invalid port %q: %w", port, err)
	}

	config := &ssh.ClientConfig{
		User: strings.TrimSpace(opts.User),
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), //nolint:gosec // MVP
		Timeout:         30 * time.Second,
	}

	addr := net.JoinHostPort(strings.TrimSpace(opts.Host), strconv.Itoa(portNum))
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
