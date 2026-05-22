package wizard

import (
	"errors"
	"fmt"
	"strings"

	"github.com/fuckssh/fuckssh/internal/config"
	"github.com/fuckssh/fuckssh/internal/i18n"
	"github.com/fuckssh/fuckssh/internal/keys"
)

// resolveAliasCandidate 将用户输入或 HostName 推导为待写入的 Host 别名。
func resolveAliasCandidate(raw, hostName string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return keys.SanitizeAlias(hostName)
	}
	return keys.SanitizeAlias(raw)
}

// aliasFieldValidate 在别名步骤内联校验冲突（避免确认后再弹第二层表单）。
func aliasFieldValidate(configPath string, hostName *string) func(string) error {
	return func(raw string) error {
		host := strings.TrimSpace(*hostName)
		if host == "" {
			return errors.New(i18n.T(i18n.KeyWizardErrEmpty))
		}
		alias := resolveAliasCandidate(raw, host)
		if alias == "" {
			return errors.New(i18n.T(i18n.KeyWizardErrAliasGen))
		}
		exists, err := config.HostAliasExists(configPath, alias)
		if err != nil {
			return err
		}
		if exists {
			return fmt.Errorf("%s", i18n.T(i18n.KeyWizardAliasConflictNote, alias))
		}
		return nil
	}
}
