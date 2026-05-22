package wizard

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/fuckssh/fuckssh/internal/config"
	"github.com/fuckssh/fuckssh/internal/keys"
	"github.com/fuckssh/fuckssh/internal/platform"
	"github.com/fuckssh/fuckssh/internal/sshclient"
)

// KeyAlgorithm 表示密码模式生成的密钥算法（MVP 仅 Ed25519）。
type KeyAlgorithm string

const (
	AlgorithmEd25519 KeyAlgorithm = "ed25519"
)

// PasswordModeInput 为密码连接模式的用户输入（可在测试中直接构造）。
type PasswordModeInput struct {
	HostName  string
	User      string
	Password  string
	Port      string
	Alias     string
	Algorithm KeyAlgorithm
}

// passwordFlowDeps 注入备份、写密钥、追加 config、部署等步骤，便于单测验证调用顺序。
type passwordFlowDeps struct {
	backup     func(configPath string) (string, error)
	writeKeys  func(sshDir, alias string) (privPath, pubLine string, err error)
	appendHost func(configPath string, entry config.HostEntry) error
	deploy     func(ctx context.Context, opts sshclient.DeployOpts) error
}

// RunPasswordMode 通过 huh 收集密码模式字段并执行完整编排。
// configPath 为 ssh config 路径（与 cmd --config 一致）。
func RunPasswordMode(ctx context.Context, configPath string) (*WizardResult, string, error) {
	var in PasswordModeInput
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("服务器 IP 或域名").
				Value(&in.HostName).
				Validate(nonEmpty("不能为空")),
			huh.NewInput().
				Title("SSH 用户名").
				Value(&in.User).
				Validate(nonEmpty("不能为空")),
			huh.NewInput().
				Title("SSH 密码").
				EchoMode(huh.EchoModePassword).
				Value(&in.Password).
				Validate(nonEmpty("不能为空")),
			huh.NewInput().
				Title("SSH 端口").
				Placeholder("22").
				Value(&in.Port),
			huh.NewInput().
				Title("Host 别名").
				Description("回车则根据 IP/域名自动生成").
				Value(&in.Alias),
		),
	)
	if err := form.Run(); err != nil {
		return nil, "", err
	}
	defer clearPassword(&in.Password)

	final, err := finalizePasswordModeInput(in)
	if err != nil {
		return nil, "", err
	}

	if strings.TrimSpace(configPath) == "" {
		return nil, "", fmt.Errorf("%w: config 路径不能为空", ErrInvalidInput)
	}
	sshDir, err := platform.SSHDir()
	if err != nil {
		return nil, "", err
	}

	return executePasswordFlow(ctx, final, configPath, defaultPasswordFlowDeps(sshDir))
}

func defaultPasswordFlowDeps(sshDir string) passwordFlowDeps {
	return passwordFlowDeps{
		backup: config.Backup,
		writeKeys: func(dir, alias string) (string, string, error) {
			kp, err := keys.GenerateEd25519()
			if err != nil {
				return "", "", err
			}
			privName, _ := keys.KeyPaths(alias)
			if err := keys.WriteKeyPair(dir, privName, kp); err != nil {
				return "", "", err
			}
			return filepath.Join(dir, privName), kp.PublicLine, nil
		},
		appendHost: config.AppendHost,
		deploy:     sshclient.DeployPublicKey,
	}
}

// executePasswordFlow 按顺序：备份 config → 生成密钥 → 追加 Host → 部署公钥。
func executePasswordFlow(ctx context.Context, in PasswordModeInput, configPath string, deps passwordFlowDeps) (*WizardResult, string, error) {
	bakPath, err := deps.backup(configPath)
	if err != nil {
		return nil, "", err
	}

	sshDir, err := platform.SSHDir()
	if err != nil {
		return nil, bakPath, err
	}
	if err := ensureSSHDir(sshDir); err != nil {
		return nil, bakPath, err
	}

	privPath, pubLine, err := deps.writeKeys(sshDir, in.Alias)
	if err != nil {
		return nil, bakPath, err
	}

	entry := config.HostEntry{
		Alias:        in.Alias,
		HostName:     in.HostName,
		User:         in.User,
		Port:         in.Port,
		IdentityFile: privPath,
	}
	if err := deps.appendHost(configPath, entry); err != nil {
		return nil, bakPath, err
	}

	deployOpts := sshclient.DeployOpts{
		Host:       in.HostName,
		Port:       in.Port,
		User:       in.User,
		Password:   in.Password,
		PublicLine: pubLine,
	}
	if err := deps.deploy(ctx, deployOpts); err != nil {
		if bakPath != "" {
			return nil, bakPath, fmt.Errorf("%w（已备份 config 至 %s）", err, bakPath)
		}
		return nil, bakPath, err
	}

	return &WizardResult{
		Alias:        in.Alias,
		HostName:     in.HostName,
		User:         in.User,
		Port:         in.Port,
		IdentityFile: privPath,
	}, bakPath, nil
}

// finalizePasswordModeInput 校验并补全默认值。
func finalizePasswordModeInput(in PasswordModeInput) (PasswordModeInput, error) {
	in.HostName = strings.TrimSpace(in.HostName)
	in.User = strings.TrimSpace(in.User)
	in.Password = strings.TrimSpace(in.Password)
	in.Port = strings.TrimSpace(in.Port)
	in.Alias = strings.TrimSpace(in.Alias)

	if in.HostName == "" || in.User == "" {
		return PasswordModeInput{}, fmt.Errorf("%w: 请填写 IP/域名与用户名", ErrInvalidInput)
	}
	if in.Password == "" {
		return PasswordModeInput{}, fmt.Errorf("%w: 请填写 SSH 密码", ErrInvalidInput)
	}

	if in.Port == "" {
		in.Port = "22"
	}
	if in.Alias == "" {
		in.Alias = keys.SanitizeAlias(in.HostName)
		if in.Alias == "" {
			return PasswordModeInput{}, fmt.Errorf("%w: 无法根据 HostName 生成别名", ErrInvalidInput)
		}
	} else {
		in.Alias = keys.SanitizeAlias(in.Alias)
	}

	if in.Algorithm == "" {
		in.Algorithm = AlgorithmEd25519
	}
	if in.Algorithm != AlgorithmEd25519 {
		return PasswordModeInput{}, fmt.Errorf("%w: 当前仅支持 Ed25519", ErrInvalidInput)
	}

	return in, nil
}

func nonEmpty(msg string) func(string) error {
	return func(s string) error {
		if strings.TrimSpace(s) == "" {
			return errors.New(msg)
		}
		return nil
	}
}

func ensureSSHDir(dir string) error {
	// 0700：仅用户可访问，符合 OpenSSH 惯例。
	return os.MkdirAll(dir, 0o700)
}

// clearPassword 在返回前 best-effort 清零密码字符串（Go 字符串不可变，仅降低残留风险）。
func clearPassword(pw *string) {
	if pw == nil || *pw == "" {
		return
	}
	b := []byte(*pw)
	for i := range b {
		b[i] = 0
	}
	*pw = ""
}
