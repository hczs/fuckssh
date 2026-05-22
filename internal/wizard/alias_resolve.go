package wizard

import (
	"errors"
	"fmt"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/fuckssh/fuckssh/internal/config"
	"github.com/fuckssh/fuckssh/internal/i18n"
	"github.com/fuckssh/fuckssh/internal/keys"
)

// ensureAvailableAlias 若别名已存在则交互式要求用户输入新别名，直到可用或取消。
func ensureAvailableAlias(configPath, alias string) (string, error) {
	alias = keys.SanitizeAlias(strings.TrimSpace(alias))
	if alias == "" {
		return "", fmt.Errorf("%w: %s", ErrInvalidInput, i18n.T(i18n.KeyWizardErrAliasGen))
	}
	for {
		exists, err := config.HostAliasExists(configPath, alias)
		if err != nil {
			return "", err
		}
		if !exists {
			return alias, nil
		}
		newAlias, err := promptNewAlias(configPath, alias)
		if err != nil {
			return "", err
		}
		alias = keys.SanitizeAlias(newAlias)
		if alias == "" {
			return "", fmt.Errorf("%w: %s", ErrInvalidInput, i18n.T(i18n.KeyWizardErrAliasGen))
		}
	}
}

func promptNewAlias(configPath, conflict string) (string, error) {
	var newAlias string
	emptyMsg := i18n.T(i18n.KeyWizardErrEmpty)
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewNote().
				Title(stepTitle(6, i18n.KeyWizardAlias)).
				Description(i18n.T(i18n.KeyWizardAliasConflictNote, conflict)),
			huh.NewInput().
				Title(i18n.T(i18n.KeyWizardAliasNew)).
				Value(&newAlias).
				Validate(func(s string) error {
					if strings.TrimSpace(s) == "" {
						return errors.New(emptyMsg)
					}
					candidate := keys.SanitizeAlias(s)
					exists, err := config.HostAliasExists(configPath, candidate)
					if err != nil {
						return err
					}
					if exists {
						return errors.New(i18n.T(i18n.KeyWizardAliasStillExists))
					}
					return nil
				}),
		),
	)
	if err := form.Run(); err != nil {
		return "", err
	}
	return strings.TrimSpace(newAlias), nil
}
