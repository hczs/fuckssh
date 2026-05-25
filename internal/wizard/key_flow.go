package wizard

import (
	"os"
	"path/filepath"

	"github.com/fuckssh/fuckssh/internal/config"
	"github.com/fuckssh/fuckssh/internal/i18n"
	"github.com/fuckssh/fuckssh/internal/keys"
	"github.com/fuckssh/fuckssh/internal/platform"
)

// keyFlowDeps 注入密钥模式写盘步骤，便于单测验证调用顺序。
type keyFlowDeps struct {
	backup     func(configPath string) (string, error)
	stageKey   func(alias, srcPriv string) (destPriv string, copied bool, err error)
	appendHost func(configPath string, entry config.HostEntry) error
	onProgress func(step, total int, msg string)
}

func defaultKeyFlowDeps() keyFlowDeps {
	return keyFlowDeps{
		backup:     config.Backup,
		stageKey:   stageKeyForConfig,
		appendHost: config.AppendHost,
		onProgress: reportProgressStep,
	}
}

// RunKeyFlow 在向导收集完密钥模式输入后，执行备份、落盘私钥与追加 config。
func RunKeyFlow(configPath string, result *WizardResult) error {
	return runKeyFlow(configPath, result, defaultKeyFlowDeps())
}

func runKeyFlow(configPath string, result *WizardResult, deps keyFlowDeps) error {
	if result == nil {
		return ErrInvalidInput
	}

	configExisted := configFileExists(configPath)
	total := KeyFlowProgressTotal

	deps.onProgress(1, total, i18n.T(i18n.KeyWizardProgressBackup))
	bakPath, err := deps.backup(configPath)
	if err != nil {
		return err
	}

	deps.onProgress(2, total, i18n.T(i18n.KeyWizardProgressStageKey))
	destPriv, copied, err := deps.stageKey(result.Alias, result.IdentityFile)
	if err != nil {
		_ = config.RollbackAfterAddFailure(configPath, bakPath, configExisted, false)
		return err
	}

	identityRef, err := platform.IdentityFileRef(destPriv)
	if err != nil {
		if copied {
			_ = keys.RemoveKeyPair(destPriv)
		}
		_ = config.RollbackAfterAddFailure(configPath, bakPath, configExisted, false)
		return err
	}

	deps.onProgress(3, total, i18n.T(i18n.KeyWizardProgressWriteCfg))
	entry := config.HostEntry{
		Alias:        result.Alias,
		HostName:     result.HostName,
		User:         result.User,
		Port:         result.Port,
		IdentityFile: identityRef,
	}
	if err := deps.appendHost(configPath, entry); err != nil {
		_ = config.RollbackAfterAddFailure(configPath, bakPath, configExisted, true)
		if copied {
			_ = keys.RemoveKeyPair(destPriv)
		}
		return err
	}

	result.BackupPath = bakPath
	result.IdentityFile = identityRef
	return nil
}

// stageKeyForConfig 将用户私钥复制到 ~/.ssh/keys/（若尚未在该路径），返回落盘绝对路径与是否新复制。
func stageKeyForConfig(alias, srcPriv string) (destPriv string, copied bool, err error) {
	keysDir, err := platform.KeysDir()
	if err != nil {
		return "", false, err
	}
	if err := os.MkdirAll(keysDir, 0o700); err != nil {
		return "", false, err
	}

	privName, _ := keys.KeyPaths(alias)
	destPriv = filepath.Join(keysDir, privName)
	if filepath.Clean(srcPriv) == filepath.Clean(destPriv) {
		return destPriv, false, nil
	}

	destPriv, err = keys.CopyKeyPair(srcPriv, keysDir, privName)
	if err != nil {
		return "", false, err
	}
	return destPriv, true, nil
}
