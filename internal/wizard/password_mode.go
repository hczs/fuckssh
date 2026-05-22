package wizard

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fuckssh/fuckssh/internal/config"
	"github.com/fuckssh/fuckssh/internal/i18n"
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
	backup              func(configPath string) (string, error)
	writeKeys           func(sshDir, alias string) (privPath, pubLine string, err error)
	appendHost          func(configPath string, entry config.HostEntry) error
	deploy              func(ctx context.Context, opts sshclient.DeployOpts) error
	promptPermissionFix func(ctx context.Context, perm *sshclient.AuthorizedKeysNotWritableError) error
	onProgress          func(msg string)
}

// RunPasswordMode 通过 huh 逐项收集密码模式字段并执行完整编排。
// configPath 为 ssh config 路径（与 cmd --config 一致）。
func RunPasswordMode(ctx context.Context, configPath string) (*WizardResult, string, error) {
	if strings.TrimSpace(configPath) == "" {
		return nil, "", fmt.Errorf("%w: config 路径不能为空", ErrInvalidInput)
	}

	var draft PasswordModeInput
	var final PasswordModeInput
	for {
		in, err := collectPasswordModeInput(ctx, nil, &draft)
		if err != nil {
			return nil, "", err
		}
		draft = in

		final, err = finalizePasswordModeInput(in)
		clearPassword(&in.Password)
		if err != nil {
			return nil, "", err
		}

		final.Alias, err = ensureAvailableAlias(configPath, final.Alias)
		if err != nil {
			return nil, "", err
		}

		if err := confirmPasswordRun(final, configPath); err != nil {
			if errors.Is(err, ErrWizardRetryForm) {
				// 返回修改：保留全部已填项（含密码）。
				draft = final
				continue
			}
			return nil, "", err
		}
		break
	}

	sshDir, err := platform.SSHDir()
	if err != nil {
		return nil, "", err
	}

	deps := defaultPasswordFlowDeps(sshDir)
	return executePasswordFlow(ctx, final, configPath, deps)
}

func defaultPasswordFlowDeps(sshDir string) passwordFlowDeps {
	const progressTotal = 4
	var step int
	advance := func(msg string) {
		step++
		reportProgressStep(step, progressTotal, msg)
	}
	return passwordFlowDeps{
		backup: config.Backup,
		writeKeys: func(sshDir, alias string) (string, string, error) {
			kp, err := keys.GenerateEd25519()
			if err != nil {
				return "", "", err
			}
			privName, _ := keys.KeyPaths(alias)
			keysDir := filepath.Join(sshDir, "keys")
			if err := keys.WriteKeyPair(keysDir, privName, kp); err != nil {
				return "", "", err
			}
			return filepath.Join(keysDir, privName), kp.PublicLine, nil
		},
		appendHost: config.AppendHost,
		deploy:     sshclient.DeployPublicKey,
		onProgress: advance,
	}
}

type passwordSetupState struct {
	privPath            string
	pubLine             string
	bakPath             string
	configExistedBefore bool
	hostAppended        bool
}

// needsConfigRollback 为 true 时表示本次流程已备份或已写入 config，失败时需恢复。
func (s passwordSetupState) needsConfigRollback() bool {
	return s.bakPath != "" || s.hostAppended
}

// setupPasswordFlow 按顺序：校验别名 → 备份 config → 生成密钥 → 追加 Host（部署在后续步骤）。
// 别名冲突在备份前返回，不触发 config 回滚。
func setupPasswordFlow(ctx context.Context, in PasswordModeInput, configPath string, deps passwordFlowDeps) (passwordSetupState, error) {
	if err := ctx.Err(); err != nil {
		return passwordSetupState{}, err
	}

	exists, err := config.HostAliasExists(configPath, in.Alias)
	if err != nil {
		return passwordSetupState{}, err
	}
	if exists {
		return passwordSetupState{}, fmt.Errorf("%w: %q（请换别名或手动编辑 config）", config.ErrHostExists, in.Alias)
	}

	state := passwordSetupState{configExistedBefore: configFileExists(configPath)}

	progress := deps.onProgress
	if progress == nil {
		progress = func(string) {}
	}

	progress("正在备份 SSH config…")
	bakPath, err := deps.backup(configPath)
	if err != nil {
		return state, err
	}
	state.bakPath = bakPath

	if err := ctx.Err(); err != nil {
		return state, err
	}

	sshDir, err := platform.SSHDir()
	if err != nil {
		return state, err
	}
	if err := ensureSSHLayout(sshDir); err != nil {
		return state, err
	}

	progress("正在生成 Ed25519 密钥…")
	privPath, pubLine, err := deps.writeKeys(sshDir, in.Alias)
	if err != nil {
		return state, err
	}
	state.privPath = privPath
	state.pubLine = pubLine

	if err := ctx.Err(); err != nil {
		return state, err
	}

	progress("正在写入 SSH config…")
	identityRef, err := platform.IdentityFileRef(privPath)
	if err != nil {
		return state, err
	}
	entry := config.HostEntry{
		Alias:        in.Alias,
		HostName:     in.HostName,
		User:         in.User,
		Port:         in.Port,
		IdentityFile: identityRef,
	}
	if err := deps.appendHost(configPath, entry); err != nil {
		return state, err
	}
	state.hostAppended = true

	return state, nil
}

func deployPublicKey(ctx context.Context, in PasswordModeInput, pubLine string, deps passwordFlowDeps) error {
	progress := deps.onProgress
	if progress == nil {
		progress = func(string) {}
	}
	progress("正在连接服务器并部署公钥…")
	return deps.deploy(ctx, sshclient.DeployOpts{
		Host:       in.HostName,
		Port:       in.Port,
		User:       in.User,
		Password:   in.Password,
		PublicLine: pubLine,
	})
}

// rollbackPasswordChanges 在失败时撤销本次流程对本地 config 与密钥的修改。
// 仅当已备份或已追加 Host 时才恢复 config，避免「别名已存在」等前置错误误删原文件。
func rollbackPasswordChanges(configPath string, state passwordSetupState) {
	if state.needsConfigRollback() {
		_ = config.RollbackAfterAddFailure(configPath, state.bakPath, state.configExistedBefore, state.hostAppended)
	}
	if state.privPath != "" {
		_ = keys.RemoveKeyPair(state.privPath)
	}
}

// executePasswordFlow 供单测与 RunPasswordMode 执行完整编排；备份后的任一步失败则回滚本地更改。
func executePasswordFlow(ctx context.Context, in PasswordModeInput, configPath string, deps passwordFlowDeps) (*WizardResult, string, error) {
	setup, err := setupPasswordFlow(ctx, in, configPath, deps)
	if err != nil {
		rollbackPasswordChanges(configPath, setup)
		return nil, setup.bakPath, err
	}
	if err := deployPublicKeyWithRetry(ctx, in, setup.pubLine, deps); err != nil {
		rollbackPasswordChanges(configPath, setup)
		return nil, setup.bakPath, formatPasswordDeployError(err, setup.bakPath, true)
	}
	identityRef, err := platform.IdentityFileRef(setup.privPath)
	if err != nil {
		return nil, setup.bakPath, err
	}
	return &WizardResult{
		Alias:                in.Alias,
		HostName:             in.HostName,
		User:                 in.User,
		Port:                 in.Port,
		IdentityFile:         identityRef,
		BackupPath:           setup.bakPath,
		PasswordFlowComplete: true,
	}, setup.bakPath, nil
}

// formatPasswordDeployError 将底层 deploy 错误转为用户可读中文（保留 %w 供退出码映射）。
// rolledBack 为 true 时表示本地 config/密钥已回滚。
func formatPasswordDeployError(err error, bakPath string, rolledBack bool) error {
	if errors.Is(err, ErrDeployAborted) {
		if rolledBack {
			return fmt.Errorf("%s，已撤销本次对本地 config 与密钥的修改", i18n.T(i18n.KeyWizardPermFixCancelled))
		}
		return fmt.Errorf("%s", i18n.T(i18n.KeyWizardPermFixCancelled))
	}
	if rolledBack {
		if errors.Is(err, sshclient.ErrDeployAuthFailed) {
			return fmt.Errorf("SSH 密码认证失败，已撤销本次对本地 config 与密钥的修改: %w", err)
		}
		return fmt.Errorf("部署公钥失败，已撤销本次对本地 config 与密钥的修改: %w", err)
	}
	if errors.Is(err, sshclient.ErrDeployAuthFailed) {
		if bakPath != "" {
			return fmt.Errorf("SSH 密码认证失败（config 与密钥已写入，备份位于 %s）: %w",
				bakPath, err)
		}
		return fmt.Errorf("SSH 密码认证失败: %w", err)
	}
	if bakPath != "" {
		return fmt.Errorf("部署公钥失败（config 已备份至 %s）: %w", bakPath, err)
	}
	return fmt.Errorf("部署公钥失败: %w", err)
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
	if err := validatePort(in.Port); err != nil {
		return PasswordModeInput{}, err
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

func ensureSSHLayout(dir string) error {
	// 0700：仅用户可访问，符合 OpenSSH 惯例。
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	return os.MkdirAll(filepath.Join(dir, "keys"), 0o700)
}

func configFileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
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
